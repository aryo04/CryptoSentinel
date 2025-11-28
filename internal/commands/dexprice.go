package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var dexHTTPClient = &http.Client{Timeout: 15 * time.Second}

type dexScreenerResponse struct {
	Pairs []struct {
		ChainId string `json:"chainId"`
		DexId   string `json:"dexId"`
		URL     string `json:"url"`

		BaseToken struct {
			Address string `json:"address"`
			Symbol  string `json:"symbol"`
			Name    string `json:"name"`
		} `json:"baseToken"`
		QuoteToken struct {
			Address string `json:"address"`
			Symbol  string `json:"symbol"`
			Name    string `json:"name"`
		} `json:"quoteToken"`

		PriceUsd string `json:"priceUsd"`

		Liquidity struct {
			Usd float64 `json:"usd"`
		} `json:"liquidity"`

		Volume struct {
			H24 float64 `json:"h24"`
		} `json:"volume"`
	} `json:"pairs"`
}

func CmdDexPrice(ctx context.Context, args []string) (string, error) {
	if len(args) < 2 {
		return "Usage: dexprice [token] [chain]\nExample: dexprice pepe ethereum", nil
	}
	token := strings.ToLower(args[0])
	chain := strings.ToLower(args[1])

	query := token + " " + chain
	endpoint := "https://api.dexscreener.com/latest/dex/search?q=" + url.QueryEscape(query)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")

	resp, err := dexHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("dexscreener error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("dexscreener status: %d", resp.StatusCode)
	}

	var data dexScreenerResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("decode dexscreener error: %w", err)
	}
	if len(data.Pairs) == 0 {
		return fmt.Sprintf("No DEX pairs found for %s on %s", token, chain), nil
	}

	// Pilih pair yang chain-nya match dulu
	var best *dexScreenerResponse
	var pairIndex int
	for i, p := range data.Pairs {
		if strings.EqualFold(p.ChainId, chain) || strings.Contains(strings.ToLower(p.ChainId), chain) {
			best = &data
			pairIndex = i
			break
		}
	}
	if best == nil {
		best = &data
		pairIndex = 0
	}

	p := best.Pairs[pairIndex]
	price := p.PriceUsd
	if price == "" {
		price = "N/A"
	}

	out := fmt.Sprintf(
		"🧬 DEX price — %s on %s\n"+
			"Pair: %s / %s (%s)\n"+
			"Price: $%s\n"+
			"Liquidity: $%.0f\n"+
			"Volume 24h: $%.0f\n"+
			"DEX: %s\n"+
			"Link: %s",
		strings.ToUpper(p.BaseToken.Symbol),
		p.ChainId,
		p.BaseToken.Symbol,
		p.QuoteToken.Symbol,
		p.QuoteToken.Name,
		price,
		p.Liquidity.Usd,
		p.Volume.H24,
		p.DexId,
		p.URL,
	)

	return out, nil
}
