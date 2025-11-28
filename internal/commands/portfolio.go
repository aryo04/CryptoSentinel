package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

var portfolioHTTPClient = &http.Client{Timeout: 20 * time.Second}

// Struktur response untuk /v1/allchains/address/{walletAddress}/balances/
type allChainsBalancesResponse struct {
	UpdatedAt    string `json:"updated_at"`
	CursorBefore string `json:"cursor_before"`
	CursorAfter  string `json:"cursor_after"`

	QuoteCurrency string `json:"quote_currency"`

	Items []struct {
		ChainName           string  `json:"chain_name"`
		ContractName        string  `json:"contract_name"`
		ContractTicker      string  `json:"contract_ticker_symbol"`
		Type                string  `json:"type"` // "cryptocurrency", "dust", "nft", dst.
		Balance             string  `json:"balance"`
		ContractDecimals    int     `json:"contract_decimals"`
		Quote               float64 `json:"quote"`      // total value token dlm quote_currency
		QuoteRate           float64 `json:"quote_rate"` // harga per 1 token
		LastTransferredAt   string  `json:"last_transferred_at"`
		IsSpam              bool    `json:"is_spam"`
		IsNft               bool    `json:"is_nft"`
		Symbol              string  `json:"contract_ticker_symbol"`
		NativeToken         bool    `json:"is_native_token"`
		SupportsErc         []string `json:"supports_erc"`
	} `json:"items"`

	Error        bool   `json:"error"`
	ErrorMessage string `json:"error_message"`
}

type chainPortfolio struct {
	Chain string
	USD   float64
}

// CmdPortfolio
//   portfolio [address]
//   Example: portfolio 0x1234...abcd
func CmdPortfolio(ctx context.Context, args []string) (string, error) {
	if len(args) < 1 {
		return "Usage: portfolio [address]\nExample: portfolio 0x1234...abcd", nil
	}
	address := strings.TrimSpace(args[0])
	if address == "" {
		return "Wallet address is required.", nil
	}

	apiKey := os.Getenv("COVALENT_API_KEY")
	if apiKey == "" {
		return "", errors.New("missing COVALENT_API_KEY in environment")
	}

	// allchains endpoint — otomatis scan beberapa EVM chain
	u := fmt.Sprintf(
		"https://api.covalenthq.com/v1/allchains/address/%s/balances/?quote-currency=USD&limit=100",
		url.PathEscape(address),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := portfolioHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Covalent request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Covalent status %d (check API key / plan / IP allowlist)", resp.StatusCode)
	}

	var payload allChainsBalancesResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode Covalent response error: %v", err)
	}

	if payload.Error {
		msg := payload.ErrorMessage
		if msg == "" {
			msg = "unknown Covalent API error"
		}
		return fmt.Sprintf("Covalent error while loading portfolio: %s", msg), nil
	}

	if len(payload.Items) == 0 {
		return fmt.Sprintf("No portfolio data found for address %s on the scanned chains.", address), nil
	}

	// Akumulasi total per-chain
	byChain := make(map[string]float64)
	var totalUSD float64

	for _, it := range payload.Items {
		// Skip NFT / spam / balance nol
		if it.IsNft || it.IsSpam || it.Balance == "0" {
			continue
		}

		value := it.Quote
		if (value <= 0 || math.IsNaN(value) || math.IsInf(value, 0)) && it.QuoteRate > 0 {
			// Hitung manual dari balance * quote_rate
			if bal, err := parseScaledFloat(it.Balance, it.ContractDecimals); err == nil {
				value = bal * it.QuoteRate
			}
		}

		if value <= 0 || math.IsNaN(value) || math.IsInf(value, 0) {
			continue
		}

		chain := it.ChainName
		if chain == "" {
			chain = "Unknown"
		}
		byChain[chain] += value
		totalUSD += value
	}

	if totalUSD <= 0 {
		return fmt.Sprintf(
			"No priced portfolio data found for address %s on the scanned chains.\n"+
				"This often happens if assets are on unsupported networks or pricing is unavailable.",
			address,
		), nil
	}

	// Susun list per-chain & sort desc
	var chains []chainPortfolio
	for ch, v := range byChain {
		if v <= 0 {
			continue
		}
		chains = append(chains, chainPortfolio{Chain: ch, USD: v})
	}
	sort.Slice(chains, func(i, j int) bool {
		return chains[i].USD > chains[j].USD
	})

	var b strings.Builder
	fmt.Fprintf(&b, "📊 Portfolio summary (Covalent allchains)\nAddress: `%s`\n\n", address)
	fmt.Fprintf(&b, "Estimated total value: **$%s**\n\n", formatUSD(totalUSD))

	if len(chains) > 0 {
		b.WriteString("Per chain:\n")
		for _, c := range chains {
			fmt.Fprintf(&b, "• %s: $%s\n", c.Chain, formatUSD(c.USD))
		}
		b.WriteString("\n")
	}

	b.WriteString("Note: Values are approximate and only include EVM chains and tokens covered by Covalent.\n")
	if payload.UpdatedAt != "" {
		fmt.Fprintf(&b, "Last updated: %s\n", payload.UpdatedAt)
	}
	b.WriteString("Source: Covalent GoldRush allchains/balances")

	return b.String(), nil
}

// parseScaledFloat: balance (string integer) + decimals → float64
// contoh: "1230000000000000000", 18 → 1.23
func parseScaledFloat(balance string, decimals int) (float64, error) {
	i := new(big.Int)
	if _, ok := i.SetString(balance, 10); !ok {
		return 0, fmt.Errorf("invalid balance %q", balance)
	}
	if decimals <= 0 {
		return new(big.Float).SetInt(i).Float64()
	}

	f := new(big.Float).SetInt(i)
	denom := new(big.Float).SetFloat64(math.Pow10(decimals))
	f.Quo(f, denom)
	out, _ := f.Float64()
	return out, nil
}

// formatUSD: 1234567.89 → "1,234,567.89"
func formatUSD(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	parts := strings.SplitN(s, ".", 2)
	intPart := parts[0]
	fracPart := ""
	if len(parts) == 2 {
		fracPart = parts[1]
	}

	n := len(intPart)
	if n <= 3 {
		if fracPart == "" {
			return intPart
		}
		return intPart + "." + fracPart
	}

	var chunks []string
	for n > 3 {
		chunks = append([]string{intPart[n-3 : n]}, chunks...)
		n -= 3
	}
	chunks = append([]string{intPart[:n]}, chunks...)
	intWithCommas := strings.Join(chunks, ",")

	if fracPart == "" {
		return intWithCommas
	}
	return intWithCommas + "." + fracPart
}
