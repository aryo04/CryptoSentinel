package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"teneo-agent-demo1/internal/clients"
	"teneo-agent-demo1/internal/services"
)

func CmdGainersCEX(ctx context.Context, cl *clients.Clients, args []string) (string, error) {
	limit := 5
	if len(args) > 0 {
		if n, err := strconv.Atoi(args[0]); err == nil && n > 0 && n <= 50 {
			limit = n
		}
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Top gainers per CEX (limit %d):\n", limit))

	for _, ex := range clients.SupportedCexList() {
		gainers, err := cl.FetchCexGainers(ctx, ex, limit)
		if err != nil {
			b.WriteString(fmt.Sprintf("\n[%s] error: %v\n", strings.ToUpper(ex), err))
			continue
		}
		if len(gainers) == 0 {
			b.WriteString(fmt.Sprintf("\n[%s] no data\n", strings.ToUpper(ex)))
			continue
		}
		b.WriteString(fmt.Sprintf("\n[%s]\n", strings.ToUpper(ex)))
		for i, g := range gainers {
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

func CmdGainersCompare(ctx context.Context, cl *clients.Clients, args []string) (string, error) {
	if len(args) < 3 {
		return "Usage: gainers_compare [limit] [cex1] [cex2]\nExample: gainers_compare 5 binance okx", nil
	}
	limit, err := strconv.Atoi(args[0])
	if err != nil || limit <= 0 || limit > 50 {
		return "Invalid limit. Use number 1–50.", nil
	}
	cex1 := strings.ToLower(args[1])
	cex2 := strings.ToLower(args[2])

	if !clients.IsSupportedCex(cex1) || !clients.IsSupportedCex(cex2) {
		return "Supported CEX: binance, okx, bybit, kucoin, bitget", nil
	}

	g1, err1 := cl.FetchCexGainers(ctx, cex1, limit)
	g2, err2 := cl.FetchCexGainers(ctx, cex2, limit)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Gainers comparison (%s vs %s), top %d:\n",
		strings.ToUpper(cex1), strings.ToUpper(cex2), limit))

	if err1 != nil {
		b.WriteString(fmt.Sprintf("[%s] error: %v\n", strings.ToUpper(cex1), err1))
	} else {
		b.WriteString(fmt.Sprintf("\n[%s]\n", strings.ToUpper(cex1)))
		for i, g := range g1 {
			b.WriteString(fmt.Sprintf("%d) %s — $%s | 24h: %s%%\n",
				i+1, g.Symbol, services.F(g.Price), services.F(g.ChangePct)))
		}
	}

	if err2 != nil {
		b.WriteString(fmt.Sprintf("\n[%s] error: %v\n", strings.ToUpper(cex2), err2))
	} else {
		b.WriteString(fmt.Sprintf("\n[%s]\n", strings.ToUpper(cex2)))
		for i, g := range g2 {
			b.WriteString(fmt.Sprintf("%d) %s — $%s | 24h: %s%%\n",
				i+1, g.Symbol, services.F(g.Price), services.F(g.ChangePct)))
		}
	}

	return b.String(), nil
}
