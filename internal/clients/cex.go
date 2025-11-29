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

// CEX yang dipakai agent sekarang:
// - coinbase
// - binance
// - okx
// - bybit
// - mexc
// - kucoin
// - bitget
func SupportedCexList() []string {
	return []string{"coinbase", "binance", "okx", "bybit", "mexc", "kucoin", "bitget"}
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
	case "coinbase":
		return c.fetchCoinbaseGainers(ctx, limit)
	case "binance":
		return c.fetchBinanceGainers(ctx, limit)
	case "okx":
		return c.fetchOkxGainers(ctx, limit)
	case "bybit":
		return c.fetchBybitGainers(ctx, limit)
	case "mexc":
		return c.fetchMexcGainers(ctx, limit)
	case "kucoin":
		return c.fetchKucoinGainers(ctx, limit)
	case "bitget":
		return c.fetchBitgetGainers(ctx, limit)
	default:
		return nil, fmt.Errorf("unsupported cex: %s", cex)
	}
}

//
// -------- Coinbase (Exchange, public market data) ----------
//
// NOTE:
// - Kita tidak pakai Advanced Trade / brokerage public products
//   karena butuh Authorization Bearer (JWT).
// - Sebagai gantinya, pakai Coinbase Exchange Market Data API:
//   - GET https://api.exchange.coinbase.com/products/{product_id}/stats
//   - Respons berisi open, last, volume, dll (24 jam).
// - Untuk menghindari terlalu banyak request, kita batasi
//   ke beberapa spot pair besar USD saja.
//
func (c *Clients) fetchCoinbaseGainers(ctx context.Context, limit int) ([]models.CexGainer, error) {
	// Daftar pair besar di Coinbase (bisa ditambah kalau mau)
	productIDs := []string{
		"BTC-USD",
		"ETH-USD",
		"SOL-USD",
		"XRP-USD",
		"DOGE-USD",
		"ADA-USD",
		"AVAX-USD",
		"LINK-USD",
		"MATIC-USD",
		"LTC-USD",
		"BCH-USD",
		"TON-USD",
	}

	type cbStats struct {
		Open  string `json:"open"`
		Last  string `json:"last"`
		High  string `json:"high"`
		Low   string `json:"low"`
		Vol   string `json:"volume"`
		Vol30 string `json:"volume_30day"`
	}

	var list []models.CexGainer

	for _, pid := range productIDs {
		endpoint := "https://api.exchange.coinbase.com/products/" + pid + "/stats"

		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		resp, err := c.HTTP.Do(req)
		if err != nil {
			// kalau 1 pair error, skip aja, jangan gagal total
			continue
		}
		func() {
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				// skip kalau status bukan 200
				return
			}

			var s cbStats
			if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
				return
			}

			open, err1 := strconv.ParseFloat(s.Open, 64)
			last, err2 := strconv.ParseFloat(s.Last, 64)
			if err1 != nil || err2 != nil || open <= 0 {
				return
			}
			chg := (last - open) / open * 100.0

			list = append(list, models.CexGainer{
				Exchange:  "coinbase",
				Symbol:    pid,
				Price:     last,
				ChangePct: chg,
			})
		}()
	}

	if len(list) == 0 {
		return nil, fmt.Errorf("no coinbase data")
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].ChangePct > list[j].ChangePct
	})
	if len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}

//
// -------- Binance ----------
//
// Endpoint: GET https://api.binance.com/api/v3/ticker/24hr
// Mengembalikan list semua simbol dengan lastPrice & priceChangePercent.
//
func (c *Clients) fetchBinanceGainers(ctx context.Context, limit int) ([]models.CexGainer, error) {
	endpoint := "https://api.binance.com/api/v3/ticker/24hr"

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance status: %d", resp.StatusCode)
	}

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
		price, err1 := strconv.ParseFloat(t.LastPrice, 64)
		chg, err2 := strconv.ParseFloat(t.PriceChangePercent, 64)
		if err1 != nil || err2 != nil {
			continue
		}
		list = append(list, models.CexGainer{
			Exchange:  "binance",
			Symbol:    t.Symbol,
			Price:     price,
			ChangePct: chg,
		})
	}

	if len(list) == 0 {
		return nil, fmt.Errorf("no binance data")
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].ChangePct > list[j].ChangePct
	})
	if len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}

//
// -------- OKX (tetap seperti sebelumnya) ----------
//
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

//
// -------- Bybit ----------
//
// Endpoint: GET https://api.bybit.com/v5/market/tickers?category=spot
// Field utama:
//   - symbol
//   - lastPrice
//   - price24hPcnt  (rasio, misal 0.1234 = 12.34%)
//
func (c *Clients) fetchBybitGainers(ctx context.Context, limit int) ([]models.CexGainer, error) {
	endpoint := "https://api.bybit.com/v5/market/tickers?category=spot"

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bybit status: %d", resp.StatusCode)
	}

	var data struct {
		Result struct {
			List []struct {
				Symbol       string `json:"symbol"`
				LastPrice    string `json:"lastPrice"`
				Price24hPcnt string `json:"price24hPcnt"`
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
		last, err1 := strconv.ParseFloat(t.LastPrice, 64)
		pcnt, err2 := strconv.ParseFloat(t.Price24hPcnt, 64)
		if err1 != nil || err2 != nil {
			continue
		}
		// Bybit kirim rasio (0.1234); kita ubah ke %.
		chg := pcnt * 100

		list = append(list, models.CexGainer{
			Exchange:  "bybit",
			Symbol:    t.Symbol,
			Price:     last,
			ChangePct: chg,
		})
	}

	if len(list) == 0 {
		return nil, fmt.Errorf("no bybit data")
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].ChangePct > list[j].ChangePct
	})
	if len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}

//
// -------- MEXC ----------
//
// Menggunakan Spot v3 Market Data API:
//   GET https://api.mexc.com/api/v3/ticker/24hr
//   - Mengembalikan list semua simbol dengan priceChangePercent & lastPrice.
//
func (c *Clients) fetchMexcGainers(ctx context.Context, limit int) ([]models.CexGainer, error) {
	endpoint := "https://api.mexc.com/api/v3/ticker/24hr"

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
		price, err1 := strconv.ParseFloat(t.LastPrice, 64)
		chg, err2 := strconv.ParseFloat(t.PriceChangePercent, 64)
		if err1 != nil || err2 != nil {
			continue
		}
		list = append(list, models.CexGainer{
			Exchange:  "mexc",
			Symbol:    t.Symbol,
			Price:     price,
			ChangePct: chg,
		})
	}

	if len(list) == 0 {
		return nil, fmt.Errorf("no mexc data")
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].ChangePct > list[j].ChangePct
	})
	if len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}

//
// -------- KuCoin (tanpa perubahan) ----------
//
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

//
// -------- Bitget (tanpa perubahan) ----------
//
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
