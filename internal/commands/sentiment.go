package commands

import (
	"context"
	"fmt"

	"teneo-agent-demo1/internal/clients"
	"teneo-agent-demo1/internal/services"
)

// CmdSentiment analyzes short-term market sentiment for a single asset
// Usage: sentiment [symbol]
// Example: sentiment btc
func CmdSentiment(ctx context.Context, cl *clients.Clients, args []string) (string, error) {
	if len(args) == 0 {
		return "Usage: sentiment [symbol]\nExample: sentiment btc", nil
	}
	sym := args[0]

	id, err := cl.ResolveCoinID(ctx, sym)
	if err != nil {
		return fmt.Sprintf("Could not resolve asset '%s': %v", sym, err), nil
	}

	mkt, err := cl.FetchCoinMarket(ctx, id)
	if err != nil {
		return fmt.Sprintf("Failed to fetch market data for '%s': %v", sym, err), nil
	}

	// Basic metrics
	price := mkt.CurrentPrice
	ch24 := mkt.PriceChangePct24h
	ch7d := mkt.PriceChangePct7d
	vol := mkt.TotalVolume
	mcap := mkt.MarketCap

	// Volume / market cap ratio as a simple activity proxy
	volRatio := 0.0
	if mcap > 0 {
		volRatio = vol / mcap * 100.0
	}

	// Simple scoring heuristic
	bullScore := 0
	bearScore := 0

	// 24h change
	switch {
	case ch24 > 5:
		bullScore += 3
	case ch24 > 2:
		bullScore += 2
	case ch24 > 0.5:
		bullScore++
	case ch24 < -5:
		bearScore += 3
	case ch24 < -2:
		bearScore += 2
	case ch24 < -0.5:
		bearScore++
	}

	// 7d change
	switch {
	case ch7d > 15:
		bullScore += 3
	case ch7d > 8:
		bullScore += 2
	case ch7d > 3:
		bullScore++
	case ch7d < -15:
		bearScore += 3
	case ch7d < -8:
		bearScore += 2
	case ch7d < -3:
		bearScore++
	}

	// Activity
	activityLabel := "Low"
	switch {
	case volRatio > 20:
		activityLabel = "Very High"
	case volRatio > 10:
		activityLabel = "High"
	case volRatio > 3:
		activityLabel = "Moderate"
	}

	// Final mood classification
	mood := "Neutral 😐"
	summary := "Price action is relatively balanced with no clear dominance from buyers or sellers."

	if bullScore >= bearScore+2 {
		mood = "Bullish 📈"
		summary = "Buyers are in control with positive momentum over the last few sessions."
	} else if bearScore >= bullScore+2 {
		mood = "Bearish 📉"
		summary = "Sellers are dominating, with visible downside pressure in recent days."
	} else if bullScore > bearScore {
		mood = "Slightly Bullish 🙂"
		summary = "Slight bullish tilt — upward bias but not an explosive trend."
	} else if bearScore > bullScore {
		mood = "Slightly Bearish 🙁"
		summary = "Slight bearish tilt — some downside pressure but not a full breakdown."
	}

	out := fmt.Sprintf(
		"Sentiment — %s (%s)\n"+
			"Price: $%s\n"+
			"24h change: %s%%\n"+
			"7d change: %s%%\n"+
			"Volume / Market Cap: %s%% (%s activity)\n\n"+
			"Market mood: %s\n"+
			"Summary: %s\n"+
			"Source: CoinGecko",
		mkt.Name, strings.ToUpper(mkt.Symbol),
		services.F(price),
		services.F(ch24),
		services.F(ch7d),
		services.F(volRatio), activityLabel,
		mood,
		summary,
	)

	return out, nil
}
