package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

var gasHTTPClient = &http.Client{Timeout: 15 * time.Second}

// Mapping nama chain yang user ketik → chainId Owlracle
var gasChainMap = map[string]string{
	"eth":       "eth",
	"ethereum":  "eth",
	"bsc":       "bsc",
	"bnb":       "bsc",
	"polygon":   "polygon",
	"matic":     "polygon",
	"arbitrum":  "arbitrum",
	"arb":       "arbitrum",
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

// Struktur Owlracle (pakai anonymous struct seperti sebelumnya)
type owlracleGasResponse struct {
	Chain  string `json:"chain"`
	Speeds []struct {
		Name       string  `json:"name"`
		Estimated  float64 `json:"estimatedFee,omitempty"`
		GasPrice   float64 `json:"gasPrice"`
		Unit       string  `json:"unit"`
		Confidence float64 `json:"confidence,omitempty"`
	} `json:"speeds"`
}

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
		// Dibikin user-friendly, bukan error teknis
		return "Gas tracker is not configured (missing OWLRACLE_API_KEY in environment).", nil
	}

	url := fmt.Sprintf("https://api.owlracle.info/v4/%s/gas?apikey=%s", chainID, apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")

	resp, err := gasHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("gas API error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gas API status %d for chain %s", resp.StatusCode, chainID)
	}

	var data owlracleGasResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("decode gas JSON error: %w", err)
	}
	if len(data.Speeds) == 0 {
		return fmt.Sprintf("No gas data for chain %s", raw), nil
	}

	// Pakai index biar nggak ribet tipe struct
	slowIdx, normalIdx, fastIdx := -1, -1, -1

	for i := range data.Speeds {
		s := data.Speeds[i]
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

	// fallback kalau label tidak pas
	if normalIdx == -1 && len(data.Speeds) > 0 {
		normalIdx = 0
	}

	builder := &strings.Builder{}
	fmt.Fprintf(builder, "⛽ Gas tracker — %s\n", titleCase(raw))

	if slowIdx >= 0 {
		s := data.Speeds[slowIdx]
		fmt.Fprintf(builder, "• Slow:   %.2f %s\n", s.GasPrice, s.Unit)
	}
	if normalIdx >= 0 {
		s := data.Speeds[normalIdx]
		fmt.Fprintf(builder, "• Normal: %.2f %s\n", s.GasPrice, s.Unit)
	}
	if fastIdx >= 0 {
		s := data.Speeds[fastIdx]
		fmt.Fprintf(builder, "• Fast:   %.2f %s\n", s.GasPrice, s.Unit)
	}

	builder.WriteString("\nSource: Owlracle Gas API")
	return builder.String(), nil
}

// titleCase pengganti strings.Title (deprecated)
func titleCase(s string) string {
	if s == "" {
		return s
	}
	if len(s) == 1 {
		return strings.ToUpper(s)
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}
