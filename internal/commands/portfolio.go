package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"teneo-agent-demo1/internal/services"
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

// Struktur response Covalent portfolio_v2 (disederhanakan)
type covalentPortfolioResponse struct {
	Data struct {
		Address       string `json:"address"`
		UpdatedAt     string `json:"updated_at"`
		QuoteCurrency string `json:"quote_currency"`
		ChainName     string `json:"chain_name"`
		Items         []struct {
			ContractName         string `json:"contract_name"`
			ContractTickerSymbol string `json:"contract_ticker_symbol"`
			// holdings berisi history per hari
			Holdings []struct {
				Timestamp string  `json:"timestamp"`
				Quote     float64 `json:"quote"` // beberapa chain langsung punya field ini
				Close     struct {
					Quote float64 `json:"quote"` // di chain lain, quote ada di dalam close.quote
				} `json:"close"`
			} `json:"holdings"`
		} `json:"items"`
	} `json:"data"`
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

	// Ambil 14 hari terakhir, quote USD
	endpoint := fmt.Sprintf(
		"https://api.covalenthq.com/v1/%s/address/%s/portfolio_v2/?days=14&quote-currency=USD",
		chainInput,
		url.PathEscape(address),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")

	resp, err := portfolioHTTP.Do(req)
	if err != nil {
		// Deteksi timeout
		if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
			return "⏱️ Covalent portfolio_v2 request timed out. Jaringan sedang lambat atau alamat punya data sangat besar. Coba lagi beberapa saat lagi.", nil
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

	var body covalentPortfolioResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("decode portfolio JSON error: %w", err)
	}

	items := body.Data.Items
	if len(items) == 0 {
		return fmt.Sprintf("No portfolio data found for address %s on %s.", address, chainInput), nil
	}

	// Hitung total dan ringkasan per token berdasarkan holding terakhir
	type tokenSummary struct {
		Name   string
		Symbol string
		Value  float64
	}

	var (
		total   float64
		tokens  []tokenSummary
		nonZero int
	)

	for _, it := range items {
		if len(it.Holdings) == 0 {
			continue
		}
		last := it.Holdings[len(it.Holdings)-1]

		q := last.Quote
		if q == 0 && last.Close.Quote != 0 {
			q = last.Close.Quote
		}
		if q <= 0 {
			continue
		}

		nonZero++
		total += q
		tokens = append(tokens, tokenSummary{
			Name:   it.ContractName,
			Symbol: it.ContractTickerSymbol,
			Value:  q,
		})
	}

	if nonZero == 0 {
		return fmt.Sprintf(
			"Portfolio appears empty or Covalent returned 0 USD value for address %s on %s.\nThis can happen if there are only small/illiquid tokens that do not have a fiat price on Covalent.",
			address, chainInput,
		), nil
	}

	// Urutkan token dari yang paling besar
	sort.Slice(tokens, func(i, j int) bool {
		return tokens[i].Value > tokens[j].Value
	})

	// Batasi top N token
	topN := 10
	if len(tokens) < topN {
		topN = len(tokens)
	}

	var b strings.Builder
	b.WriteString("📊 **Portfolio snapshot**\n")
	b.WriteString(fmt.Sprintf("Address: `%s`\n", address))
	b.WriteString(fmt.Sprintf("Chain: **%s**\n", body.Data.ChainName))
	b.WriteString(fmt.Sprintf("Quote: **%s**\n", body.Data.QuoteCurrency))
	if body.Data.UpdatedAt != "" {
		b.WriteString(fmt.Sprintf("Updated: %s\n", body.Data.UpdatedAt))
	}
	b.WriteString("\n")

	b.WriteString(fmt.Sprintf("Total value (approx): **$%s**\n\n", services.F(total)))

	b.WriteString(fmt.Sprintf("Top %d tokens by value:\n", topN))
	for i := 0; i < topN; i++ {
		t := tokens[i]
		name := t.Name
		if name == "" {
			name = strings.ToUpper(t.Symbol)
		}
		b.WriteString(fmt.Sprintf(
			"%d) %s (%s) — $%s\n",
			i+1,
			name,
			strings.ToUpper(t.Symbol),
			services.F(t.Value),
		))
	}

	b.WriteString("\nSource: Covalent Portfolio v2 API")

	return b.String(), nil
}
