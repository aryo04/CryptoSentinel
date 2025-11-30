package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"teneo-agent-demo1/internal/clients"
	"teneo-agent-demo1/internal/services"
)

var trendHTTPClient = &http.Client{Timeout: 15 * 1e9} // 15s

type cgMarketChartTrend struct {
	Prices [][]float64 `json:"prices"`
}

// CmdTrend shows short/mid/long-term trend using moving averages.
// Usage: trend [symbol]
// Example: trend btc
func CmdTrend(ctx context.Context, cl *clients.Clients, args []string) (string, error) {
	if len(args) == 0 {
		return "Usage: trend [symbol]\nExample: trend btc", nil
	}
	sym := args[0]

	id, err := cl.ResolveCoinID(ctx, sym)
	if err != nil {
		return fmt.Sprintf("Could not resolve asset '%s': %v", sym, err), nil
	}

	mkt, err := cl.FetchCoinMarket(ctx, id)
	if err != nil {
		return fmt.Sprintf("Failed to fetch market data for '%s': %v", sym, err), nil
	}

	// Fetch ~90d market chart
	endpoint := fmt.Sprintf(
		"https://api.coingecko.com/api/v3/coins/%s/market_chart?vs_currency=usd&days=90",
		url.PathEscape(id),
	)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")

	resp, err := trendHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("CoinGecko market_chart error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("CoinGecko market_chart status: %d", resp.StatusCode)
	}

	var data cgMarketChartTrend
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("decode market_chart error: %v", err)
	}
	if len(data.Prices) == 0 {
		return "No historical price data available for this asset.", nil
	}

	// Extract prices
	prices := make([]float64, 0, len(data.Prices))
	for _, p := range data.Prices {
		if len(p) != 2 {
			continue
		}
		prices = append(prices, p[1])
	}
	if len(prices) == 0 {
		return "No valid price points returned by CoinGecko.", nil
	}

	n := len(prices)
	if n < 10 {
		return "Not enough historical data to compute trend (need more price points).", nil
	}

	// Helper: trailing average over last k points
	trailingAvg := func(k int) float64 {
		if k <= 0 {
			k = 1
		}
		if k > n {
			k = n
		}
		sum := 0.0
		for i := n - k; i < n; i++ {
			sum += prices[i]
		}
		return sum / float64(k)
	}

	// Approximate window sizes based on 90d span
	w7 := n * 7 / 90
	w30 := n * 30 / 90
	w90 := n // full history

	ma7 := trailingAvg(w7)
	ma30 := trailingAvg(w30)
	ma90 := trailingAvg(w90)

	// Trend classification
	trend := "Sideways / Mixed"
	interpret := "Short, mid, and long-term averages are not clearly aligned. Price may be ranging or transitioning."

	if ma7 > ma30 && ma30 > ma90 {
		trend = "Bullish 📈"
		interpret = "Short-term price is above mid and long-term averages, indicating an ongoing uptrend."
	} else if ma7 < ma30 && ma30 < ma90 {
		trend = "Bearish 📉"
		interpret = "Short-term price is below mid and long-term averages, indicating a sustained downtrend."
	}

	out := fmt.Sprintf(
		"Trend — %s (%s)\n"+
			"Current price: $%s\n\n"+
			"7d MA (short-term):  $%s\n"+
			"30d MA (mid-term):   $%s\n"+
			"90d MA (long-term):  $%s\n\n"+
			"Trend bias: %s\n"+
			"Interpretation: %s\n"+
			"Source: CoinGecko (90d market chart)",
		mkt.Name, strings.ToUpper(mkt.Symbol),
		services.F(mkt.CurrentPrice),
		services.F(ma7),
		services.F(ma30),
		services.F(ma90),
		trend,
		interpret,
	)

	return out, nil
}
