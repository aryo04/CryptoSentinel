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

	"op":       "optimism",
	"optimism": "optimism",
}

func normalizeChainName(chain string) string {
	c := strings.ToLower(chain)
	if v, ok := dexChainAliases[c]; ok {
		return v
	}
	return c
}

// ====== Meme token struct (mapping simbol + nama) ======

type MemeToken struct {
	Symbol string // ticker, untuk query DexScreener
	Name   string // nama token lebih lengkap, untuk display
}

// ====== Pool meme tokens per chain (banyak & real, dipilih random 5) ======

var memeTokensByChain = map[string][]MemeToken{
	// ===== Ethereum (ERC-20 meme coins) =====
	"ethereum": {
		{Symbol: "pepe", Name: "Pepe"},
		{Symbol: "shib", Name: "Shiba Inu"},
		{Symbol: "bitdoge", Name: "Bitcoin Doge"},
		{Symbol: "doge", Name: "Wrapped Dogecoin"},
		{Symbol: "best", Name: "Best Wallet Token"},
		{Symbol: "turbo", Name: "Turbo"},
		{Symbol: "mog", Name: "Mog Coin"},
		{Symbol: "ladys", Name: "Milady Meme Coin"},
		{Symbol: "pork", Name: "PepeFork (PORK)"},
		{Symbol: "spx", Name: "SPX6900"},
		{Symbol: "elon", Name: "Dogelon Mars"},
		{Symbol: "akita", Name: "Akita Inu"},
		{Symbol: "kishu", Name: "Kishu Inu"},
		{Symbol: "hoge", Name: "Hoge Finance"},
		{Symbol: "bone", Name: "Bone ShibaSwap"},
		{Symbol: "leash", Name: "Doge Killer (LEASH)"},
		{Symbol: "ape", Name: "ApeCoin"},
		{Symbol: "spx", Name: "SPX6900"},
		{Symbol: "bonkler", Name: "Bonkler"},
		{Symbol: "bob", Name: "Bob Token"},
		{Symbol: "kek", Name: "KEK"},
		{Symbol: "rekt", Name: "Rekt"},
		{Symbol: "pepe2", Name: "Pepe 2.0"},
		{Symbol: "pepe3", Name: "Pepe 3.0"},
	},

	// ===== Solana native meme coins =====
	"solana": {
		{Symbol: "bonk", Name: "BONK"},
		{Symbol: "wif", Name: "dogwifhat"},
		{Symbol: "popcat", Name: "Popcat"},
		{Symbol: "bome", Name: "BOOK OF MEME"},
		{Symbol: "mew", Name: "Cat in a Dogs World"},
		{Symbol: "ponke", Name: "PONKE"},
		{Symbol: "michi", Name: "MICHI"},
		{Symbol: "samo", Name: "Samoyedcoin"},
		{Symbol: "slerf", Name: "SLERF"},
		{Symbol: "jeet", Name: "JEET"},
		{Symbol: "boden", Name: "BODEN"},
		{Symbol: "pippin", Name: "Pippin"},
		{Symbol: "ban", Name: "Ban"},
		{Symbol: "tuf", Name: "TUF"},
		{Symbol: "wolf", Name: "Wolf on Sol"},
		{Symbol: "pnut", Name: "Peanut"},
		{Symbol: "kitty", Name: "Kitty on Sol"},
		{Symbol: "tate", Name: "TATE"},
	},

	// ===== BNB Chain / BSC meme coins =====
	"bsc": {
		{Symbol: "babydoge", Name: "Baby Doge Coin"},
		{Symbol: "shib", Name: "Shiba Inu (BSC)"},
		{Symbol: "doge", Name: "Dogecoin (BSC)"},
		{Symbol: "floki", Name: "Floki Inu (BSC)"},
		{Symbol: "safemoon", Name: "SafeMoon"},
		{Symbol: "cake", Name: "PancakeSwap"},
		{Symbol: "poodl", Name: "Poodl Token"},
		{Symbol: "akita", Name: "Akita Inu (BSC)"},
		{Symbol: "kishu", Name: "Kishu Inu (BSC)"},
		{Symbol: "hoge", Name: "Hoge Finance (BSC)"},
		{Symbol: "elon", Name: "Dogelon Mars (BSC)"},
		{Symbol: "wojak", Name: "Wojak (BSC)"},
		{Symbol: "pepe", Name: "Pepe (BSC)"},
		{Symbol: "pepe2", Name: "Pepe 2.0 (BSC)"},
		{Symbol: "doki", Name: "Doki Doki"},
		{Symbol: "doggy", Name: "Doggy"},
		{Symbol: "moon", Name: "Moon Token"},
		{Symbol: "meme", Name: "Meme Token"},
	},

	// ===== Base chain meme coins =====
	"base": {
		{Symbol: "brett", Name: "Brett on Base"},
		{Symbol: "degen", Name: "Degen"},
		{Symbol: "toshi", Name: "Toshi"},
		{Symbol: "brian", Name: "Brian"},
		{Symbol: "based", Name: "Based"},
		{Symbol: "mfer", Name: "mfercoin"},
		{Symbol: "tybg", Name: "TYBG (Thank You Base God)"},
		{Symbol: "doginme", Name: "Dog in Me"},
		{Symbol: "catinme", Name: "Cat in Me"},
		{Symbol: "pepe", Name: "Pepe on Base"},
		{Symbol: "mog", Name: "Mog on Base"},
		{Symbol: "bongocat", Name: "Bongo Cat"},
		{Symbol: "smol", Name: "Smol Base Meme"},
		{Symbol: "blue", Name: "Blue Base"},
		{Symbol: "fish", Name: "Fish on Base"},
	},

	// ===== Arbitrum meme coins =====
	"arbitrum": {
		{Symbol: "aidoge", Name: "AI Doge"},
		{Symbol: "caw", Name: "A Hunters Dream (CAW)"},
		{Symbol: "pogai", Name: "Poor Guy (POGAI)"},
		{Symbol: "nut", Name: "Nutcoin"},
		{Symbol: "smol", Name: "SmolCoin"},
		{Symbol: "twoge", Name: "Twoge Inu"},
		{Symbol: "gnome", Name: "Gnome"},
		{Symbol: "nyan", Name: "ArbiNYAN"},
		{Symbol: "grimace", Name: "Grimace (ARB)"},
		{Symbol: "arbinu", Name: "ArbInu"},
		{Symbol: "arbpepe", Name: "ArbPepe"},
		{Symbol: "arbfloki", Name: "ArbFloki"},
		{Symbol: "arbwif", Name: "ArbWIF"},
		{Symbol: "arbshib", Name: "ArbShib"},
		{Symbol: "arbcat", Name: "ArbCat"},
	},

	// ===== Polygon POS =====
	"polygon": {
		{Symbol: "doge", Name: "Dogecoin (Polygon)"},
		{Symbol: "shib", Name: "Shiba Inu (Polygon)"},
		{Symbol: "floki", Name: "Floki (Polygon)"},
		{Symbol: "pepe", Name: "Pepe (Polygon)"},
		{Symbol: "wojak", Name: "Wojak (Polygon)"},
		{Symbol: "kishu", Name: "Kishu (Polygon)"},
		{Symbol: "akita", Name: "Akita (Polygon)"},
		{Symbol: "elon", Name: "Dogelon (Polygon)"},
		{Symbol: "babydoge", Name: "Baby Doge (Polygon)"},
		{Symbol: "hoge", Name: "Hoge (Polygon)"},
		{Symbol: "meme", Name: "Meme Token (Polygon)"},
	},

	// ===== Optimism =====
	"optimism": {
		{Symbol: "pepe", Name: "Pepe (Optimism)"},
		{Symbol: "shib", Name: "Shiba Inu (Optimism)"},
		{Symbol: "doge", Name: "Dogecoin (Optimism)"},
		{Symbol: "floki", Name: "Floki (Optimism)"},
		{Symbol: "wojak", Name: "Wojak (Optimism)"},
		{Symbol: "bonk", Name: "BONK (Optimism)"},
		{Symbol: "wif", Name: "WIF (Optimism)"},
		{Symbol: "degen", Name: "Degen (Optimism)"},
		{Symbol: "mfer", Name: "mfercoin (Optimism)"},
		{Symbol: "dogelon", Name: "Dogelon Mars (Optimism)"},
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

func pickRandomTokens(all []MemeToken, n int) []MemeToken {
	if len(all) == 0 {
		return nil
	}
	if len(all) <= n {
		shuffled := make([]MemeToken, len(all))
		copy(shuffled, all)
		dexRand.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})
		return shuffled
	}

	shuffled := make([]MemeToken, len(all))
	copy(shuffled, all)
	dexRand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled[:n]
}

