package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

var dexHTTPClient = &http.Client{Timeout: 15 * time.Second}

// ====== Struct untuk /latest/dex/search ======

type dexScreenerPair struct {
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
}

type dexScreenerSearchResponse struct {
	SchemaVersion string            `json:"schemaVersion"`
	Pairs         []dexScreenerPair `json:"pairs"`
}

// ====== Hardcoded meme-list per chain (bisa kamu edit sendiri) ======

var memeTokensByChain = map[string][]string{
	"ethereum": {
		"pepe",
		"mog",
		"shib",
		"floki",
		"bonkler", // ganti sesuai selera
	},
	"solana": {
		"bonk",
		"wif",
		"popcat",
		"wen",
		"micchi", // placeholder, silakan ganti
	},
	"bsc": {
		"babydoge",
		"pepe",
		"floki",
		"meme",
		"doge",
	},
	"base": {
		"degen",
		"brian",
		"toshi",
		"pepe",
		"mog",
	},
}

// untuk error message yang rapi
func supportedMemeChains() string {
	chains := make([]string, 0, len(memeTokensByChain))
	for c := range memeTokensByChain {
		chains = append(chains, c)
	}
	sort.Strings(chains)
	return strings.Join(chains, ", ")
}

// ====== Helper umum: cari pair terbaik via search ======

func searchBestDexPair(ctx context.Context, token, chain string) (*dexScreenerPair, error) {
	token = strings.ToLower(token)
	chain = strings.ToLower(chain)

	query := token + " " + chain
	endpoint := "https://api.dexscreener.com/latest/dex/search?q=" + url.QueryEscape(query)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")

	resp, err := dexHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("dexscreener error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dexscreener status: %d", resp.StatusCode)
	}

	var data dexScreenerSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode dexscreener error: %w", err)
	}
	if len(data.Pairs) == 0 {
		return nil, nil // artinya "not found"
	}

	// Pilih pair yang paling relevan (chain match dulu)
	var best *dexScreenerPair
	for i := range data.Pairs {
		p := &data.Pairs[i]
		if strings.EqualFold(p.ChainId, chain) ||
			strings.Contains(strings.ToLower(p.ChainId), chain) {
			best = p
			break
		}
	}
	if best == nil {
		// fallback: pakai pair pertama
		best = &data.Pairs[0]
	}
	return best, nil
}

// helper format harga/angka
func formatUsdFloat(v float64) string {
	if v == 0 {
		return "0"
	}
	// untuk liq/volume, bulatkan saja ke 0 decimal
	return fmt.Sprintf("%,.0f", v)
}

func formatPriceUsdStr(s string) string {
	if s == "" {
		return "N/A"
	}
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		// kalau sangat kecil, pakai lebih banyak decimal
		if v < 0.000001 {
			return fmt.Sprintf("%.10f", v)
		}
		return fmt.Sprintf("%.6f", v)
	}
	return s
}

// ====== Command lama: dexprice [token] [chain] ======

func CmdDexPrice(ctx context.Context, args []string) (string, error) {
	if len(args) < 2 {
		return "Usage: dexprice [token] [chain]\nExample: dexprice pepe ethereum", nil
	}
	token := args[0]
	chain := args[1]

	p, err := searchBestDexPair(ctx, token, chain)
	if err != nil {
		return "", err
	}
	if p == nil {
		return fmt.Sprintf("No DEX pairs found for %s on %s", token, chain), nil
	}

	priceStr := formatPriceUsdStr(p.PriceUsd)

	out := fmt.Sprintf(
		"🧬 DEX price — %s on %s\n"+
			"Pair: %s / %s (%s)\n"+
			"Price: $%s\n"+
			"Liquidity: $%s\n"+
			"Volume 24h: $%s\n"+
			"DEX: %s\n"+
			"Link: %s",
		strings.ToUpper(p.BaseToken.Symbol),
		p.ChainId,
		p.BaseToken.Symbol,
		p.QuoteToken.Symbol,
		p.QuoteToken.Name,
		priceStr,
		formatUsdFloat(p.Liquidity.Usd),
		formatUsdFloat(p.Volume.H24),
		p.DexId,
		p.URL,
	)

	return out, nil
}

// ====== Command baru: dexmeme [chain] ======
// Menampilkan 5 meme-token populer per chain (hardcoded list),
// lalu data realtime-nya diambil dari DexScreener.

func CmdDexMeme(ctx context.Context, args []string) (string, error) {
	chain := "ethereum"
	if len(args) >= 1 && args[0] != "" {
		chain = strings.ToLower(args[0])
	}

	memeList, ok := memeTokensByChain[chain]
	if !ok {
		return fmt.Sprintf(
			"Unknown chain: %s\nSupported chains: %s",
			chain,
			supportedMemeChains(),
		), nil
	}

	builder := &strings.Builder{}
	fmt.Fprintf(builder, "🐸 Top meme tokens on %s (DexScreener)\n\n", chain)

	for i, symbol := range memeList {
		p, err := searchBestDexPair(ctx, symbol, chain)
		if err != nil {
			fmt.Fprintf(builder, "%d) %s — error: %v\n\n", i+1, strings.ToUpper(symbol), err)
			continue
		}
		if p == nil {
			fmt.Fprintf(builder, "%d) %s — pair not found on %s\n\n", i+1, strings.ToUpper(symbol), chain)
			continue
		}

		priceStr := formatPriceUsdStr(p.PriceUsd)

		fmt.Fprintf(builder,
			"%d) %s (%s)\n"+
				"   Pair: %s / %s (%s)\n"+
				"   Price: $%s\n"+
				"   Liquidity: $%s\n"+
				"   Volume 24h: $%s\n"+
				"   DEX: %s\n"+
				"   Link: %s\n\n",
			i+1,
			strings.ToUpper(p.BaseToken.Symbol),
			p.ChainId,
			p.BaseToken.Symbol,
			p.QuoteToken.Symbol,
			p.QuoteToken.Name,
			priceStr,
			formatUsdFloat(p.Liquidity.Usd),
			formatUsdFloat(p.Volume.H24),
			p.DexId,
			p.URL,
		)
	}

	return builder.String(), nil
}
