package nft

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum"
)

// AgentMetadata represents the metadata for an agent NFT
type AgentMetadata struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Image        string                 `json:"image"`
	AgentID      string                 `json:"agent_id"`
	Capabilities []string               `json:"capabilities"`
	Properties   map[string]interface{} `json:"properties,omitempty"`
}

// IPFSUploadResponse represents the response from IPFS upload
type IPFSUploadResponse struct {
	Success bool   `json:"success"`
	IpfsHash string `json:"ipfsHash"`
	PinSize  int64  `json:"pinSize"`
	Error    string `json:"error,omitempty"`
}

// MintSignatureRequest represents the request to get a mint signature
type MintSignatureRequest struct {
	To       string `json:"to"`
	TokenURI string `json:"tokenURI"`
	Nonce    uint64 `json:"nonce"`
}

// MintSignatureResponse represents the response with mint signature
type MintSignatureResponse struct {
	Signature string `json:"signature"`
	Nonce     uint64 `json:"nonce"`
}

// ContractConfigResponse represents the contract configuration
type ContractConfigResponse struct {
	ContractAddress string `json:"contract_address"`
	ChainID         string `json:"chain_id"`
	NetworkName     string `json:"network_name"`
}

// NFTMinter handles NFT minting operations
type NFTMinter struct {
	client          *ethclient.Client
	contractAddress common.Address
	backendURL      string
	chainID         *big.Int
	privateKey      *ecdsa.PrivateKey
	address         common.Address
	httpClient      *http.Client
}

// NewNFTMinter creates a new NFT minter instance
func NewNFTMinter(backendURL, rpcEndpoint, privateKeyHex string) (*NFTMinter, error) {
	// Parse private key
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	// Get address from private key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create Ethereum client if RPC endpoint provided
	var ethClient *ethclient.Client
	if rpcEndpoint != "" {
		ethClient, err = ethclient.Dial(rpcEndpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
		}
	}

	return &NFTMinter{
		client:     ethClient,
		backendURL: backendURL,
		privateKey: privateKey,
		address:    address,
		httpClient: httpClient,
	}, nil
}

// MintAgent mints a new agent NFT
func (m *NFTMinter) MintAgent(metadata AgentMetadata) (uint64, error) {
	fmt.Println("   [Step 1/5] 🔍 Getting contract configuration...")
	// 1. Get contract configuration from backend
	config, err := m.getContractConfig()
	if err != nil {
		return 0, fmt.Errorf("failed to get contract config: %w", err)
	}

	// Set contract address
	m.contractAddress = common.HexToAddress(config.ContractAddress)
	fmt.Printf("   ✅ Contract address: %s\n", config.ContractAddress)
	
	// Set chain ID
	chainID, ok := new(big.Int).SetString(config.ChainID, 10)
	if !ok {
		return 0, fmt.Errorf("invalid chain ID: %s", config.ChainID)
	}
	m.chainID = chainID
	fmt.Printf("   ✅ Chain ID: %s\n", config.ChainID)

	fmt.Println("\n   [Step 2/5] 📤 Uploading metadata to IPFS...")
	// 2. Send metadata to backend (backend handles IPFS upload via Pinata)
	ipfsHash, err := m.uploadMetadataToIPFS(metadata)
	if err != nil {
		return 0, fmt.Errorf("failed to send metadata to backend: %w", err)
	}
	fmt.Printf("   ✅ IPFS URI: %s\n", ipfsHash)

	fmt.Println("\n   [Step 3/5] 🔢 Getting nonce from contract...")
	// 3. Get current nonce from contract for this wallet
	nonce, err := m.getNonce(m.address)
	if err != nil {
		return 0, fmt.Errorf("failed to get nonce: %w", err)
	}

	fmt.Printf("   ✅ Nonce: %d\n", nonce)

	fmt.Println("\n   [Step 4/5] 🔐 Requesting mint signature...")
	// 4. Request mint signature from backend (passing wallet address + IPFS URI + nonce)
	signature, err := m.requestMintSignature(m.address.Hex(), ipfsHash, nonce)
	if err != nil {
		return 0, fmt.Errorf("failed to get mint signature: %w", err)
	}

	fmt.Println("\n   [Step 5/5] ⛓️  Executing blockchain transaction...")
	// 5. Execute mint transaction on-chain with the signature
	tokenID, err := m.executeMint(signature)
	if err != nil {
		return 0, fmt.Errorf("failed to execute mint: %w", err)
	}

	return tokenID, nil
}

