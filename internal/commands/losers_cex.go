package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"teneo-agent-demo1/internal/clients"
	"teneo-agent-demo1/internal/services"
)

// losers_cex [limit]
// Example: losers_cex 5
func CmdLosersCEX(ctx context.Context, cl *clients.Clients, args []string) (string, error) {
	limit := 5
	if len(args) > 0 {
		if n, err := strconv.Atoi(args[0]); err == nil && n > 0 && n <= 50 {
			limit = n
		}
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Top losers per CEX (limit %d):\n", limit))

	for _, ex := range clients.SupportedCexList() {
		losers, err := cl.FetchCexLosers(ctx, ex, limit)
		if err != nil {
			b.WriteString(fmt.Sprintf("\n[%s] error: %v\n", strings.ToUpper(ex), err))
			continue
		}
		if len(losers) == 0 {
			b.WriteString(fmt.Sprintf("\n[%s] no data\n", strings.ToUpper(ex)))
			continue
		}
		b.WriteString(fmt.Sprintf("\n[%s]\n", strings.ToUpper(ex)))
		for i, g := range losers {
			b.WriteString(fmt.Sprintf(
				"%d) %s — $%s | 24h: %s%%\n",
				i+1,
				g.Symbol,
				services.F(g.Price),
				services.F(g.ChangePct),
			))
		}
	}

	return b.String(), nil
}
