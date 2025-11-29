package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// HTTP client khusus DexScreener
var dexHTTPClient = &http.Client{Timeout: 15 * time.Second}

// RNG untuk pilih token random
var dexRand = rand.New(rand.NewSource(time.Now().UnixNano()))

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

// ====== Mapping alias chain (eth → ethereum, dll) ======

var dexChainAliases = map[string]string{
	"eth":      "ethereum",
	"ethereum": "ethereum",

	"sol":     "solana",
	"solana":  "solana",

	"bnb": "bsc",
	"bsc": "bsc",

	"arb":      "arbitrum",
	"arbitrum": "arbitrum",

	"base": "base",

	"matic":   "polygon",
	"polygon": "polygon",

	"op":        "optimism",
	"optimism":  "optimism",
	"opt":       "optimism",
	"optimstic": "optimism",
}

func normalizeChainName(chain string) string {
	c := strings.ToLower(chain)
	if v, ok := dexChainAliases[c]; ok {
		return v
	}
	return c
}

// ====== Pool meme tokens per chain (banyak, nanti di-random 5) ======

var memeTokensByChain = map[string][]string{
	"ethereum": {
		"pepe",
		"mog",
		"shib",
		"floki",
		"doge",
		"bonkler",
		"turbo",
		"pepe2",
		"psp",
		"wifhat",
	},
	"solana": {
		"bonk",
		"wif",
		"popcat",
		"wen",
		"dogwifhat",
		"samoyed",
		"jeets",
		"pnut",
		"ponke",
		"bome",
	},
	"bsc": {
		"babydoge",
		"floki",
		"pepe",
		"meme",
		"doge",
		"cheems",
		"shib",
		"babyshib",
		"cakepunks",
		"minidoge",
	},
	"base": {
		"degen",
		"brian",
		"toshi",
		"pepe",
		"mog",
		"based",
		"basedpepe",
		"turbobase",
		"doginme",
		"catinme",
	},
	"arbitrum": {
		"arbinu",
		"arbshib",
		"magicpepe",
		"arbpepe",
		"arbfloki",
		"arbinu2",
		"arbinu3",
		"arbwif",
		"arbcat",
		"arbmemes",
	},
	"polygon": {
		"polyshib",
		"polydoge",
		"polyfloki",
		"polypepe",
		"dogechain",
		"maticdoge",
		"polymeme",
		"polybonk",
		"polycat",
		"polywif",
	},
	"optimism": {
		"opdoge",
		"opshib",
		"oppepe",
		"opmeme",
		"optimeme",
		"opfloki",
		"opbonk",
		"opwif",
		"opcat",
		"opfrog",
	},
}

func supportedMemeChains() string {
	chains := make([]string, 0, len(memeTokensByChain))
	for c := range memeTokensByChain {
		chains = append(chains, c)
	}
	sort.Strings(chains)
	return strings.Join(chains, ", ")
}

// ====== Helper: pilih N token random, tidak selalu sama ======

func pickRandomTokens(all []string, n int) []string {
	if len(all) == 0 {
		return nil
	}
	if len(all) <= n {
		// kalau stok sedikit, pakai semua (urutan acak sedikit)
		shuffled := make([]string, len(all))
		copy(shuffled, all)
		dexRand.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})
		return shuffled
	}

	shuffled := make([]string, len(all))
	copy(shuffled, all)
	dexRand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled[:n]
}

// ====== Helper umum: cari pair terbaik via DexScreener search ======

func searchBestDexPair(ctx context.Context, token, chain string) (*dexScreenerPair, error) {
	token = strings.ToLower(token)
	chain = normalizeChainName(chain)

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
		return nil, nil // not found
	}

	// Pilih pair yang paling relevan (cocok chain dulu)
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

// ====== Helper format angka ======

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
		if v < 0.000001 {
			return fmt.Sprintf("%.10f", v)
		}
		return fmt.Sprintf("%.6f", v)
	}
	return s
}

// ====== Command 1: dexprice [token] [chain] ======

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

// ====== Command 2: dexmeme [chain] ======
// Menampilkan 5 meme token random dari pool per chain (jadi nggak itu-itu aja)

func CmdDexMeme(ctx context.Context, args []string) (string, error) {
	chain := "ethereum"
	if len(args) >= 1 && args[0] != "" {
		chain = normalizeChainName(args[0])
	} else {
		chain = normalizeChainName(chain)
	}

	allTokens, ok := memeTokensByChain[chain]
	if !ok {
		return fmt.Sprintf(
			"Unknown chain: %s\nSupported chains: %s",
			chain,
			supportedMemeChains(),
		), nil
	}

	// pilih 5 token random dari pool
	selected := pickRandomTokens(allTokens, 5)

	builder := &strings.Builder{}
	fmt.Fprintf(builder, "🐸 Random meme tokens on %s (DexScreener)\n\n", chain)

	for i, symbol := range selected {
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
