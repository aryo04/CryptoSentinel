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

// Nice labels for output
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

// Simple emoji “badge” per chain
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

// One speed entry from Owlracle
type owlSpeed struct {
	Name       string  `json:"name"`
	Estimated  float64 `json:"estimatedFee,omitempty"` // est tx fee in native or USD depending on plan
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
// Example: gas ethereum
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

	url := fmt.Sprintf("https://api.owlracle.info/v4/%s/gas?apikey=%s", chainID, apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")

	resp, err := gasHTTPClient.Do(req)
	if err != nil {
		// User-facing message harus friendly
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

	// Detect slow / normal / fast indexes
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

	// Fallback if no explicit "normal" label
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

	// Helper to format each speed nicely
	writeLine := func(label string, s *owlSpeed) {
		if s == nil {
			return
		}

		// Main price part
		if s.GasPrice <= 0 && s.Estimated > 0 {
			// Some plans use estimatedFee only
			fmt.Fprintf(&b, "• %s: est. tx cost %s (%s)\n", label, services.F(s.Estimated), s.Unit)
		} else {
			fmt.Fprintf(&b, "• %s: %s %s\n", label, services.F(s.GasPrice), s.Unit)
		}

		// Extra hints (confidence + estimated fee if both exist)
		extras := make([]string, 0, 2)
		if s.Confidence > 0 {
			extras = append(extras, fmt.Sprintf("~%s%% confidence", services.F(s.Confidence)))
		}
		if s.GasPrice > 0 && s.Estimated > 0 {
			// We don’t assume USD here, just show unit as provided
			extras = append(extras, fmt.Sprintf("est. tx cost ~%s %s", services.F(s.Estimated), s.Unit))
		}
		if len(extras) > 0 {
			fmt.Fprintf(&b, "    (%s)\n", strings.Join(extras, " · "))
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

	b.WriteString("\nHints:\n")
	b.WriteString("- Slow: cheaper but may take longer to confirm.\n")
	b.WriteString("- Normal: balanced cost vs confirmation speed (default for most users).\n")
	b.WriteString("- Fast: more expensive but higher chance to be included in the next blocks.\n")
	b.WriteString("\nSource: Owlracle Gas API")

	return b.String(), nil
}
