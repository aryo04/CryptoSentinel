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

// Struktur Owlracle (disederhanakan, field utama saja)
type owlracleGasResponse struct {
	Chain  string `json:"chain"`
	Speeds []struct {
		Name      string   `json:"name"`
		Estimated float64  `json:"estimatedFee,omitempty"`
		GasPrice  float64  `json:"gasPrice"`
		Unit      string   `json:"unit"`
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

	var slow, normal, fast *struct {
		Name      string
		Estimated float64
		GasPrice  float64
		Unit      string
		Confidence float64
	}
	for i := range data.Speeds {
		s := &data.Speeds[i]
		name := strings.ToLower(s.Name)
		switch {
		case strings.Contains(name, "slow"):
			if slow == nil {
				slow = s
			}
		case strings.Contains(name, "standard"), strings.Contains(name, "normal"):
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
	if normal == nil && len(data.Speeds) > 0 {
		normal = &data.Speeds[0]
	}

	builder := &strings.Builder{}
	fmt.Fprintf(builder, "⛽ **Gas tracker — %s**\n", strings.Title(raw))
	if slow != nil {
		fmt.Fprintf(builder, "• Slow:    %.2f %s\n", slow.GasPrice, slow.Unit)
	}
	if normal != nil {
		fmt.Fprintf(builder, "• Normal:  %.2f %s\n", normal.GasPrice, normal.Unit)
	}
	if fast != nil {
		fmt.Fprintf(builder, "• Fast:    %.2f %s\n", fast.GasPrice, fast.Unit)
	}

	builder.WriteString("\nSource: Owlracle Gas API")
	return builder.String(), nil
}
