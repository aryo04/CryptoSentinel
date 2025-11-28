package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"teneo-agent-demo1/internal/services"
)

var portfolioHTTPClient = &http.Client{Timeout: 20 * time.Second}

// struktur respons Covalent (dipotong, hanya field penting)
type covalentBalancesResponse struct {
	Data struct {
		Address       string `json:"address"`
		QuoteCurrency string `json:"quote_currency"`
		ChainID       int    `json:"chain_id"`
		Items         []struct {
			Quote float64 `json:"quote"` // nilai USD per token
		} `json:"items"`
	} `json:"data"`
	Error        bool   `json:"error"`
	ErrorMessage string `json:"error_message"`
}

// chain yang akan di-scan (multi-chain summary)
var covalentChains = []struct {
	Name string // label ke user
	ID   string // chain_id di URL
}{
	{"Ethereum", "1"},
	{"BSC", "56"},
	{"Polygon", "137"},
	{"Arbitrum", "42161"},
	{"Optimism", "10"},
	{"Avalanche", "43114"},
	{"Fantom", "250"},
	{"Base", "8453"},
}

// CmdPortfolio: portfolio [address]
// contoh: portfolio 0x1234....abcd
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

	type chainSummary struct {
		Name  string
		Total float64
	}

	var summaries []chainSummary
	var grandTotal float64

	for _, ch := range covalentChains {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		u := fmt.Sprintf(
			"https://api.covalenthq.com/v1/%s/address/%s/balances_v2/?quote-currency=USD&nft=false&no-nft-fetch=true",
			ch.ID,
			url.PathEscape(address),
		)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")
		// gaya baru Covalent: Authorization: Bearer <key>
		req.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err := portfolioHTTPClient.Do(req)
		if err != nil {
			continue
		}
		func() {
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return
			}

			var data covalentBalancesResponse
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				return
			}
			if data.Error {
				return
			}

			var chainTotal float64
			for _, it := range data.Data.Items {
				chainTotal += it.Quote
			}
			if chainTotal <= 0 {
				return
			}

			summaries = append(summaries, chainSummary{
				Name:  ch.Name,
				Total: chainTotal,
			})
			grandTotal += chainTotal
		}()
	}

	if len(summaries) == 0 {
		return fmt.Sprintf("No portfolio data found for address %s on the scanned chains.", address), nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "📊 Portfolio summary\nAddress: `%s`\n\n", address)
	fmt.Fprintf(&b, "Total value (across %d chains): **$%s**\n\n", len(summaries), services.F(grandTotal))

	b.WriteString("Per chain:\n")
	for _, s := range summaries {
		fmt.Fprintf(&b, "• %s: $%s\n", s.Name, services.F(s.Total))
	}

	b.WriteString("\nSource: Covalent (approximate portfolio view)")
	return b.String(), nil
}
