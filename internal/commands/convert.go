package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"teneo-agent-demo1/internal/clients"
	"teneo-agent-demo1/internal/services"
)

func CmdConvert(ctx context.Context, cl *clients.Clients, args []string) (string, error) {
	if len(args) < 3 {
		return "Usage: convert [amount] [from] [to]\nExample: convert 1 btc eth", nil
	}
	amount, err := strconv.ParseFloat(args[0], 64)
	if err != nil || amount <= 0 {
		return "Invalid amount.", nil
	}
	fromSym := args[1]
	toSym := args[2]

	fromID, err := cl.ResolveCoinID(ctx, fromSym)
	if err != nil {
		return fmt.Sprintf("Could not resolve asset '%s': %v", fromSym, err), nil
	}
	toID, err := cl.ResolveCoinID(ctx, toSym)
	if err != nil {
		return fmt.Sprintf("Could not resolve asset '%s': %v", toSym, err), nil
	}

	rate, err := cl.FetchConversionRate(ctx, fromID, toID)
	if err != nil {
		return fmt.Sprintf("Failed to get conversion rate %s→%s: %v", fromSym, toSym, err), nil
	}

	converted := amount * rate
	out := fmt.Sprintf(
		"%s %s ≈ %s %s\nRate: 1 %s ≈ %s %s",
		services.F(amount), strings.ToUpper(fromSym),
		services.F(converted), strings.ToUpper(toSym),
		strings.ToUpper(fromSym), services.F(rate), strings.ToUpper(toSym),
	)
	return out, nil
}