// uploadMetadataToIPFS sends agent metadata to backend which handles IPFS upload
func (m *NFTMinter) uploadMetadataToIPFS(metadata AgentMetadata) (string, error) {
	// The backend handles the actual IPFS upload via Pinata
	// We just send the metadata to the backend endpoint
	
	// Prepare request body with agent metadata
	body, err := json.Marshal(metadata)
	if err != nil {
		return "", fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Create request to backend
	// Ensure backend URL doesn't have trailing slash
	backendURL := strings.TrimRight(m.backendURL, "/")
	req, err := http.NewRequest("POST", backendURL+"/api/ipfs/upload-metadata", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send metadata to backend
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to backend: %w", err)
	}
	defer resp.Body.Close()

	// Read response from backend
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read backend response: %w", err)
	}

	// Parse response - backend returns IPFS hash after uploading via Pinata
	var uploadResp IPFSUploadResponse
	if err := json.Unmarshal(respBody, &uploadResp); err != nil {
		return "", fmt.Errorf("failed to parse backend response: %w", err)
	}

	if !uploadResp.Success {
		return "", fmt.Errorf("backend upload failed: %s", uploadResp.Error)
	}

	// Return IPFS URI that backend created
	return fmt.Sprintf("ipfs://%s", uploadResp.IpfsHash), nil
}

// getContractConfig gets the contract configuration from backend
func (m *NFTMinter) getContractConfig() (*ContractConfigResponse, error) {
	// Ensure backend URL doesn't have trailing slash
	backendURL := strings.TrimRight(m.backendURL, "/")
	endpoint := backendURL + "/api/contract/config"
	
	fmt.Printf("   📡 Fetching contract config from: %s\n", endpoint)
	
	// Create request
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Send request
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("backend returned status %d: %s", resp.StatusCode, string(body))
	}

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Log response for debugging if it looks like HTML
	if len(respBody) > 0 && respBody[0] == '<' {
		preview := string(respBody)
		if len(preview) > 100 {
			preview = preview[:100]
		}
		return nil, fmt.Errorf("backend returned HTML instead of JSON. Response starts with: %s", preview)
	}

	// Parse response
	var config ContractConfigResponse
	if err := json.Unmarshal(respBody, &config); err != nil {
		// Include part of the response in error for debugging
		preview := string(respBody)
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		return nil, fmt.Errorf("failed to parse response: %w. Response: %s", err, preview)
	}

	return &config, nil
}

// getNonce gets the current nonce for an address from the contract
func (m *NFTMinter) getNonce(address common.Address) (uint64, error) {
	if m.client == nil {
		// If no Ethereum client, assume nonce is 0 (first mint)
		return 0, nil
	}

	// Parse the contract ABI
	contractABI, err := ParseABI()
	if err != nil {
		return 0, fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Pack the nonces method call
	data, err := contractABI.Pack("nonces", address)
	if err != nil {
		return 0, fmt.Errorf("failed to pack nonces call: %w", err)
	}

	// Call the contract
	result, err := m.client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &m.contractAddress,
		Data: data,
	}, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to call nonces: %w", err)
	}

	// Unpack the result
	var nonce *big.Int
	err = contractABI.UnpackIntoInterface(&nonce, "nonces", result)
	if err != nil {
		return 0, fmt.Errorf("failed to unpack nonce: %w", err)
	}

	return nonce.Uint64(), nil
}

