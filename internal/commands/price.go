package commands

import (
	"context"
	"fmt"
	"strings"

	"teneo-agent-demo1/internal/clients"
	"teneo-agent-demo1/internal/services"
)

func CmdPrice(ctx context.Context, cl *clients.Clients, args []string) (string, error) {
	if len(args) == 0 {
		return "Usage: price [symbol]\nExample: price btc", nil
	}
	sym := args[0]

	id, err := cl.ResolveCoinID(ctx, sym)
	if err != nil {
		return fmt.Sprintf("Could not resolve coin '%s': %v", sym, err), nil
	}
	mkt, err := cl.FetchCoinMarket(ctx, id)
	if err != nil {
		return fmt.Sprintf("Failed to fetch market data for '%s': %v", sym, err), nil
	}

	out := fmt.Sprintf("%s (%s)\nPrice: $%s\n24h change: %s%%\n7d change: %s%%\nMarket Cap: $%s\n24h Volume: $%s\nSource: CoinGecko",
		mkt.Name, strings.ToUpper(mkt.Symbol),
		services.F(mkt.CurrentPrice),
		services.F(mkt.PriceChangePct24h),
		services.F(mkt.PriceChangePct7d),
		services.IntComma(int64(mkt.MarketCap)),
		services.IntComma(int64(mkt.TotalVolume)),
	)
	return out, nil
}

func CmdCompare(ctx context.Context, cl *clients.Clients, args []string) (string, error) {
	if len(args) < 2 {
		return "Usage: compare [symbol1] [symbol2]\nExample: compare btc eth", nil
	}
	a, b := args[0], args[1]
	resA, _ := CmdPrice(ctx, cl, []string{a})
	resB, _ := CmdPrice(ctx, cl, []string{b})

	out := fmt.Sprintf(
		"Comparison: %s vs %s\n\n---- %s ----\n%s\n\n---- %s ----\n%s",
		strings.ToUpper(a), strings.ToUpper(b),
		strings.ToUpper(a), resA,
		strings.ToUpper(b), resB,
	)
	return out, nil
}
