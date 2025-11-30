package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"teneo-agent-demo1/internal/clients"
	"teneo-agent-demo1/internal/services"
)

// Small styling improvements
func arrow(p float64) string {
	if p > 0 {
		return "▲"
	}
	if p < 0 {
		return "▼"
	}
	return "→"
}

var cexEmoji = map[string]string{
	"coinbase": "🟦",
	"binance":  "🟨",
	"okx":      "⬛",
	"bybit":    "🟧",
	"mexc":     "🟩",
	"kucoin":   "🟪",
	"bitget":   "🔵",
}

func CmdGainersCEX(ctx context.Context, cl *clients.Clients, args []string) (string, error) {
	limit := 5
	if len(args) > 0 {
		if n, err := strconv.Atoi(args[0]); err == nil && n > 0 && n <= 50 {
			limit = n
		}
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("📈 **Top %d gainers per CEX**\n", limit))
	b.WriteString("Across major centralized exchanges:\n")

	for _, ex := range clients.SupportedCexList() {
		emo := cexEmoji[ex]
		if emo == "" {
			emo = "📊"
		}

		gainers, err := cl.FetchCexGainers(ctx, ex, limit)
		if err != nil {
			b.WriteString(fmt.Sprintf("\n%s %s — error: %v\n", emo, strings.ToUpper(ex), err))
			continue
		}
		if len(gainers) == 0 {
			b.WriteString(fmt.Sprintf("\n%s %s — no data\n", emo, strings.ToUpper(ex)))
			continue
		}

		b.WriteString(fmt.Sprintf("\n%s **%s**\n", emo, strings.ToUpper(ex)))
		for i, g := range gainers {
			b.WriteString(fmt.Sprintf(
				"%d) %s — $%s | 24h: %s %s%%\n",
				i+1,
				g.Symbol,
				services.F(g.Price),
				arrow(g.ChangePct),
				services.F(g.ChangePct),
			))
		}
	}

	return b.String(), nil
}

func CmdGainersCompare(ctx context.Context, cl *clients.Clients, args []string) (string, error) {
	if len(args) < 3 {
		return "Usage: gainers_compare [limit] [cex1] [cex2]\n" +
			"Example: gainers_compare 5 binance okx", nil
	}

	limit, err := strconv.Atoi(args[0])
	if err != nil || limit <= 0 || limit > 50 {
		return "Invalid limit. Use number 1–50.", nil
	}

	cex1 := strings.ToLower(args[1])
	cex2 := strings.ToLower(args[2])

	if !clients.IsSupportedCex(cex1) || !clients.IsSupportedCex(cex2) {
		return "Supported CEX: coinbase, binance, okx, bybit, mexc, kucoin, bitget", nil
	}

	emo1 := cexEmoji[cex1]
	if emo1 == "" {
		emo1 = "📊"
	}
	emo2 := cexEmoji[cex2]
	if emo2 == "" {
		emo2 = "📊"
	}

	g1, err1 := cl.FetchCexGainers(ctx, cex1, limit)
	g2, err2 := cl.FetchCexGainers(ctx, cex2, limit)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("📊 **Gainers comparison: %s vs %s** (top %d)\n",
		strings.ToUpper(cex1), strings.ToUpper(cex2), limit))

	// Summary comparison
	if err1 == nil && err2 == nil && len(g1) > 0 && len(g2) > 0 {
		avg1 := 0.0
		avg2 := 0.0
		for _, x := range g1 {
			avg1 += x.ChangePct
		}
		for _, x := range g2 {
			avg2 += x.ChangePct
		}
		avg1 /= float64(len(g1))
		avg2 /= float64(len(g2))

		winner := ""
		if avg1 > avg2 {
			winner = fmt.Sprintf("%s %s **shows stronger avg gain (%.2f%%)** vs %s %.2f%%",
				emo1, strings.ToUpper(cex1), avg1, strings.ToUpper(cex2), avg2)
		} else if avg2 > avg1 {
			winner = fmt.Sprintf("%s %s **shows stronger avg gain (%.2f%%)** vs %s %.2f%%",
				emo2, strings.ToUpper(cex2), avg2, strings.ToUpper(cex1), avg1)
		} else {
			winner = "Both exchanges show similar average gain."
		}

		b.WriteString("\n📌 **Summary Insight**\n")
		b.WriteString("- " + winner + "\n")
		b.WriteString("- Comparing top movers rather than overall market.\n")
	}

	// Detail output
	if err1 != nil {
		b.WriteString(fmt.Sprintf("\n%s %s — error: %v\n", emo1, strings.ToUpper(cex1), err1))
	} else {
		b.WriteString(fmt.Sprintf("\n%s **%s**\n", emo1, strings.ToUpper(cex1)))
		for i, g := range g1 {
			b.WriteString(fmt.Sprintf("%d) %s — $%s | 24h: %s %s%%\n",
				i+1, g.Symbol, services.F(g.Price), arrow(g.ChangePct), services.F(g.ChangePct)))
		}
	}

	if err2 != nil {
		b.WriteString(fmt.Sprintf("\n%s %s — error: %v\n", emo2, strings.ToUpper(cex2), err2))
	} else {
		b.WriteString(fmt.Sprintf("\n%s **%s**\n", emo2, strings.ToUpper(cex2)))
		for i, g := range g2 {
			b.WriteString(fmt.Sprintf("%d) %s — $%s | 24h: %s %s%%\n",
				i+1, g.Symbol, services.F(g.Price), arrow(g.ChangePct), services.F(g.ChangePct)))
		}
	}

	return b.String(), nil
}
