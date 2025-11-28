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

var portfolioHTTPClient = &http.Client{Timeout: 15 * time.Second}

// Disederhanakan, fokus ke total value + per chain
type debankPortfolioResponse struct {
	Data struct {
		TotalUsdValue float64 `json:"total_usd_value"`
		ChainList     []struct {
			ID                string  `json:"id"`
			Name              string  `json:"name"`
			PortfolioUsdValue float64 `json:"portfolio_usd_value"`
		} `json:"chain_list"`
	} `json:"data"`
}

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
		return "Portfolio feature is not configured (missing COVALENT_API_KEY in environment).", nil
	}

	base := "https://api.debank.com/user/total_balance"
	u := base + "?id=" + url.QueryEscape(address)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")
	req.Header.Set("AccessKey", apiKey)

	resp, err := portfolioHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("debank error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("debank status %d (check API key & IP allowlist)", resp.StatusCode)
	}

	var data debankPortfolioResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("decode debank error: %w", err)
	}

	total := data.Data.TotalUsdValue
	if total == 0 && len(data.Data.ChainList) == 0 {
		return fmt.Sprintf("No portfolio data found for address %s", address), nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "📊 Portfolio summary\nAddress: `%s`\n\n", address)
	fmt.Fprintf(&b, "Total value: $%.2f\n\n", total)

	if len(data.Data.ChainList) > 0 {
		b.WriteString("Per chain:\n")
		for _, c := range data.Data.ChainList {
			if c.PortfolioUsdValue <= 0 {
				continue
			}
			fmt.Fprintf(&b, "• %s: $%.2f\n", titleCase(c.Name), c.PortfolioUsdValue)
		}
	}

	b.WriteString("\nSource: DeBank (approximate, portfolio view only)")
	return b.String(), nil
}
