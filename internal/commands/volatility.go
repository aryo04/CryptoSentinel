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

var volHTTPClient = &http.Client{Timeout: 15 * 1e9} // 15s

// simple struct for CoinGecko market_chart
type cgMarketChart struct {
	Prices [][]float64 `json:"prices"` // [timestamp, price]
}

// CmdVolatility shows recent price volatility for an asset.
// Usage: volatility [symbol]
// Example: volatility eth
func CmdVolatility(ctx context.Context, cl *clients.Clients, args []string) (string, error) {
	if len(args) == 0 {
		return "Usage: volatility [symbol]\nExample: volatility eth", nil
	}
	sym := args[0]

	id, err := cl.ResolveCoinID(ctx, sym)
	if err != nil {
		return fmt.Sprintf("Could not resolve asset '%s': %v", sym, err), nil
	}

	// Current market snapshot (for price)
	mkt, err := cl.FetchCoinMarket(ctx, id)
	if err != nil {
		return fmt.Sprintf("Failed to fetch market data for '%s': %v", sym, err), nil
	}

	// Fetch 7d market chart
	endpoint := fmt.Sprintf(
		"https://api.coingecko.com/api/v3/coins/%s/market_chart?vs_currency=usd&days=7",
		url.PathEscape(id),
	)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")

	resp, err := volHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("CoinGecko market_chart error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("CoinGecko market_chart status: %d", resp.StatusCode)
	}

	var data cgMarketChart
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("decode market_chart error: %v", err)
	}
	if len(data.Prices) == 0 {
		return "No volatility data available for this asset (empty market_chart).", nil
	}

	// Extract prices only
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

	// Compute simple 7d volatility as high/low range percentage
	minPrice, maxPrice := prices[0], prices[0]
	for _, v := range prices[1:] {
		if v < minPrice {
			minPrice = v
		}
		if v > maxPrice {
			maxPrice = v
		}
	}

	rangePct := 0.0
	if minPrice > 0 {
		rangePct = (maxPrice - minPrice) / minPrice * 100.0
	}

	// Classify volatility
	class := "Low"
	switch {
	case rangePct > 40:
		class = "Extreme"
	case rangePct > 25:
		class = "High"
	case rangePct > 10:
		class = "Medium"
	}

	// Build output
	out := fmt.Sprintf(
		"Volatility — %s (%s)\n"+
			"Current price: $%s\n\n"+
			"7d price range: $%s → $%s\n"+
			"7d volatility (range): %s%%\n\n"+
			"Volatility class: %s\n"+
			"Source: CoinGecko (7d market chart)",
		mkt.Name, strings.ToUpper(mkt.Symbol),
		services.F(mkt.CurrentPrice),
		services.F(minPrice),
		services.F(maxPrice),
		services.F(rangePct),
		class,
	)

	return out, nil
}
