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

// CmdTVLChain shows total TVL for a chain from DefiLlama,
// along with its rank and share of total DeFi TVL.
// Usage: tvlchain [chain]
// Example: tvlchain ethereum
func CmdTVLChain(ctx context.Context, cl *clients.Clients, args []string) (string, error) {
	if len(args) == 0 {
		return "Usage: tvlchain [chain]\nExample: tvlchain ethereum", nil
	}

	chainInput := strings.Join(args, " ")

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

	target := strings.ToLower(strings.TrimSpace(chainInput))

	// Find target chain, compute global stats (total TVL, rank, dominance)
	var (
		foundIndex = -1
		totalTVL   float64
	)
	for i, c := range chains {
		totalTVL += c.TVL
		if strings.EqualFold(c.Name, target) {
			foundIndex = i
		}
	}

	if foundIndex == -1 {
		return fmt.Sprintf("Chain '%s' not found on DefiLlama.", chainInput), nil
	}

	targetChain := chains[foundIndex]
	tvl := targetChain.TVL
	tvlRounded := int64(tvl + 0.5)

	// Rank: how many chains have higher TVL
	rank := 1
	for _, c := range chains {
		if c.TVL > tvl {
			rank++
		}
	}
	totalChains := len(chains)

	// Share of total DeFi TVL (in %)
	dominance := 0.0
	if totalTVL > 0 {
		dominance = tvl / totalTVL * 100.0
	}

	out := fmt.Sprintf(
		"Chain TVL — %s\n"+
			"Rank: #%d of %d tracked chains\n"+
			"TVL: $%s\n"+
			"Share of total DeFi TVL: %s%%\n\n"+
			"Notes:\n"+
			"- Rank and dominance are relative to all chains tracked by DefiLlama.\n"+
			"- TVL is an approximate snapshot and may lag a few minutes behind on-chain data.\n\n"+
			"Source: DefiLlama",
		targetChain.Name,
		rank,
		totalChains,
		services.IntComma(tvlRounded),
		services.F(dominance),
	)

	return out, nil
}
