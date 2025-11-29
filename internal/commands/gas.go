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

// Mapping user input → Owlracle chain id
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

// Pretty display name per chain
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

// Emoji per chain (biar outputnya enak dibaca)
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

// Single speed from Owlracle
type owlSpeed struct {
	Name       string  `json:"name"`
	Estimated  float64 `json:"estimatedFee,omitempty"` // often in USD if gasPrice == 0
	GasPrice   float64 `json:"gasPrice"`
	Unit       string  `json:"unit"`
	Confidence float64 `json:"confidence,omitempty"`
}

// Owlracle response
type owlracleGasResponse struct {
	Chain  string     `json:"chain"`
	Speeds []owlSpeed `json:"speeds"`
}

// CmdGas: gas [chain]
// Examples:
//   gas ethereum
//   gas arb
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
		// ini error konfigurasi, biarkan bubble up supaya kelihatan di log
		return "", errors.New("OWLRACLE_API_KEY is not set in environment")
	}

	url := fmt.Sprintf("https://api.owlracle.info/v4/%s/gas?apikey=%s", chainID, apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")

	resp, err := gasHTTPClient.Do(req)
	if err != nil {
		// Jaringan / timeout → tampilkan pesan ramah ke user
		return "⚠️ Gas API error. Please try again later.", nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("⚠️ Gas API status %d for chain %s", resp.StatusCode, chainID), nil
	}

	var data owlracleGasResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "⚠️ Failed to decode gas data from Owlracle.", nil
	}
	if len(data.Speeds) == 0 {
		return fmt.Sprintf("No gas data for chain %s", raw), nil
	}

	// Detect slow / normal / fast buckets
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

	// Fallback: kalau tidak ada label normal/standard, pakai index 0
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

	writeLine := func(label string, s *owlSpeed) {
		if s == nil {
			return
		}

		// Case 1: only estimated fee available (often in USD)
		if s.GasPrice <= 0 && s.Estimated > 0 {
			unit := s.Unit
			if unit == "" {
				unit = "USD"
			}
			fmt.Fprintf(&b, "• %s: est. tx cost %s %s", label, services.F(s.Estimated), unit)
		} else {
			// Case 2: normal gas price, usually gwei
			fmt.Fprintf(&b, "• %s: %s %s", label, services.F(s.GasPrice), s.Unit)
		}

		// Add confidence if available
		if s.Confidence > 0 {
			fmt.Fprintf(&b, "  (confidence ~%.0f%%)", s.Confidence)
		}
		b.WriteString("\n")
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

	// Extra explanation block
	b.WriteString("\nHints:\n")
	b.WriteString("- Slow: cheaper but may take longer to confirm.\n")
	b.WriteString("- Normal: balanced cost vs confirmation speed (typical default).\n")
	b.WriteString("- Fast: higher fee, better chance to be included in the next blocks.\n")

	b.WriteString("\nSource: Owlracle Gas API")
	return b.String(), nil
}
