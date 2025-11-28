package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"teneo-agent-demo1/internal/clients"
	"teneo-agent-demo1/internal/services"
)

// CmdLosers menampilkan top daily losers (di antara top 100 market cap)
// Usage: losers [limit]
// Example: losers 10
func CmdLosers(ctx context.Context, cl *clients.Clients, args []string) (string, error) {
	limit := 10
	if len(args) > 0 {
		if n, err := strconv.Atoi(args[0]); err == nil && n > 0 && n <= 50 {
			limit = n
		}
	}

	const endpoint = "https://api.coingecko.com/api/v3/coins/markets" +
		"?vs_currency=usd&order=market_cap_desc&per_page=100&page=1" +
		"&sparkline=false&price_change_percentage=24h"

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")

	resp, err := cl.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("coingecko request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("coingecko status: %d", resp.StatusCode)
	}

	type coin struct {
		Name   string  `json:"name"`
		Symbol string  `json:"symbol"`
		Price  float64 `json:"current_price"`
		Mcap   float64 `json:"market_cap"`
		Change float64 `json:"price_change_percentage_24h"`
	}

	var arr []coin
	if err := json.NewDecoder(resp.Body).Decode(&arr); err != nil {
		return "", fmt.Errorf("decode error: %v", err)
	}

	// sort dari yang paling turun
	sort.Slice(arr, func(i, j int) bool {
		return arr[i].Change < arr[j].Change
	})

	if len(arr) < limit {
		limit = len(arr)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Top %d daily losers (among top 100 by market cap):\n", limit))
	for i := 0; i < limit; i++ {
		c := arr[i]
		b.WriteString(fmt.Sprintf(
			"%d) %s (%s) — $%s | 24h: %s%% | Mcap: $%s\n",
			i+1,
			c.Name,
			strings.ToUpper(c.Symbol),
			services.F(c.Price),
			services.F(c.Change),
			services.IntComma(int64(c.Mcap)),
		))
	}

	return b.String(), nil
}
