package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"teneo-agent-demo1/internal/clients"
	"teneo-agent-demo1/internal/services"
)

// CmdGainers menampilkan top daily gainers (24h) di antara top 100 coins by market cap (CoinGecko)
// Usage: gainers [limit]
// Contoh:
//   gainers
//   gainers 15
func CmdGainers(ctx context.Context, cl *clients.Clients, args []string) (string, error) {
	limit := 10
	if len(args) > 0 {
		if n, err := strconv.Atoi(args[0]); err == nil && n > 0 && n <= 50 {
			limit = n
		}
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	// Ambil 100 coin teratas by market cap, lalu sort sendiri berdasarkan 24h change
	endpoint := "https://api.coingecko.com/api/v3/coins/markets" +
		"?vs_currency=usd" +
		"&order=market_cap_desc" +
		"&per_page=100" +
		"&page=1" +
		"&sparkline=false" +
		"&price_change_percentage=24h"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("build request error: %v", err)
	}
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
	if len(arr) == 0 {
		return "No data for gainers.", nil
	}

	// sort descending by 24h change
	sort.Slice(arr, func(i, j int) bool {
		return arr[i].Change > arr[j].Change
	})

	if len(arr) < limit {
		limit = len(arr)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Top %d daily gainers (among top 100 by market cap):\n", limit))
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
