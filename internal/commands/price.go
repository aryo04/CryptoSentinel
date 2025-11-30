package commands

import {
	"context"
	"fmt"
	"strings"

	"teneo-agent-demo1/internal/clients"
	"teneo-agent-demo1/internal/services"
)

// formatChange bikin string seperti:
//   ▲ +5.30%
//   ▼ -2.10%
//   → 0.00%
func formatChange(pct float64) string {
	arrow := "→"
	sign := ""
	if pct > 0 {
		arrow = "▲"
		sign = "+"
	} else if pct < 0 {
		arrow = "▼"
	}
	return fmt.Sprintf("%s %s%s%%", arrow, sign, services.F(pct))
}

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

	chg24 := formatChange(mkt.PriceChangePct24h)
	chg7 := formatChange(mkt.PriceChangePct7d)

	var volPct float64
	if mkt.MarketCap > 0 {
		volPct = (mkt.TotalVolume / mkt.MarketCap) * 100
	}

	var b strings.Builder
	fmt.Fprintf(&b, "📊 %s (%s)\n", mkt.Name, strings.ToUpper(mkt.Symbol))
	fmt.Fprintf(&b, "Price: $%s\n", services.F(mkt.CurrentPrice))
	fmt.Fprintf(&b, "24h: %s | 7d: %s\n", chg24, chg7)
	fmt.Fprintf(&b, "Market Cap: $%s\n", services.IntComma(int64(mkt.MarketCap)))
	fmt.Fprintf(&b, "24h Volume: $%s", services.IntComma(int64(mkt.TotalVolume)))
	if volPct > 0 {
		fmt.Fprintf(&b, " (≈%s%% of mcap)", services.F(volPct))
	}
	b.WriteString("\n")
	b.WriteString("Source: CoinGecko\n")
	b.WriteString("Tip: use `compare ")
	b.WriteString(strings.ToLower(mkt.Symbol))
	b.WriteString(" btc` or another asset to compare performance.\n")

	return b.String(), nil
}

func CmdCompare(ctx context.Context, cl *clients.Clients, args []string) (string, error) {
	if len(args) < 2 {
		return "Usage: compare [symbol1] [symbol2]\nExample: compare btc eth", nil
	}
	aSym, bSym := args[0], args[1]

	idA, err := cl.ResolveCoinID(ctx, aSym)
	if err != nil {
		return fmt.Sprintf("Could not resolve asset '%s': %v", aSym, err), nil
	}
	idB, err := cl.ResolveCoinID(ctx, bSym)
	if err != nil {
		return fmt.Sprintf("Could not resolve asset '%s': %v", bSym, err), nil
	}

	mktA, err := cl.FetchCoinMarket(ctx, idA)
	if err != nil {
		return fmt.Sprintf("Failed to fetch market data for '%s': %v", aSym, err), nil
	}
	mktB, err := cl.FetchCoinMarket(ctx, idB)
	if err != nil {
		return fmt.Sprintf("Failed to fetch market data for '%s': %v", bSym, err), nil
	}

	// Pre-calc numbers
	chg24A := formatChange(mktA.PriceChangePct24h)
	chg24B := formatChange(mktB.PriceChangePct24h)
	chg7A := formatChange(mktA.PriceChangePct7d)
	chg7B := formatChange(mktB.PriceChangePct7d)

	var volPctA, volPctB float64
	if mktA.MarketCap > 0 {
		volPctA = (mktA.TotalVolume / mktA.MarketCap) * 100
	}
	if mktB.MarketCap > 0 {
		volPctB = (mktB.TotalVolume / mktB.MarketCap) * 100
	}

	var b strings.Builder
	fmt.Fprintf(&b, "📊 Comparison: %s vs %s\n\n",
		strings.ToUpper(aSym), strings.ToUpper(bSym))

	// Asset A block
	fmt.Fprintf(&b, "%s (%s)\n",
		mktA.Name, strings.ToUpper(mktA.Symbol))
	fmt.Fprintf(&b, "  Price: $%s\n", services.F(mktA.CurrentPrice))
	fmt.Fprintf(&b, "  24h:   %s\n", chg24A)
	fmt.Fprintf(&b, "  7d:    %s\n", chg7A)
	fmt.Fprintf(&b, "  Mcap:  $%s\n", services.IntComma(int64(mktA.MarketCap)))
	fmt.Fprintf(&b, "  Vol 24h: $%s", services.IntComma(int64(mktA.TotalVolume)))
	if volPctA > 0 {
		fmt.Fprintf(&b, " (≈%s%% of mcap)", services.F(volPctA))
	}
	b.WriteString("\n\n")

	// Asset B block
	fmt.Fprintf(&b, "%s (%s)\n",
		mktB.Name, strings.ToUpper(mktB.Symbol))
	fmt.Fprintf(&b, "  Price: $%s\n", services.F(mktB.CurrentPrice))
	fmt.Fprintf(&b, "  24h:   %s\n", chg24B)
	fmt.Fprintf(&b, "  7d:    %s\n", chg7B)
	fmt.Fprintf(&b, "  Mcap:  $%s\n", services.IntComma(int64(mktB.MarketCap)))
	fmt.Fprintf(&b, "  Vol 24h: $%s", services.IntComma(int64(mktB.TotalVolume)))
	if volPctB > 0 {
		fmt.Fprintf(&b, " (≈%s%% of mcap)", services.F(volPctB))
	}
	b.WriteString("\n\n")

	// Small summary / “winner” hints
	b.WriteString("Summary:\n")

	// 24h winner
	if mktA.PriceChangePct24h > mktB.PriceChangePct24h {
		fmt.Fprintf(&b, "- 24h performance: %s outperformed %s (%s vs %s)\n",
			strings.ToUpper(mktA.Symbol), strings.ToUpper(mktB.Symbol),
			chg24A, chg24B)
	} else if mktA.PriceChangePct24h < mktB.PriceChangePct24h {
		fmt.Fprintf(&b, "- 24h performance: %s outperformed %s (%s vs %s)\n",
			strings.ToUpper(mktB.Symbol), strings.ToUpper(mktA.Symbol),
			chg24B, chg24A)
	} else {
		fmt.Fprintf(&b, "- 24h performance: similar moves (%s vs %s)\n", chg24A, chg24B)
	}

	// 7d winner
	if mktA.PriceChangePct7d > mktB.PriceChangePct7d {
		fmt.Fprintf(&b, "- 7d trend: %s led the week (%s vs %s)\n",
			strings.ToUpper(mktA.Symbol), chg7A, chg7B)
	} else if mktA.PriceChangePct7d < mktB.PriceChangePct7d {
		fmt.Fprintf(&b, "- 7d trend: %s led the week (%s vs %s)\n",
			strings.ToUpper(mktB.Symbol), chg7B, chg7A)
	} else {
		fmt.Fprintf(&b, "- 7d trend: roughly equal (%s vs %s)\n", chg7A, chg7B)
	}

	// Market cap size
	if mktA.MarketCap > mktB.MarketCap {
		fmt.Fprintf(&b, "- Size: %s has the larger market cap.\n", mktA.Name)
	} else if mktA.MarketCap < mktB.MarketCap {
		fmt.Fprintf(&b, "- Size: %s has the larger market cap.\n", mktB.Name)
	} else {
		b.WriteString("- Size: both have similar market caps.\n")
	}

	b.WriteString("\nSource: CoinGecko")

	return b.String(), nil
}
