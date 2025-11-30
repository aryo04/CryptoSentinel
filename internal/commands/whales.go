package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"teneo-agent-demo1/internal/clients"
)

func CmdWhales(ctx context.Context, cl *clients.Clients, args []string) (string, error) {
	if cl == nil || cl.Whale == nil {
		return "Whale tracking client not initialized.", nil
	}

	if len(args) == 0 {
		return "Usage: whales [symbol]\nSupported: btc, eth, bsc\nExample: whales btc", nil
	}

	symbol := strings.ToLower(args[0])
	if symbol == "bnb" {
		symbol = "bsc"
	}

	const limit = 10

	transfers, err := cl.Whale.GetLargeTransfers(ctx, symbol, nil, limit)
	if err != nil {
		return fmt.Sprintf("Failed to fetch whale data for %s: %v", strings.ToUpper(symbol), err), nil
	}

	if len(transfers) == 0 {
		return fmt.Sprintf("No recent whale transfers found for %s above $100,000.", strings.ToUpper(symbol)), nil
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Whale activity for %s (value ≥ $100,000):\n\n", strings.ToUpper(symbol)))

	for i, t := range transfers {
		// pendekin address kalau ada
		from := t.From
		if len(from) > 14 {
			from = from[:6] + "…" + from[len(from)-4:]
		}
		to := t.To
		if len(to) > 14 {
			to = to[:6] + "…" + to[len(to)-4:]
		}

		b.WriteString(fmt.Sprintf(
			"%d) [Chain: %s]\n   From: %s\n   To:   %s\n   Amount: %.4f %s (~$%.0f)\n   Time: %s\n",
			i+1,
			strings.Title(t.Chain),
			from,
			to,
			t.Amount,
			strings.ToUpper(t.Symbol),
			t.AmountUSD,
			t.Timestamp.UTC().Format(time.RFC3339),
		))

		if t.TxURL != "" {
			b.WriteString(fmt.Sprintf("   Tx: %s\n", t.TxURL))
		}
		b.WriteString("\n")
	}

	return b.String(), nil
}
