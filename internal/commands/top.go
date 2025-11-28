package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"teneo-agent-demo1/internal/clients"
	"teneo-agent-demo1/internal/services"
)

func CmdTop(ctx context.Context, cl *clients.Clients, args []string) (string, error) {
	limit := 10
	if len(args) > 0 {
		if n, err := strconv.Atoi(args[0]); err == nil && n > 0 && n <= 50 {
			limit = n
		}
	}

	mkts, err := cl.FetchTopMarkets(ctx, limit)
	if err != nil {
		return "", fmt.Errorf("Failed to fetch top markets: %v", err)
	}
	if len(mkts) == 0 {
		return "No data for top coins.", nil
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Top %d coins by market cap:\n", len(mkts)))
	for i, c := range mkts {
		b.WriteString(fmt.Sprintf(
			"%d) %s (%s) — $%s | 24h: %s%% | Mcap: $%s\n",
			i+1,
			c.Name,
			strings.ToUpper(c.Symbol),
			services.F(c.CurrentPrice),
			services.F(c.PriceChangePct24h),
			services.IntComma(int64(c.MarketCap)),
		))
	}
	b.WriteString("\nTip: use `price [symbol]` for more details.")
	return b.String(), nil
}
