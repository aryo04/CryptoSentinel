package clients

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"time"

	"teneo-agent-demo1/internal/models"
)

type WhaleClient struct {
	http *http.Client
}

func NewWhaleClient(httpClient *http.Client) *WhaleClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	return &WhaleClient{
		http: httpClient,
	}
}

// GetLargeTransfers:
// - symbol: btc, eth, bsc
// - chains: sekarang diabaikan, routing berdasarkan symbol
// - limit: max hasil yang dikembalikan
func (c *WhaleClient) GetLargeTransfers(
	ctx context.Context,
	symbol string,
	chains []string,
	limit int,
) ([]models.WhaleTransfer, error) {

	if c.http == nil {
		return nil, errors.New("whale client http not initialized")
	}

	if limit <= 0 {
		limit = 5
	}
	if limit > 50 {
		limit = 50
	}

	s := symbol
	switch s {
	case "btc":
		return c.fetchBitcoinWhales(ctx, limit)
	case "eth":
		return c.fetchEVMWhalesBlockchair(ctx, "ethereum", "ETH", "ethereum", limit)
	case "bsc", "bnb":
		return c.fetchEVMWhalesBlockchair(ctx, "binance-smart-chain", "BNB", "binancecoin", limit)
	default:
		return nil, fmt.Errorf("unsupported symbol for whales: %s (supported: btc, eth, bsc)", s)
	}
}

// ------------------------- BTC (mempool.space + CoinGecko) -------------------------

type cgSimplePriceResp map[string]map[string]float64

func (c *WhaleClient) fetchBTCPriceUSD(ctx context.Context) (float64, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=usd",
		nil,
	)
	if err != nil {
		return 0, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("coingecko simple price status: %d", resp.StatusCode)
	}

	var body cgSimplePriceResp
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return 0, err
	}

	btcData, ok := body["bitcoin"]
	if !ok {
		return 0, errors.New("missing 'bitcoin' field in coingecko response")
	}
	price, ok := btcData["usd"]
	if !ok || price <= 0 {
		return 0, errors.New("invalid btc price from coingecko")
	}
	return price, nil
}

func (c *WhaleClient) fetchBitcoinWhales(ctx context.Context, limit int) ([]models.WhaleTransfer, error) {
	priceUSD, err := c.fetchBTCPriceUSD(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch BTC price: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://mempool.space/api/mempool/recent",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("build mempool request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mempool request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("mempool.space returned status %d", resp.StatusCode)
	}

	var raw []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode mempool response: %w", err)
	}
	if len(raw) == 0 {
		return nil, errors.New("no recent mempool transactions")
	}

	const minUSD = 100_000.0 // threshold whale

	out := make([]models.WhaleTransfer, 0, limit)

	for _, tx := range raw {
		valSats, ok := tx["value"].(float64)
		if !ok {
			continue
		}
		btcAmount := valSats / 1e8
		usd := btcAmount * priceUSD
		if usd < minUSD {
			continue
		}

		txid, _ := tx["txid"].(string)

		var ts time.Time
		if tUnix, ok := tx["time"].(float64); ok {
			sec := int64(tUnix)
			ts = time.Unix(sec, 0).UTC()
		} else {
			ts = time.Now().UTC()
		}

		out = append(out, models.WhaleTransfer{
			TxID:       txid,
			Symbol:     "btc",
			Chain:      "bitcoin",
			From:       "unknown",
			To:         "unknown",
			Amount:     btcAmount,
			AmountUSD:  usd,
			Timestamp:  ts,
			Direction:  "unknown",
			TxURL:      fmt.Sprintf("https://mempool.space/tx/%s", url.PathEscape(txid)),
			RawMessage: "Large unconfirmed BTC transaction from mempool.space",
		})

		if len(out) >= limit {
			break
		}
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("no BTC transactions above %.0f USD threshold", minUSD)
	}
	return out, nil
}

// ------------------------- EVM (ETH / BSC via Blockchair + CG) -------------------------

