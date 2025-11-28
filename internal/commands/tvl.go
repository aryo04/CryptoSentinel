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
	protocol := strings.Join(args, " ")

	res, err := cl.FetchProtocolTVL(ctx, protocol)
	if err != nil {
		return fmt.Sprintf("Failed to fetch TVL for '%s': %v", protocol, err), nil
	}

	out := fmt.Sprintf("%s (chain: %s)\nTVL (latest): $%s\nSource: DefiLlama",
		res.Name, res.Chain, services.IntComma(int64(res.TVL)))
	return out, nil
}