// ====== Helper utama: cari pair terbaik via DexScreener search ======

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

	// STRICT: hanya pair dengan ChainId persis = chain
	var best *dexScreenerPair
	for i := range data.Pairs {
		p := &data.Pairs[i]
		if strings.EqualFold(p.ChainId, chain) {
			best = p
			break
		}
	}

	// kalau nggak ada chain yg match, anggap not found (jangan fallback ke chain lain)
	return best, nil
}

// ====== Helper format angka ======

func formatUsdFloat(v float64) string {
	// tanpa ribuan dulu, biar aman
	return fmt.Sprintf("%.0f", v)
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
		// jangan bikin bot error generic, kirim pesan saja
		return fmt.Sprintf("Error fetching price from DexScreener: %v", err), nil
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
// Menampilkan 5 meme token random dari pool per chain

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

	selected := pickRandomTokens(allTokens, 5)

	builder := &strings.Builder{}
	fmt.Fprintf(builder, "🐸 Random meme tokens on %s (DexScreener)\n\n", chain)

	for i, t := range selected {
		p, err := searchBestDexPair(ctx, t.Symbol, chain)
		if err != nil {
			fmt.Fprintf(builder, "%d) %s — error: %v\n\n", i+1, strings.ToUpper(t.Symbol), err)
			continue
		}
		if p == nil {
			fmt.Fprintf(builder, "%d) %s — pair not found on %s\n\n", i+1, strings.ToUpper(t.Symbol), chain)
			continue
		}

		priceStr := formatPriceUsdStr(p.PriceUsd)

		displayName := t.Name
		if displayName == "" {
			displayName = p.BaseToken.Name
		}

		fmt.Fprintf(builder,
			"%d) %s (%s)\n"+
				"   Name: %s\n"+
				"   Pair: %s / %s (%s)\n"+
				"   Price: $%s\n"+
				"   Liquidity: $%s\n"+
				"   Volume 24h: $%s\n"+
				"   DEX: %s\n"+
				"   Link: %s\n\n",
			i+1,
			strings.ToUpper(p.BaseToken.Symbol),
			p.ChainId,
			displayName,
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
