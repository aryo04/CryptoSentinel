package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// HTTP client khusus portfolio (Covalent)
var portfolioHTTPClient = &http.Client{Timeout: 20 * time.Second}

// Daftar chain yang discan oleh CryptoSentinel AI.
// Tambah/kurangi sesuai kebutuhan (harus chain EVM yang didukung Covalent).
var covalentChains = []struct {
	Name    string // label buat user
	ChainID int    // chain_id Covalent
}{
	{"Ethereum", 1},
	{"Polygon", 137},
	{"BSC", 56},
	{"Arbitrum", 42161},
	{"Optimism", 10},
	{"Base", 8453},
	{"Avalanche", 43114},
}

// Struktur respons Covalent yang kita pakai (disimplify)
type covalentBalancesResponse struct {
	Data struct {
		ChainID   int    `json:"chain_id"`
		ChainName string `json:"chain_name"`
		Items     []struct {
			Balance          string  `json:"balance"`
			ContractDecimals int     `json:"contract_decimals"`
			Quote            float64 `json:"quote"` // USD value
		} `json:"items"`
	} `json:"data"`
	Error        bool   `json:"error"`
	ErrorMessage string `json:"error_message"`
	ErrorCode    int    `json:"error_code"`
}

// CmdPortfolio men-scan beberapa EVM chain via Covalent dan memberi ringkasan nilai USD.
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

	type chainPortfolio struct {
		Name  string
		Value float64
	}

	var (
		perChain   []chainPortfolio
		totalValue float64
	)

	// Loop semua chain yang kita dukung
	for _, ch := range covalentChains {
		// Contoh endpoint:
		// https://api.covalenthq.com/v1/1/address/0x.../balances_v2/?key=API_KEY&quote-currency=USD&nft=false&no-nft-fetch=true
		u := fmt.Sprintf(
			"https://api.covalenthq.com/v1/%d/address/%s/balances_v2/?quote-currency=USD&nft=false&no-nft-fetch=true&key=%s",
			ch.ChainID,
			url.PathEscape(address),
			url.QueryEscape(apiKey),
		)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			// Kalau request gagal dibuat, skip chain ini saja
			continue
		}
		req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")

		resp, err := portfolioHTTPClient.Do(req)
		if err != nil {
			// Network / DNS error, skip chain
			continue
		}
		func() {
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				// Kalau token expired / key salah, Covalent akan balas non-200.
				// Kita tidak return error ke user, tapi skip chain ini.
				return
			}

			var data covalentBalancesResponse
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				return
			}
			if data.Error {
				// Error level Covalent (misalnya chain belum didukung)
				return
			}

			// Hitung total USD untuk chain ini
			var chainTotal float64
			for _, it := range data.Data.Items {
				if it.Quote <= 0 {
					continue
				}
				chainTotal += it.Quote
			}
			if chainTotal <= 0 {
				return
			}

			perChain = append(perChain, chainPortfolio{
				Name:  ch.Name,
				Value: chainTotal,
			})
			totalValue += chainTotal
		}()
	}

	if len(perChain) == 0 || totalValue <= 0 {
		return fmt.Sprintf(
			"No portfolio data found for address %s on the scanned chains.",
			address,
		), nil
	}

	// Sort per-chain desc by value (optional, tapi rapi)
	// Simple bubble-ish sort (biar tanpa import sort kalau kamu mau minimal).
	for i := 0; i < len(perChain); i++ {
		for j := i + 1; j < len(perChain); j++ {
			if perChain[j].Value > perChain[i].Value {
				perChain[i], perChain[j] = perChain[j], perChain[i]
			}
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "📊 Portfolio summary\nAddress: `%s`\n\n", address)
	fmt.Fprintf(&b, "Total value (approx): **$%s**\n\n", formatUSD(totalValue))

	b.WriteString("Per chain:\n")
	for _, c := range perChain {
		fmt.Fprintf(&b, "• %s: $%s\n", c.Name, formatUSD(c.Value))
	}

	b.WriteString("\nSource: Covalent (USD estimate, EVM chains only)")
	return b.String(), nil
}

// formatUSD memformat angka float ke string dengan koma ribuan, 2 desimal
func formatUSD(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	parts := strings.Split(s, ".")
	intPart := parts[0]
	decPart := ""
	if len(parts) > 1 {
		decPart = parts[1]
	}

	// tambahkan koma ribuan ke intPart
	n := len(intPart)
	if n <= 3 {
		if decPart == "" {
			return intPart
		}
		return intPart + "." + decPart
	}
	var chunks []string
	for n > 3 {
		chunks = append([]string{intPart[n-3:]}, chunks...)
		n -= 3
	}
	chunks = append([]string{intPart[:n]}, chunks...)
	intPart = strings.Join(chunks, ",")

	if decPart == "" {
		return intPart
	}
	return intPart + "." + decPart
}