// requestMintSignature requests a mint signature from the backend
func (m *NFTMinter) requestMintSignature(to string, tokenURI string, nonce uint64) (string, error) {
	// Show progress
	fmt.Printf("   📝 Requesting mint signature from backend...\n")
	
	// Prepare request
	// Note: tokenURI is not used in signature generation but sent for compatibility
	reqBody := MintSignatureRequest{
		To:       to,
		TokenURI: "", // Empty as per backend expectation
		Nonce:    nonce,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	// Ensure backend URL doesn't have trailing slash
	backendURL := strings.TrimRight(m.backendURL, "/")
	endpoint := backendURL + "/api/signature/generate-mint"
	
	fmt.Printf("   📡 Sending request to: %s\n", endpoint)
	fmt.Printf("   📦 Request data: to=%s, nonce=%d, tokenURI=\"\" (not used in signature)\n", to, nonce)
	
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("backend returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Log the response status
	fmt.Printf("   📨 Response status: %d\n", resp.StatusCode)
	
	// Check if response is HTML (error page)
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") || (len(respBody) > 0 && respBody[0] == '<') {
		preview := string(respBody)
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		
		// Try to extract meaningful error from HTML
		errorMsg := "Backend returned HTML instead of JSON. "
		if strings.Contains(preview, "404") || strings.Contains(preview, "Not Found") {
			errorMsg += "The mint signature endpoint may not be available at this URL. "
		} else if strings.Contains(preview, "502") || strings.Contains(preview, "Bad Gateway") {
			errorMsg += "The backend server may be down or unreachable. "
		}
		
		fmt.Printf("   ❌ Error: %s\n", errorMsg)
		fmt.Printf("   📄 HTML Response preview:\n%s\n", preview)
		
		return "", fmt.Errorf("%sPlease check the backend URL configuration", errorMsg)
	}
	
	// Don't log raw response as it contains sensitive signature data

	// Parse response
	var sigResp MintSignatureResponse
	if err := json.Unmarshal(respBody, &sigResp); err != nil {
		// Check for common JSON parsing errors
		if strings.Contains(err.Error(), "invalid character 'p'") {
			// This often means we got a plain text error starting with 'p'
			if strings.HasPrefix(string(respBody), "property") || strings.HasPrefix(string(respBody), "please") {
				return "", fmt.Errorf("backend error: %s", string(respBody))
			}
		}
		
		// Include part of the response in error for debugging
		preview := string(respBody)
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		return "", fmt.Errorf("failed to parse JSON response: %w. Response: %s", err, preview)
	}

	// Validate signature response
	if sigResp.Signature == "" {
		return "", fmt.Errorf("backend returned empty signature")
	}
	
	fmt.Printf("   ✅ Received signature successfully\n")
	fmt.Printf("   ✅ Nonce confirmed: %d\n", sigResp.Nonce)
	return sigResp.Signature, nil
}

// executeMint executes the mint transaction on the blockchain
func (m *NFTMinter) executeMint(signature string) (uint64, error) {
	if m.client == nil {
		return 0, fmt.Errorf("ethereum client not initialized")
	}

	// Parse the contract ABI
	contractABI, err := ParseABI()
	if err != nil {
		return 0, fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Decode signature from hex
	sigBytes, err := hex.DecodeString(strings.TrimPrefix(signature, "0x"))
	if err != nil {
		return 0, fmt.Errorf("failed to decode signature: %w", err)
	}

	// Pack the mint method call
	data, err := contractABI.Pack("mint", m.address, sigBytes)
	if err != nil {
		return 0, fmt.Errorf("failed to pack mint call: %w", err)
	}

	// Get the current gas price
	gasPrice, err := m.client.SuggestGasPrice(context.Background())
	if err != nil {
		return 0, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Get the nonce for the transaction
	nonce, err := m.client.PendingNonceAt(context.Background(), m.address)
	if err != nil {
		return 0, fmt.Errorf("failed to get account nonce: %w", err)
	}

	// Create the transaction
	tx := types.NewTransaction(
		nonce,
		m.contractAddress,
		DefaultMintPrice(), // 0.01 ETH mint price
		uint64(300000),     // Gas limit
		gasPrice,
		data,
	)

	// Sign the transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(m.chainID), m.privateKey)
	if err != nil {
		return 0, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send the transaction
	err = m.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return 0, fmt.Errorf("failed to send transaction: %w", err)
	}

	fmt.Printf("Mint transaction sent: %s\n", signedTx.Hash().Hex())

	// Wait for transaction receipt
	receipt, err := m.WaitForTransaction(context.Background(), signedTx)
	if err != nil {
		return 0, fmt.Errorf("failed to wait for transaction: %w", err)
	}

	// Extract token ID from logs
	// The Transfer event has the token ID as the third topic
	for _, log := range receipt.Logs {
		if len(log.Topics) >= 4 && log.Address == m.contractAddress {
			// Transfer event signature
			transferEventSig := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
			if log.Topics[0] == transferEventSig {
				// Token ID is in the third topic
				tokenID := new(big.Int).SetBytes(log.Topics[3].Bytes())
				return tokenID.Uint64(), nil
			}
		}
	}

	// If we couldn't find the token ID in logs, return an error
	return 0, fmt.Errorf("could not extract token ID from transaction logs")
}

// GenerateMetadataHash generates a SHA256 hash of agent metadata
func GenerateMetadataHash(metadata AgentMetadata) string {
	// Create deterministic string representation
	data := fmt.Sprintf("%s:%s:%s:%s",
		metadata.Name,
		metadata.Description,
		metadata.Image,
		strings.Join(metadata.Capabilities, ","))
	
	// Generate SHA256 hash
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// SendMetadataHashToBackend sends the metadata hash for an existing agent
func (m *NFTMinter) SendMetadataHashToBackend(hash string, tokenID uint64, walletAddress string) error {
	// TODO: Implement backend endpoint for metadata hash submission
	// This would send the hash along with the token ID to verify ownership
	
	reqBody := map[string]interface{}{
		"hash":          hash,
		"tokenId":       tokenID,
		"walletAddress": walletAddress,
	}

	_, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// For now, we'll just log this operation
	// In production, this would make an actual HTTP request
	fmt.Printf("Would send metadata hash: %s for token ID: %d\n", hash, tokenID)
	
	return nil
}

// GetAddress returns the address associated with the minter
func (m *NFTMinter) GetAddress() common.Address {
	return m.address
}

// WaitForTransaction waits for a transaction to be confirmed
func (m *NFTMinter) WaitForTransaction(ctx context.Context, tx *types.Transaction) (*types.Receipt, error) {
	// Wait for transaction receipt
	receipt, err := bind.WaitMined(ctx, m.client, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for transaction: %w", err)
	}

	// Check if transaction was successful
	if receipt.Status == 0 {
		return nil, fmt.Errorf("transaction failed")
	}

	return receipt, nil
}