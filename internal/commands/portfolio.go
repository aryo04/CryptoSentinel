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

var portfolioHTTP = &http.Client{Timeout: 25 * time.Second}

// Struktur response Covalent allchains balances (disederhanakan ke field yang kita pakai)
type covalentBalancesResponse struct {
	Data struct {
		Address       string `json:"address"`
		UpdatedAt     string `json:"updated_at"`
		QuoteCurrency string `json:"quote_currency"`
		Items         []struct {
			ChainName           string  `json:"chain_name"`
			ContractDisplayName string  `json:"contract_display_name"`
			ContractTicker      string  `json:"contract_ticker_symbol"`
			Quote               float64 `json:"quote"`
		} `json:"items"`
	} `json:"data"`
	Error        bool   `json:"error"`
	ErrorMessage string `json:"error_message"`
}

type tokenSnapshot struct {
	Chain  string
	Name   string
	Symbol string
	Value  float64
}

// CmdPortfolio
// Usage:
//   portfolio [address]
//   portfolio [address] [chain]
// Example:
//   portfolio 0x1234...
//   portfolio 0x1234... arbitrum
func CmdPortfolio(ctx context.Context, args []string) (string, error) {
	if len(args) == 0 {
		return "Usage: portfolio [address] [optional_chain]\nExamples:\n  portfolio 0x1234...\n  portfolio 0x1234... arbitrum", nil
	}

	addr := strings.TrimSpace(args[0])
	if addr == "" || !strings.HasPrefix(addr, "0x") {
		return "Invalid wallet address. Use full 0x... address.", nil
	}

	// chain optional (buat filter output aja, API tetap allchains)
	filterChain := ""
	if len(args) > 1 {
		filterChain = strings.ToLower(strings.TrimSpace(args[1]))
	}

	apiKey := os.Getenv("COVALENT_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("missing COVALENT_API_KEY in environment")
	}

	// Minta balances untuk beberapa L1/L2 populer
	chainsParam := "eth-mainnet,arbitrum-mainnet,polygon-mainnet,bsc-mainnet,base-mainnet,optimism-mainnet,avalanche-mainnet"

	endpoint := fmt.Sprintf(
		"https://api.covalenthq.com/v1/allchains/address/%s/balances/?quote-currency=USD&limit=200&chains=%s",
		url.PathEscape(addr),
		url.QueryEscape(chainsParam),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")

	resp, err := portfolioHTTP.Do(req)
	if err != nil {
		if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
			return "⏱️ Portfolio request ke Covalent timeout. Jaringan lagi lambat atau alamat punya token sangat banyak. Coba lagi sebentar lagi.", nil
		}
		return "", fmt.Errorf("portfolio API error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return "🚫 Covalent returned 403 untuk balances endpoint.\nCek API key / plan / IP allowlist di dashboard Covalent.", nil
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Covalent status %d", resp.StatusCode)
	}

	var payload covalentBalancesResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode balances JSON error: %w", err)
	}
	if payload.Error {
		if payload.ErrorMessage != "" {
			return fmt.Sprintf("Covalent error: %s", payload.ErrorMessage), nil
		}
		return "Covalent returned an error for this portfolio request.", nil
	}

	if len(payload.Data.Items) == 0 {
		return fmt.Sprintf("No portfolio data found for address %s on the scanned chains.", addr), nil
	}

	var (
		snapshots []tokenSnapshot
		total     float64
		perChain  = map[string]float64{}
	)

	for _, it := range payload.Data.Items {
		if it.Quote <= 0 {
			continue
		}

		chain := it.ChainName
		if chain == "" {
			chain = "unknown"
		}

		// filter chain kalau user kasih argumen chain kedua
		if filterChain != "" &&
			!strings.EqualFold(chain, filterChain) &&
			!strings.EqualFold(chainAlias(chain), filterChain) {
			continue
		}

		name := it.ContractDisplayName
		if name == "" {
			name = strings.ToUpper(it.ContractTicker)
		}

		snapshots = append(snapshots, tokenSnapshot{
			Chain:  chain,
			Name:   name,
			Symbol: it.ContractTicker,
			Value:  it.Quote,
		})

		total += it.Quote
		perChain[chain] += it.Quote
	}

	if len(snapshots) == 0 {
		if filterChain != "" {
			return fmt.Sprintf("No non-zero balances found for address %s on chain %s.", addr, filterChain), nil
		}
		return fmt.Sprintf("No non-zero balances found for address %s on the scanned chains.", addr), nil
	}

	// sort token dari paling besar
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Value > snapshots[j].Value
	})

	topN := 10
	if len(snapshots) < topN {
		topN = len(snapshots)
	}

	var b strings.Builder
	b.WriteString("📊 **Portfolio snapshot (multichain)**\n")
	b.WriteString(fmt.Sprintf("Address: `%s`\n", addr))
	b.WriteString(fmt.Sprintf("Quote: **%s**\n", payload.Data.QuoteCurrency))
	if payload.Data.UpdatedAt != "" {
		b.WriteString(fmt.Sprintf("Updated: %s\n", payload.Data.UpdatedAt))
	}
	if filterChain != "" {
		b.WriteString(fmt.Sprintf("Filter chain: %s\n", filterChain))
	}
	b.WriteString("\n")

	b.WriteString(fmt.Sprintf("Estimated total value: **$%s**\n\n", services.F(total)))

	// ringkasan per-chain
	if len(perChain) > 1 {
		b.WriteString("Per-chain breakdown:\n")
		type chainPair struct {
			Name  string
			Value float64
		}
		cp := make([]chainPair, 0, len(perChain))
		for name, v := range perChain {
			cp = append(cp, chainPair{name, v})
		}
		sort.Slice(cp, func(i, j int) bool { return cp[i].Value > cp[j].Value })
		for _, c := range cp {
			b.WriteString(fmt.Sprintf("• %s — $%s\n", c.Name, services.F(c.Value)))
		}
		b.WriteString("\n")
	}

	b.WriteString(fmt.Sprintf("Top %d positions:\n", topN))
	for i := 0; i < topN; i++ {
		t := snapshots[i]
		sym := strings.ToUpper(t.Symbol)
		if sym == "" {
			sym = "—"
		}
		b.WriteString(fmt.Sprintf(
			"%d) [%s] %s (%s) — $%s\n",
			i+1,
			t.Chain,
			t.Name,
			sym,
			services.F(t.Value),
		))
	}

	b.WriteString("\nSource: Covalent allchains balances API")

	return b.String(), nil
}

// bantu cocokkan nama chain panjang → alias pendek (biar filter "arb"/"arbitrum" tetap nyangkut)
func chainAlias(chain string) string {
	l := strings.ToLower(chain)
	switch l {
	case "eth-mainnet":
		return "eth"
	case "polygon-mainnet":
		return "polygon"
	case "arbitrum-mainnet":
		return "arbitrum"
	case "bsc-mainnet":
		return "bsc"
	case "optimism-mainnet":
		return "optimism"
	case "base-mainnet":
		return "base"
	case "avalanche-mainnet":
		return "avalanche"
	default:
		return l
	}
}
