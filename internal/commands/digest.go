package commands

import (
	"context"
	"fmt"
	"strings"

	"teneo-agent-demo1/internal/clients"
	"teneo-agent-demo1/internal/services"
)

func CmdDigest(ctx context.Context, cl *clients.Clients, args []string) (string, error) {
	period := "now"
	if len(args) > 0 {
		period = strings.ToLower(args[0])
	}

	mkts, err := cl.FetchTopMarkets(ctx, 5)
	if err != nil {
		return "", fmt.Errorf("Failed to fetch digest: %v", err)
	}
	if len(mkts) == 0 {
		return "No data for digest.", nil
	}

	title := "Market Digest (Top 5 by Market Cap)"
	if period == "daily" {
		title = "Daily Market Digest (Top 5)"
	}

	var b strings.Builder
	b.WriteString(title + ":\n")
	for i, c := range mkts {
		b.WriteString(fmt.Sprintf(
			"%d) %s (%s) — $%s | 24h: %s%% | 7d: %s%% | Market Cap: $%s\n",
			i+1,
			c.Name,
			strings.ToUpper(c.Symbol),
			services.F(c.CurrentPrice),
			services.F(c.PriceChangePct24h),
			services.F(c.PriceChangePct7d),
			services.IntComma(int64(c.MarketCap)),
		))
	}
	b.WriteString("\nTip: use `price [symbol]` or `compare [a] [b]` for deeper checks.")
	return b.String(), nil
}
