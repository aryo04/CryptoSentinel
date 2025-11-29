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

// HTTP client untuk DexScreener
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

// ====== Chain alias & normalisasi nama chain ======

// mapping alias → canonical name
var dexChainAliases = map[string]string{
	// Ethereum
	"eth":      "ethereum",
	"ethereum": "ethereum",

	// Solana
	"sol":     "solana",
	"solana":  "solana",

	// BSC
	"bsc": "bsc",
	"bnb": "bsc",

	// Base
	"base": "base",

	// Arbitrum
	"arb":      "arbitrum",
	"arbitrum": "arbitrum",

	// Avalanche
	"avax":      "avalanche",
	"avalanche": "avalanche",
}

func normalizeChainName(chain string) string {
	c := strings.ToLower(strings.TrimSpace(chain))
	if v, ok := dexChainAliases[c]; ok {
		return v
	}
	return c
}

func supportedDexChains() string {
	// ambil semua nilai unik dari alias map
	set := make(map[string]struct{})
	for _, v := range dexChainAliases {
		set[v] = struct{}{}
	}

	chains := make([]string, 0, len(set))
	for c := range set {
		chains = append(chains, c)
	}
	sort.Strings(chains)
	return strings.Join(chains, ", ")
}

// ====== Helper HTTP ke DexScreener ======

func searchDexPairs(ctx context.Context, query string) ([]dexScreenerPair, error) {
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

	return data.Pairs, nil
}

// cari pair terbaik untuk 1 token di 1 chain
func searchBestDexPair(ctx context.Context, token, chain string) (*dexScreenerPair, error) {
	token = strings.ToLower(strings.TrimSpace(token))
	chain = normalizeChainName(chain)

	query := token + " " + chain
	pairs, err := searchDexPairs(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(pairs) == 0 {
		return nil, nil
	}

	// Filter by chain
	var candidates []*dexScreenerPair
	for i := range pairs {
		p := &pairs[i]
		cid := strings.ToLower(p.ChainId)
		if cid == chain || strings.Contains(cid, chain) {
			candidates = append(candidates, p)
		}
	}

	if len(candidates) == 0 {
		// fallback: pakai semua pair
		for i := range pairs {
			candidates = append(candidates, &pairs[i])
		}
	}

	// pilih yang liquidity-nya terbesar
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Liquidity.Usd > candidates[j].Liquidity.Usd
	})

	return candidates[0], nil
}

// filter berdasarkan chain
func filterPairsByChain(pairs []dexScreenerPair, chain string) []dexScreenerPair {
	chain = normalizeChainName(chain)
	out := make([]dexScreenerPair, 0, len(pairs))
	for _, p := range pairs {
		cid := strings.ToLower(p.ChainId)
		if cid == chain || strings.Contains(cid, chain) {
			out = append(out, p)
		}
	}
	return out
}

// deduplicate by base token address
func dedupPairsByBaseToken(pairs []dexScreenerPair) []dexScreenerPair {
	seen := make(map[string]struct{})
	out := make([]dexScreenerPair, 0, len(pairs))
	for _, p := range pairs {
		addr := strings.ToLower(strings.TrimSpace(p.BaseToken.Address))
		if addr == "" {
			// kalau nggak ada address, pakai symbol sebagai fallback
			addr = "sym:" + strings.ToLower(p.BaseToken.Symbol)
		}
		if _, ok := seen[addr]; ok {
			continue
		}
		seen[addr] = struct{}{}
		out = append(out, p)
	}
	return out
}

// pilih top N berdasarkan volume 24h
func pickTopByVolume(pairs []dexScreenerPair, limit int) []dexScreenerPair {
	if len(pairs) == 0 {
		return pairs
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Volume.H24 > pairs[j].Volume.H24
	})
	if len(pairs) > limit {
		return pairs[:limit]
	}
	return pairs
}

// ====== Helper format angka/harga ======

func formatUsdFloat(v float64) string {
	if v == 0 {
		return "0"
	}
	return fmt.Sprintf("%,.0f", v)
}

func formatPriceUsdStr(s string) string {
	if s == "" {
		return "N/A"
	}
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		if v == 0 {
			return "0"
		}
		if v < 0.000001 {
			return fmt.Sprintf("%.10f", v)
		}
		return fmt.Sprintf("%.6f", v)
	}
	return s
}

// ====== Command: dexprice [token] [chain] ======

func CmdDexPrice(ctx context.Context, args []string) (string, error) {
	if len(args) < 2 {
		return "Usage: dexprice [token] [chain]\nExample: dexprice pepe eth", nil
	}

	token := args[0]
	chain := normalizeChainName(args[1])

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

// ====== Command: dexmeme [chain] ======
// Dinamis: cari token bertema "meme" di suatu chain, ambil top 5 by volume 24h.

func CmdDexMeme(ctx context.Context, args []string) (string, error) {
	chain := "ethereum"
	if len(args) >= 1 && args[0] != "" {
		chain = args[0]
	}
	chain = normalizeChainName(chain)

	// query generik "meme <chain>"
	query := "meme " + chain
	pairs, err := searchDexPairs(ctx, query)
	if err != nil {
		return "", err
	}
	if len(pairs) == 0 {
		return fmt.Sprintf("No meme-like pairs found for chain %s", chain), nil
	}

	// filter berdasarkan chain yang diminta
	pairs = filterPairsByChain(pairs, chain)
	if len(pairs) == 0 {
		return fmt.Sprintf("No meme-like pairs found for chain %s", chain), nil
	}

	// deduplicate base tokens dan ambil top 5 by volume
	pairs = dedupPairsByBaseToken(pairs)
	pairs = pickTopByVolume(pairs, 5)

	builder := &strings.Builder{}
	fmt.Fprintf(builder, "🐸 Top meme-style tokens on %s (DexScreener)\n\n", chain)

	for i, p := range pairs {
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
