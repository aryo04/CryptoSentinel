package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"teneo-agent-demo1/internal/services"
)

var gasHTTPClient = &http.Client{Timeout: 15 * time.Second}

// Mapping nama chain user → network Owlracle
var gasChainMap = map[string]string{
	"eth":       "eth",
	"ethereum":  "eth",
	"bsc":       "bsc",
	"bnb":       "bsc",
	"polygon":   "polygon",
	"matic":     "polygon",
	"arbitrum":  "arb",
	"arb":       "arb",
	"optimism":  "optimism",
	"op":        "optimism",
	"base":      "base",
	"avax":      "avax",
	"avalanche": "avax",
	"fantom":    "fantom",
	"ftm":       "fantom",
	"gnosis":    "gnosis",
	"gno":       "gnosis",
	"celo":      "celo",
}

// Label cantik + emoji buat output
var gasPrettyName = map[string]string{
	"eth":      "Ethereum",
	"bsc":      "BSC",
	"polygon":  "Polygon",
	"arb":      "Arbitrum",
	"optimism": "Optimism",
	"base":     "Base",
	"avax":     "Avalanche",
	"fantom":   "Fantom",
	"gnosis":   "Gnosis",
	"celo":     "Celo",
}

var gasEmoji = map[string]string{
	"eth":      "🟦",
	"bsc":      "🟨",
	"polygon":  "🟪",
	"arb":      "🧵",
	"optimism": "🟥",
	"base":     "🟦",
	"avax":     "🏔️",
	"fantom":   "👻",
	"gnosis":   "🟩",
	"celo":     "🟨",
}

// Struktur respons Owlracle (disederhanakan)
type owlracleGasResponse struct {
	Chain  string `json:"chain"`
	Speeds []struct {
		Name       string  `json:"name"`
		Estimated  float64 `json:"estimatedFee,omitempty"` // biasanya dalam USD (default feeinusd=true)
		GasPrice   float64 `json:"gasPrice"`              // kadang 0 kalau pakai feeinusd
		Unit       string  `json:"unit"`
		Confidence float64 `json:"confidence,omitempty"`
	} `json:"speeds"`
}

// CmdGas: gas [chain]
// contoh: gas bsc, gas arbitrum
func CmdGas(ctx context.Context, args []string) (string, error) {
	if len(args) == 0 {
		return "Usage: gas [chain]\nExample: gas ethereum", nil
	}

	raw := strings.ToLower(args[0])
	chainID, ok := gasChainMap[raw]
	if !ok {
		return "Unsupported chain. Supported: ethereum, arbitrum, base, bsc, polygon, optimism, avalanche, fantom, gnosis, celo", nil
	}

	apiKey := os.Getenv("OWLRACLE_API_KEY")
	if apiKey == "" {
		return "", errors.New("missing OWLRACLE_API_KEY in environment")
	}

	// basic URL, pakai default param Owlracle
	url := fmt.Sprintf("https://api.owlracle.info/v4/%s/gas?apikey=%s", chainID, apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")

	resp, err := gasHTTPClient.Do(req)
	if err != nil {
		return "⚠️ Gas API error. Please try again later.", nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// jangan lempar error ke luar, tapi kirimkan pesan ke user
		return fmt.Sprintf("⚠️ Gas API status %d for chain %s", resp.StatusCode, chainID), nil
	}

	var data owlracleGasResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "⚠️ Failed to decode gas data from Owlracle.", nil
	}
	if len(data.Speeds) == 0 {
		return fmt.Sprintf("No gas data for chain %s", raw), nil
	}

	// Cari indeks slow / normal / fast (pakai index, bukan pointer ke anonymous type)
	slowIdx, normalIdx, fastIdx := -1, -1, -1

	for i := range data.Speeds {
		s := &data.Speeds[i]
		name := strings.ToLower(s.Name)

		switch {
			case strings.Contains(name, "slow"):
				if slowIdx == -1 {
					slowIdx = i
				}
			case strings.Contains(name, "standard"), strings.Contains(name, "normal"):
				if normalIdx == -1 {
					normalIdx = i
				}
			case strings.Contains(name, "fast"):
				if fastIdx == -1 {
					fastIdx = i
				}
		}
	}

	// fallback kalau tidak ketemu label "normal"
	if normalIdx == -1 {
		normalIdx = 0
	}

	prettyName := gasPrettyName[chainID]
	if prettyName == "" {
		prettyName = strings.Title(raw)
	}
	flag := gasEmoji[chainID]
	if flag == "" {
		flag = "⛽"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s Gas tracker — %s\n", flag, prettyName)

	writeLine := func(label string, s *struct {
		Name       string
		Estimated  float64
		GasPrice   float64
		Unit       string
		Confidence float64
	}) {
		if s == nil {
			return
		}

		// Kalau gasPrice=0 tapi Estimated>0 → kemungkinan API pakai feeinusd
		if s.GasPrice <= 0 && s.Estimated > 0 {
			fmt.Fprintf(&b, "• %s: est. fee %s USD\n", label, services.F(s.Estimated))
		} else {
			fmt.Fprintf(&b, "• %s: %s %s\n", label, services.F(s.GasPrice), s.Unit)
		}
	}

	if slowIdx != -1 {
		writeLine("Slow", &data.Speeds[slowIdx])
	}
	if normalIdx != -1 {
		writeLine("Normal", &data.Speeds[normalIdx])
	}
	if fastIdx != -1 {
		writeLine("Fast", &data.Speeds[fastIdx])
	}

	b.WriteString("\nSource: Owlracle Gas API")
	return b.String(), nil
}
