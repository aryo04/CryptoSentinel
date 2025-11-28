package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var portfolioHTTP = &http.Client{Timeout: 15 * time.Second}

var supportedPortfolioChains = map[string]string{
	"eth":      "eth-mainnet",
	"ethereum": "eth-mainnet",
	"polygon":  "polygon-mainnet",
	"matic":    "polygon-mainnet",
	"arb":      "arbitrum-mainnet",
	"arbitrum": "arbitrum-mainnet",
	"bsc":      "bsc-mainnet",
	"bnb":      "bsc-mainnet",
	"op":       "optimism-mainnet",
	"optimism": "optimism-mainnet",
	"base":     "base-mainnet",
	"avax":     "avalanche-mainnet",
	"avalanche": "avalanche-mainnet",
}

// JSON RESPONSE STRUCT
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

func CmdPortfolio(ctx context.Context, args []string) (string, error) {
	if len(args) == 0 {
		return "Usage: portfolio [wallet_address] [chain]\nExample: portfolio 0x1234 eth", nil
	}

	address := args[0]
	if !strings.HasPrefix(address, "0x") {
		return "Invalid address format. Must start with 0x.", nil
	}

	// Chain optional — default ETH
	chainInput := "eth-mainnet"
	if len(args) > 1 {
		key := strings.ToLower(args[1])
		if m, ok := supportedPortfolioChains[key]; ok {
			chainInput = m
		} else {
			return "Unsupported chain. Try: eth, polygon, arbitrum, bsc, optimism, base.", nil
		}
	}

	apiKey := os.Getenv("COVALENT_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("missing COVALENT_API_KEY in environment")
	}

	baseURL := fmt.Sprintf("https://api.covalenthq.com/v1/%s/address/%s/portfolio_v2/", chainInput, url.QueryEscape(address))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")

	resp, err := portfolioHTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("portfolio API error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 {
		return "🚫 Portfolio endpoint refused access (403).\nMake sure your Covalent API key supports portfolio_v2.", nil
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Covalent status %d", resp.StatusCode)
	}

	var body struct {
		Data portfolioV2Response `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("decode portfolio error: %w", err)
	}

	items := body.Data.Items
	if len(items) == 0 {
		return fmt.Sprintf("No portfolio data found for address %s on %s.", address, chainInput), nil
	}

	var b strings.Builder
	b.WriteString("📊 **Portfolio value history**\n")
	b.WriteString(fmt.Sprintf("Address: `%s`\nChain: **%s**\n\n", address, chainInput))

	limit := 10
	if len(items) < limit {
		limit = len(items)
	}

	b.WriteString("**Last 10 days:**\n")
	for i := 0; i < limit; i++ {
		d := items[i]
		b.WriteString(fmt.Sprintf("• %s → $%.2f\n", d.Date, d.TotalValue))
	}

	b.WriteString("\nSource: Covalent Portfolio v2 API")

	return b.String(), nil
}
