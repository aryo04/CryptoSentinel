package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"teneo-agent-demo1/internal/models"
)

const cgBase = "https://api.coingecko.com/api/v3"

func (c *Clients) ResolveCoinID(ctx context.Context, symbolOrName string) (string, error) {
	query := strings.TrimSpace(symbolOrName)
	if query == "" {
		return "", fmt.Errorf("empty symbol")
	}
	endpoint := fmt.Sprintf("%s/search?query=%s", cgBase, url.QueryEscape(query))

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("coingecko search status: %d", resp.StatusCode)
	}

	var sr struct {
		Coins []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Symbol string `json:"symbol"`
		} `json:"coins"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return "", err
	}
	if len(sr.Coins) == 0 {
		return "", fmt.Errorf("no match on CoinGecko")
	}

	q := strings.ToLower(query)
	for _, coin := range sr.Coins {
		if strings.ToLower(coin.Symbol) == q {
			return coin.ID, nil
		}
	}
	return sr.Coins[0].ID, nil
}

func (c *Clients) FetchCoinMarket(ctx context.Context, id string) (*models.CoinMarket, error) {
	endpoint := fmt.Sprintf(
		"%s/coins/markets?vs_currency=usd&ids=%s&order=market_cap_desc&per_page=1&page=1&sparkline=false&price_change_percentage=24h,7d",
		cgBase, url.PathEscape(id),
	)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("coingecko status: %d", resp.StatusCode)
	}

	var arr []struct {
		ID                         string  `json:"id"`
		Symbol                     string  `json:"symbol"`
		Name                       string  `json:"name"`
		CurrentPrice               float64 `json:"current_price"`
		MarketCap                  float64 `json:"market_cap"`
		TotalVolume                float64 `json:"total_volume"`
		PriceChangePercentage24h   float64 `json:"price_change_percentage_24h"`
		PriceChangePercentage7dInC float64 `json:"price_change_percentage_7d_in_currency"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&arr); err != nil {
		return nil, err
	}
	if len(arr) == 0 {
		return nil, fmt.Errorf("no market data")
	}

	a := arr[0]
	return &models.CoinMarket{
		ID:                a.ID,
		Symbol:            a.Symbol,
		Name:              a.Name,
		CurrentPrice:      a.CurrentPrice,
		MarketCap:         a.MarketCap,
		TotalVolume:       a.TotalVolume,
		PriceChangePct24h: a.PriceChangePercentage24h,
		PriceChangePct7d:  a.PriceChangePercentage7dInC,
	}, nil
}

func (c *Clients) FetchTopMarkets(ctx context.Context, limit int) ([]models.CoinMarket, error) {
	endpoint := fmt.Sprintf(
		"%s/coins/markets?vs_currency=usd&order=market_cap_desc&per_page=%d&page=1&sparkline=false&price_change_percentage=24h,7d",
		cgBase, limit,
	)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("coingecko status: %d", resp.StatusCode)
	}

	var arr []struct {
		ID                         string  `json:"id"`
		Symbol                     string  `json:"symbol"`
		Name                       string  `json:"name"`
		CurrentPrice               float64 `json:"current_price"`
		MarketCap                  float64 `json:"market_cap"`
		TotalVolume                float64 `json:"total_volume"`
		PriceChangePercentage24h   float64 `json:"price_change_percentage_24h"`
		PriceChangePercentage7dInC float64 `json:"price_change_percentage_7d_in_currency"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&arr); err != nil {
		return nil, err
	}
	res := make([]models.CoinMarket, 0, len(arr))
	for _, a := range arr {
		res = append(res, models.CoinMarket{
			ID:                a.ID,
			Symbol:            a.Symbol,
			Name:              a.Name,
			CurrentPrice:      a.CurrentPrice,
			MarketCap:         a.MarketCap,
			TotalVolume:       a.TotalVolume,
			PriceChangePct24h: a.PriceChangePercentage24h,
			PriceChangePct7d:  a.PriceChangePercentage7dInC,
		})
	}
	return res, nil
}

func (c *Clients) FetchConversionRate(ctx context.Context, fromID, toID string) (float64, error) {
	// pakai simple: ambil harga USD masing2 lalu rate = fromUSD / toUSD
	from, err := c.FetchCoinMarket(ctx, fromID)
	if err != nil {
		return 0, err
	}
	to, err := c.FetchCoinMarket(ctx, toID)
	if err != nil {
		return 0, err
	}
	if to.CurrentPrice == 0 {
		return 0, fmt.Errorf("zero price for %s", toID)
	}
	return from.CurrentPrice / to.CurrentPrice, nil
}
