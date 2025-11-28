package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var portfolioHTTP = &http.Client{Timeout: 45 * time.Second}

// Mapping input user -> chainName Covalent
var supportedPortfolioChains = map[string]string{
	"eth":       "eth-mainnet",
	"ethereum":  "eth-mainnet",
	"polygon":   "polygon-mainnet",
	"matic":     "polygon-mainnet",
	"arb":       "arbitrum-mainnet",
	"arbitrum":  "arbitrum-mainnet",
	"bsc":       "bsc-mainnet",
	"bnb":       "bsc-mainnet",
	"op":        "optimism-mainnet",
	"optimism":  "optimism-mainnet",
	"base":      "base-mainnet",
	"avax":      "avalanche-mainnet",
	"avalanche": "avalanche-mainnet",
}

// Struktur response data utama dari Covalent portfolio_v2
type portfolioV2Response struct {
	Address       string `json:"address"`
	UpdatedAt     string `json:"updated_at"`
	QuoteCurrency string `json:"quote_currency"`
	ChainName     string `json:"chain_name"`
	Items         []struct {
		Date       string  `json:"date"`
		TotalValue float64 `json:"total_value"`
	} `json:"items"`
}

// CmdPortfolio
// Usage: portfolio [wallet_address] [chain]
// Example: portfolio 0x1234... eth
func CmdPortfolio(ctx context.Context, args []string) (string, error) {
	if len(args) == 0 {
		return "Usage: portfolio [wallet_address] [chain]\nExample: portfolio 0x1234... eth", nil
	}

	address := strings.TrimSpace(args[0])
	if address == "" {
		return "Wallet address is required.", nil
	}
	if !strings.HasPrefix(address, "0x") {
		return "Invalid address format. Must start with 0x.", nil
	}

	// Chain optional, default: Ethereum mainnet
	chainInput := "eth-mainnet"
	if len(args) > 1 {
		key := strings.ToLower(strings.TrimSpace(args[1]))
		if mapped, ok := supportedPortfolioChains[key]; ok {
			chainInput = mapped
		} else {
			return "Unsupported chain. Try one of: eth, polygon, arbitrum, bsc, optimism, base, avalanche.", nil
		}
	}

	apiKey := os.Getenv("COVALENT_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("missing COVALENT_API_KEY in environment")
	}

	// Batasi ke 14 hari terakhir & quote USD, biar response tidak terlalu berat
	baseURL := fmt.Sprintf(
		"https://api.covalenthq.com/v1/%s/address/%s/portfolio_v2/?days=14&quote-currency=USD",
		chainInput,
		url.PathEscape(address),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")

	resp, err := portfolioHTTP.Do(req)
	if err != nil {
		// Deteksi timeout dari http.Client (Client.Timeout exceeded)
		if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
			return "⏱️ Portfolio request to Covalent timed out.\nAlamat ini kemungkinan sangat besar atau jaringan sedang lambat. Coba lagi nanti atau gunakan chain lain (misalnya polygon-mainnet, arbitrum-mainnet).", nil
		}
		return "", fmt.Errorf("portfolio API error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return "🚫 Covalent returned 403 for portfolio_v2.\nPeriksa: API key, plan, atau batasan penggunaan di dashboard Covalent.", nil
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Covalent status %d", resp.StatusCode)
	}

	var body struct {
		Data portfolioV2Response `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("decode portfolio JSON error: %w", err)
	}

	items := body.Data.Items
	if len(items) == 0 {
		return fmt.Sprintf("No portfolio data found for address %s on %s.", address, chainInput), nil
	}

	var b strings.Builder
	b.WriteString("📊 **Portfolio value history**\n")
	b.WriteString(fmt.Sprintf("Address: `%s`\nChain: **%s**\nQuote: **%s**\n\n",
		address,
		body.Data.ChainName,
		body.Data.QuoteCurrency,
	))

	// Ambil max 10 hari terakhir, dari yang terbaru ke yang lebih lama
	limit := 10
	if len(items) < limit {
		limit = len(items)
	}

	b.WriteString("**Last 10 days (most recent first):**\n")

	start := len(items) - limit
	for i := len(items) - 1; i >= start; i-- {
		d := items[i]
		b.WriteString(fmt.Sprintf("• %s → $%.2f\n", d.Date, d.TotalValue))
	}

	b.WriteString("\nSource: Covalent Portfolio v2 API")

	return b.String(), nil
}
