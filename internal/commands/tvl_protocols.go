package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"teneo-agent-demo1/internal/clients"
	"teneo-agent-demo1/internal/services"
)

// CmdTVLProtocols menampilkan top protokol DeFi berdasarkan TVL pada sebuah chain
// Usage: tvlprotocols [chain]
// Example: tvlprotocols arbitrum
func CmdTVLProtocols(ctx context.Context, cl *clients.Clients, args []string) (string, error) {
	if len(args) == 0 {
		return "Usage: tvlprotocols [chain]\nExample: tvlprotocols arbitrum", nil
	}

	chain := strings.Join(args, " ")

	endpoint := "https://api.llama.fi/protocols"

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")

	resp, err := cl.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("DefiLlama protocols error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("DefiLlama status: %d", resp.StatusCode)
	}

	var protos []struct {
		Name   string   `json:"name"`
		TVL    float64  `json:"tvl"`
		Chains []string `json:"chains"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&protos); err != nil {
		return "", fmt.Errorf("decode error: %v", err)
	}

	target := strings.ToLower(strings.TrimSpace(chain))

	type protoInfo struct {
		Name string
		TVL  float64
	}

	var filtered []protoInfo
	for _, p := range protos {
		for _, ch := range p.Chains {
			if strings.EqualFold(ch, target) {
				filtered = append(filtered, protoInfo{
					Name: p.Name,
					TVL:  p.TVL,
				})
				break
			}
		}
	}

	if len(filtered) == 0 {
		return fmt.Sprintf("No protocols found for chain '%s'.", chain), nil
	}

	// Urutkan dari TVL terbesar ke terkecil
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].TVL > filtered[j].TVL
	})

	limit := 10
	if len(filtered) < limit {
		limit = len(filtered)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Top %d protocols by TVL on %s:\n", limit, chain))

	for i := 0; i < limit; i++ {
		p := filtered[i]
		tvlRounded := int64(p.TVL + 0.5)

		b.WriteString(fmt.Sprintf(
			"%d) %s — TVL: $%s\n",
			i+1,
			p.Name,
			services.IntComma(tvlRounded),
		))
	}

	b.WriteString("\nSource: DefiLlama")
	return b.String(), nil
}
