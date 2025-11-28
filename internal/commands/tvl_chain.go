package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"teneo-agent-demo1/internal/clients"
	"teneo-agent-demo1/internal/services"
)

// CmdTVLChain menampilkan total TVL per chain dari DefiLlama
// Usage: tvlchain [chain]
// Example: tvlchain ethereum
func CmdTVLChain(ctx context.Context, cl *clients.Clients, args []string) (string, error) {
	if len(args) == 0 {
		return "Usage: tvlchain [chain]\nExample: tvlchain ethereum", nil
	}

	chain := strings.Join(args, " ")

	endpoint := "https://api.llama.fi/v2/chains"

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")

	resp, err := cl.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("DefiLlama chains error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("DefiLlama status: %d", resp.StatusCode)
	}

	var chains []struct {
		Name string  `json:"name"`
		TVL  float64 `json:"tvl"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&chains); err != nil {
		return "", fmt.Errorf("decode error: %v", err)
	}

	target := strings.ToLower(strings.TrimSpace(chain))
	for _, c := range chains {
		if strings.EqualFold(c.Name, target) {
			// TVL dari DefiLlama adalah float, kita bulatkan dan format pakai pemisah ribuan
			tvlRounded := int64(c.TVL + 0.5)

			return fmt.Sprintf(
				"Chain TVL — %s\nTVL: $%s\nSource: DefiLlama",
				c.Name,
				services.IntComma(tvlRounded),
			), nil
		}
	}

	return fmt.Sprintf("Chain '%s' not found on DefiLlama.", chain), nil
}
