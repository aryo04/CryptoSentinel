package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"teneo-agent-demo1/internal/models"
)

func SupportedCexList() []string {
	return []string{"binance", "okx", "bybit", "kucoin", "bitget"}
}

func IsSupportedCex(name string) bool {
	n := strings.ToLower(name)
	for _, c := range SupportedCexList() {
		if c == n {
			return true
		}
	}
	return false
}

func (c *Clients) FetchCexGainers(ctx context.Context, cex string, limit int) ([]models.CexGainer, error) {
	switch strings.ToLower(cex) {
	case "binance":
		return c.fetchBinanceGainers(ctx, limit)
	case "okx":
		return c.fetchOkxGainers(ctx, limit)
	case "bybit":
		return c.fetchBybitGainers(ctx, limit)
	case "kucoin":
		return c.fetchKucoinGainers(ctx, limit)
	case "bitget":
		return c.fetchBitgetGainers(ctx, limit)
	default:
		return nil, fmt.Errorf("unsupported cex: %s", cex)
	}
}

// -------- Binance ----------
func (c *Clients) fetchBinanceGainers(ctx context.Context, limit int) ([]models.CexGainer, error) {
	endpoint := "https://api.binance.com/api/v3/ticker/24hr"

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var arr []struct {
		Symbol             string `json:"symbol"`
		LastPrice          string `json:"lastPrice"`
		PriceChangePercent string `json:"priceChangePercent"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&arr); err != nil {
		return nil, err
	}

	var list []models.CexGainer
	for _, t := range arr {
		// fokus USDT pairs
		if !strings.HasSuffix(t.Symbol, "USDT") {
			continue
		}
		price, _ := strconv.ParseFloat(t.LastPrice, 64)
		chg, _ := strconv.ParseFloat(t.PriceChangePercent, 64)
		list = append(list, models.CexGainer{
			Exchange:  "binance",
			Symbol:    t.Symbol,
			Price:     price,
			ChangePct: chg,
		})
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].ChangePct > list[j].ChangePct
	})

	if len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}

// -------- OKX ----------
func (c *Clients) fetchOkxGainers(ctx context.Context, limit int) ([]models.CexGainer, error) {
	endpoint := "https://www.okx.com/api/v5/market/tickers?instType=SPOT"

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		Code string `json:"code"`
		Data []struct {
			InstID string `json:"instId"`
			Last   string `json:"last"`
			Open24 string `json:"open24h"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	var list []models.CexGainer
	for _, t := range data.Data {
		if !strings.HasSuffix(t.InstID, "-USDT") {
			continue
		}
		last, err1 := strconv.ParseFloat(t.Last, 64)
		open, err2 := strconv.ParseFloat(t.Open24, 64)
		if err1 != nil || err2 != nil || open <= 0 {
			continue
		}
		chg := (last - open) / open * 100
		list = append(list, models.CexGainer{
			Exchange:  "okx",
			Symbol:    t.InstID,
			Price:     last,
			ChangePct: chg,
		})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].ChangePct > list[j].ChangePct
	})
	if len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}

// -------- Bybit ----------
func (c *Clients) fetchBybitGainers(ctx context.Context, limit int) ([]models.CexGainer, error) {
	endpoint := "https://api.bybit.com/v5/market/tickers?category=spot"

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		Result struct {
			List []struct {
				Symbol string `json:"symbol"`
				Last   string `json:"lastPrice"`
				High   string `json:"highPrice24h"`
				Low    string `json:"lowPrice24h"`
				Prev   string `json:"prevPrice24h"`
			} `json:"list"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var list []models.CexGainer
	for _, t := range data.Result.List {
		if !strings.HasSuffix(t.Symbol, "USDT") {
			continue
		}
		last, err1 := strconv.ParseFloat(t.Last, 64)
		prev, err2 := strconv.ParseFloat(t.Prev, 64)
		if err1 != nil || err2 != nil || prev <= 0 {
			continue
		}
		chg := (last - prev) / prev * 100
		list = append(list, models.CexGainer{
			Exchange:  "bybit",
			Symbol:    t.Symbol,
			Price:     last,
			ChangePct: chg,
		})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].ChangePct > list[j].ChangePct
	})
	if len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}

// -------- KuCoin ----------
func (c *Clients) fetchKucoinGainers(ctx context.Context, limit int) ([]models.CexGainer, error) {
	endpoint := "https://api.kucoin.com/api/v1/market/allTickers"

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		Data struct {
			Tickers []struct {
				Symbol   string `json:"symbol"`
				Last     string `json:"last"`
				ChangePt string `json:"changeRate"`
			} `json:"ticker"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var list []models.CexGainer
	for _, t := range data.Data.Tickers {
		if !strings.Contains(t.Symbol, "-USDT") {
			continue
		}
		last, _ := strconv.ParseFloat(t.Last, 64)
		changeRate, _ := strconv.ParseFloat(t.ChangePt, 64) // ex: 0.12
		chgPct := changeRate * 100
		list = append(list, models.CexGainer{
			Exchange:  "kucoin",
			Symbol:    t.Symbol,
			Price:     last,
			ChangePct: chgPct,
		})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].ChangePct > list[j].ChangePct
	})
	if len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}

// -------- Bitget ----------
func (c *Clients) fetchBitgetGainers(ctx context.Context, limit int) ([]models.CexGainer, error) {
	endpoint := "https://api.bitget.com/api/v2/spot/market/tickers"

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			Symbol string `json:"symbol"`
			Open   string `json:"open"`
			LastPr string `json:"lastPr"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	if data.Code != "00000" {
		return nil, fmt.Errorf("bitget error: %s %s", data.Code, data.Msg)
	}

	var list []models.CexGainer
	for _, t := range data.Data {
		if !strings.HasSuffix(t.Symbol, "USDT") {
			continue
		}
		open, err1 := strconv.ParseFloat(t.Open, 64)
		last, err2 := strconv.ParseFloat(t.LastPr, 64)
		if err1 != nil || err2 != nil || open <= 0 || math.IsNaN(open) {
			continue
		}
		chg := (last - open) / open * 100
		list = append(list, models.CexGainer{
			Exchange:  "bitget",
			Symbol:    t.Symbol,
			Price:     last,
			ChangePct: chg,
		})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].ChangePct > list[j].ChangePct
	})
	if len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}