func (c *WhaleClient) fetchEVMPriceUSD(ctx context.Context, cgID string) (float64, error) {
	urlStr := fmt.Sprintf(
		"https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd",
		url.QueryEscape(cgID),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return 0, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("coingecko simple price status: %d", resp.StatusCode)
	}

	var body map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return 0, err
	}

	tokenData, ok := body[cgID]
	if !ok {
		return 0, fmt.Errorf("missing '%s' in coingecko response", cgID)
	}
	price, ok := tokenData["usd"]
	if !ok || price <= 0 {
		return 0, fmt.Errorf("invalid %s usd price from coingecko", cgID)
	}
	return price, nil
}

func (c *WhaleClient) fetchEVMWhalesBlockchair(
	ctx context.Context,
	network string,   // "ethereum" | "binance-smart-chain"
	symbol string,    // "ETH" | "BNB"
	cgID string,      // "ethereum" | "binancecoin"
	limit int,
) ([]models.WhaleTransfer, error) {

	priceUSD, err := c.fetchEVMPriceUSD(ctx, cgID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s price: %w", symbol, err)
	}

	// min native (ETH/BNB) untuk $100k
	const minUSD = 100_000.0
	minNative := minUSD / priceUSD
	minNative = math.Max(minNative, 0.0001)

	// Blockchair 'value' disimpan dalam satuan "wei-like" (untuk EVM chains),
	// jadi kita konversi jadi integer besar.
	minValueUnit := int64(minNative * 1e18)

	endpoint := fmt.Sprintf(
		"https://api.blockchair.com/%s/transactions?q=value(>%d)&limit=%d",
		network,
		minValueUnit,
		limit*3, // ambil lebih banyak, nanti difilter lagi di sisi kita
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build blockchair request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("blockchair request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("blockchair returned status %d for %s", resp.StatusCode, network)
	}

	// Struktur di Blockchair cukup dinamis, jadi kita pakai map generik
	var body struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode blockchair response: %w", err)
	}
	if len(body.Data) == 0 {
		return nil, errors.New("no transaction data from blockchair")
	}

	out := make([]models.WhaleTransfer, 0, limit)

	for _, item := range body.Data {
		txMap := item

		// Beberapa endpoint pakai nested "transaction"
		if t, ok := item["transaction"].(map[string]interface{}); ok {
			txMap = t
		}

		// value (wei / unit)
		rawVal, ok := txMap["value"].(float64)
		if !ok {
			continue
		}
		native := rawVal / 1e18
		usd := native * priceUSD
		if usd < minUSD {
			continue
		}

		// hash
		hash, _ := txMap["hash"].(string)
		if hash == "" {
			if h2, ok := txMap["transaction_hash"].(string); ok {
				hash = h2
			}
		}

		// time
		var ts time.Time
		if tstr, ok := txMap["time"].(string); ok && tstr != "" {
			// format biasanya "YYYY-MM-DD HH:MM:SS"
			tParsed, err := time.Parse("2006-01-02 15:04:05", tstr)
			if err == nil {
				ts = tParsed.UTC()
			}
		}
		if ts.IsZero() {
			ts = time.Now().UTC()
		}

		fromAddr, _ := txMap["sender"].(string)
		toAddr, _ := txMap["recipient"].(string)

		explorerURL := ""
		switch network {
		case "ethereum":
			explorerURL = fmt.Sprintf("https://etherscan.io/tx/%s", url.PathEscape(hash))
		case "binance-smart-chain":
			explorerURL = fmt.Sprintf("https://bscscan.com/tx/%s", url.PathEscape(hash))
		}

		out = append(out, models.WhaleTransfer{
			TxID:       hash,
			Symbol:     strings.ToLower(symbol),
			Chain:      network,
			From:       fromAddr,
			To:         toAddr,
			Amount:     native,
			AmountUSD:  usd,
			Timestamp:  ts,
			Direction:  "unknown",
			TxURL:      explorerURL,
			RawMessage: fmt.Sprintf("Large %s transaction from Blockchair", symbol),
		})

		if len(out) >= limit {
			break
		}
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("no %s transactions above %.0f USD threshold", symbol, minUSD)
	}

	return out, nil
}
