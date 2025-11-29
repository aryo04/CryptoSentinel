package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

// HTTP client khusus portfolio (Covalent)
var portfolioHTTPClient = &http.Client{
	Timeout: 90 * time.Second, // 1 menit
}

// Response struct disederhanakan sesuai /v1/allchains/address/{walletAddress}/balances/
type covalentBalancesResponse struct {
	Data struct {
		UpdatedAt     string `json:"updated_at"`
		CursorBefore  any    `json:"cursor_before"`
		QuoteCurrency string `json:"quote_currency"`
		Items         []struct {
			ContractDecimals     int     `json:"contract_decimals"`
			ContractName         string  `json:"contract_name"`
			ContractTickerSymbol string  `json:"contract_ticker_symbol"`
			ContractAddress      string  `json:"contract_address"`
			Type                 string  `json:"type"`
			Balance              string  `json:"balance"`
			QuoteRate            float64 `json:"quote_rate"`
			Quote                float64 `json:"quote"`
			PrettyQuote          string  `json:"pretty_quote"`
			ChainID              string  `json:"chain_id"`
			ChainName            string  `json:"chain_name"`
			ChainDisplayName     string  `json:"chain_display_name"`
			IsNativeToken        bool    `json:"is_native_token"`
		} `json:"items"`
	} `json:"data"`
	Error        bool    `json:"error"`
	ErrorMessage *string `json:"error_message"`
	ErrorCode    *int    `json:"error_code"`
}

// CmdPortfolio
//
// Usage dasar:
//   portfolio [address]
// Contoh:
//   portfolio 0x1234...
//
// Opsional: spesifikasikan chains (pakai nama chain Covalent) dengan argumen ke-2:
//   portfolio 0x1234... eth-mainnet
//   portfolio 0x1234... eth-mainnet,arbitrum-mainnet
func CmdPortfolio(ctx context.Context, args []string) (string, error) {
	if len(args) < 1 {
		return "Usage: portfolio [address]\nExample: portfolio 0x1234...abcd", nil
	}

	address := strings.TrimSpace(args[0])
	if address == "" {
		return "Wallet address is required.\nExample: portfolio 0x1234...abcd", nil
	}

	apiKey := os.Getenv("COVALENT_API_KEY")
	if apiKey == "" {
		return "", errors.New("missing COVALENT_API_KEY in environment")
	}

	// Base URL multichain balances
	baseURL := fmt.Sprintf("https://api.covalenthq.com/v1/allchains/address/%s/balances/", url.PathEscape(address))

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("quote-currency", "USD")

	// Kalau user kasih argumen kedua, treat sebagai nilai "chains" langsung
	// (user bisa isi: eth-mainnet,arbitrum-mainnet,base-mainnet, dll)
	if len(args) >= 2 {
		chainsArg := strings.TrimSpace(args[1])
		if chainsArg != "" {
			q.Set("chains", chainsArg)
		}
	}

	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := portfolioHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("portfolio API error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Tambahkan sedikit info status agar user ngerti ini masalah key/plan/jaringan
		return fmt.Sprintf("❌ Error: Covalent status %d", resp.StatusCode), nil
	}

	var data covalentBalancesResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("decode portfolio JSON error: %w", err)
	}

	if data.Error {
		msg := "unknown error"
		if data.ErrorMessage != nil {
			msg = *data.ErrorMessage
		}
		return fmt.Sprintf("❌ Covalent error: %s", msg), nil
	}

	if len(data.Data.Items) == 0 {
		return fmt.Sprintf("No portfolio data found for address %s on the scanned chains.", address), nil
	}

	quoteCurrency := data.Data.QuoteCurrency
	if quoteCurrency == "" {
		quoteCurrency = "USD"
	}

	// Agregasi total & per chain, plus kumpulan token untuk top holdings
	type tokenHolding struct {
		Name     string
		Symbol   string
		Chain    string
		Quote    float64
		Pretty   string
		IsNative bool
	}

	var (
		totalValue  float64
		tokens      []tokenHolding
		chainTotals = make(map[string]float64)
		chainNames  = make(map[string]string) // id → display name
	)

	for _, it := range data.Data.Items {
		if it.Quote <= 0 {
			continue
		}

		totalValue += it.Quote

		chainDisplay := it.ChainDisplayName
		if chainDisplay == "" {
			chainDisplay = it.ChainName
		}
		if chainDisplay == "" {
			chainDisplay = it.ChainID
		}
		chainTotals[chainDisplay] += it.Quote
		chainNames[chainDisplay] = chainDisplay

		name := it.ContractName
		if name == "" {
			name = it.ContractTickerSymbol
		}

		tokens = append(tokens, tokenHolding{
			Name:     name,
			Symbol:   it.ContractTickerSymbol,
			Chain:    chainDisplay,
			Quote:    it.Quote,
			Pretty:   it.PrettyQuote,
			IsNative: it.IsNativeToken,
		})
	}

	if totalValue == 0 || len(tokens) == 0 {
		return fmt.Sprintf("No portfolio data found for address %s on the scanned chains.", address), nil
	}

	// Sort chain by value desc
	type chainAgg struct {
		Name  string
		Value float64
	}
	var chains []chainAgg
	for name, val := range chainTotals {
		chains = append(chains, chainAgg{Name: name, Value: val})
	}
	sort.Slice(chains, func(i, j int) bool {
		return chains[i].Value > chains[j].Value
	})

	// Sort tokens by value desc dan ambil top 5
	sort.Slice(tokens, func(i, j int) bool {
		return tokens[i].Quote > tokens[j].Quote
	})
	maxTokens := 5
	if len(tokens) < maxTokens {
		maxTokens = len(tokens)
	}

	var b strings.Builder
	b.WriteString("📊 Portfolio snapshot\n")
	fmt.Fprintf(&b, "Address: `%s`\n", address)
	if len(chains) > 0 {
		b.WriteString("Chains: ")
		for i, c := range chains {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(c.Name)
		}
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "Quote: %s\n\n", strings.ToUpper(quoteCurrency))

	// Total
	fmt.Fprintf(&b, "Total value (approx.): $%.2f\n\n", totalValue)

	// Per chain breakdown
	if len(chains) > 0 {
		b.WriteString("Per chain:\n")
		for _, c := range chains {
			fmt.Fprintf(&b, "• %s — $%.2f\n", c.Name, c.Value)
		}
		b.WriteString("\n")
	}

	// Top tokens
	b.WriteString("Top holdings:\n")
	for i := 0; i < maxTokens; i++ {
		t := tokens[i]
		label := t.Pretty
		if label == "" {
			label = fmt.Sprintf("$%.2f", t.Quote)
		}
		prefix := ""
		if t.IsNative {
			prefix = " (native)"
		}
		fmt.Fprintf(
			&b,
			"%d) %s (%s) on %s%s — %s\n",
			i+1,
			t.Name,
			strings.ToUpper(t.Symbol),
			t.Chain,
			prefix,
			label,
		)
	}

	b.WriteString("\nSource: Covalent allchains balances API")

	return b.String(), nil
}
