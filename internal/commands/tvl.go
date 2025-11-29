package commands

import (
	"context"
	"fmt"
	"strings"

	"teneo-agent-demo1/internal/clients"
	"teneo-agent-demo1/internal/services"
)

func CmdTVL(ctx context.Context, cl *clients.Clients, args []string) (string, error) {
	if len(args) == 0 {
		return "Usage: tvl [protocol]\nExample: tvl uniswap", nil
	}
	protocolInput := strings.Join(args, " ")

	res, err := cl.FetchProtocolTVL(ctx, protocolInput)
	if err != nil {
		return fmt.Sprintf("Failed to fetch TVL for '%s': %v", protocolInput, err), nil
	}

	// Round TVL to nearest dollar and format with thousands separator
	tvlRounded := int64(res.TVL + 0.5)
	tvlStr := services.IntComma(tvlRounded)

	chainLabel := res.Chain
	if chainLabel == "" {
		chainLabel = "Unknown / multi-chain"
	}

	// Slightly richer, still concise, all in English
	out := fmt.Sprintf(
		"Protocol TVL — %s\n"+
			"Chain: %s\n"+
			"Latest TVL: $%s\n\n"+
			"Notes:\n"+
			"- TVL reflects the most recent snapshot reported by DefiLlama.\n"+
			"- For chain-wide liquidity view, try: tvlchain [chain_name]\n\n"+
			"Source: DefiLlama",
		res.Name,
		chainLabel,
		tvlStr,
	)

	return out, nil
}
