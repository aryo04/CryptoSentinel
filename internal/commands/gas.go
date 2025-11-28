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

type owlracleSpeed struct {
	Name       string  `json:"name"`
	Estimated  float64 `json:"estimatedFee,omitempty"`
	GasPrice   float64 `json:"gasPrice"`
	Unit       string  `json:"unit"`
	Confidence float64 `json:"confidence,omitempty"`
}

// Struktur Owlracle (disederhanakan)
type owlracleGasResponse struct {
	Chain  string          `json:"chain"`
	Speeds []owlracleSpeed `json:"speeds"`
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
		// Jangan return error ke SDK, jelaskan saja
		return "Gas tracker is not configured yet (missing OWLRACLE_API_KEY on the server). Ask the operator to add it if you need this feature.", nil
	}

	url := fmt.Sprintf("https://api.owlracle.info/v4/%s/gas?apikey=%s", chainID, apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "❌ Failed to build gas request.", nil
	}
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")

	resp, err := gasHTTPClient.Do(req)
	if err != nil {
		return fmt.Sprintf("❌ Gas API error: %v", err), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("❌ Gas API returned status %d for chain %s. Please try again later.", resp.StatusCode, strings.Title(raw)), nil
	}

	var data owlracleGasResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return fmt.Sprintf("❌ Failed to decode gas data: %v", err), nil
	}
	if len(data.Speeds) == 0 {
		return fmt.Sprintf("No gas data available for chain %s.", raw), nil
	}

	var slow, normal, fast *owlracleSpeed

	for i := range data.Speeds {
		s := &data.Speeds[i]
		name := strings.ToLower(s.Name)

		switch {
		case strings.Contains(name, "slow"):
			if slow == nil {
				slow = s
			}
		case strings.Contains(name, "standard") || strings.Contains(name, "normal"):
			if normal == nil {
				normal = s
			}
		case strings.Contains(name, "fast"):
			if fast == nil {
				fast = s
			}
		}
	}

	// fallback kalau label tidak pas
	if normal == nil {
		normal = &data.Speeds[0]
	}

	b := &strings.Builder{}
	fmt.Fprintf(b, "⛽ Gas tracker — %s\n", strings.Title(raw))

	// catatan: Owlracle biasanya mengembalikan Gwei
	if slow != nil {
		fmt.Fprintf(b, "• Slow:   %.2f %s\n", slow.GasPrice, slow.Unit)
	}
	if normal != nil {
		fmt.Fprintf(b, "• Normal: %.2f %s\n", normal.GasPrice, normal.Unit)
	}
	if fast != nil {
		fmt.Fprintf(b, "• Fast:   %.2f %s\n", fast.GasPrice, fast.Unit)
	}

	b.WriteString("\nSource: Owlracle Gas API")
	return b.String(), nil
}
