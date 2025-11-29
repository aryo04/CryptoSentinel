package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"teneo-agent-demo1/internal/clients"
)

// CmdFearGreed fetches the Crypto Fear & Greed Index from alternative.me
// Usage: feargreed
func CmdFearGreed(ctx context.Context, cl *clients.Clients, _ []string) (string, error) {
	// Fetch 2 data points (now + 24h ago) so we can show the change
	const endpoint = "https://api.alternative.me/fng/?limit=2"

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")

	resp, err := cl.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("Fear & Greed API error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Fear & Greed status: %d", resp.StatusCode)
	}

	var data struct {
		Data []struct {
			Value               string `json:"value"`
			ValueClassification string `json:"value_classification"`
			Timestamp           string `json:"timestamp"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("decode error: %v", err)
	}
	if len(data.Data) == 0 {
		return "No Fear & Greed data available.", nil
	}

	current := data.Data[0]

	curVal, err := strconv.Atoi(current.Value)
	if err != nil {
		curVal = 0
	}

	// Parse timestamp → human readable time (UTC)
	var lastUpdate string
	if ts, err := strconv.ParseInt(current.Timestamp, 10, 64); err == nil {
		lastUpdate = time.Unix(ts, 0).UTC().Format("2006-01-02 15:04 MST")
	} else {
		lastUpdate = current.Timestamp
	}

	// Calculate change vs 24h ago (if available)
	changeLine := ""
	if len(data.Data) > 1 {
		prev := data.Data[1]
		if prevVal, err := strconv.Atoi(prev.Value); err == nil {
			diff := curVal - prevVal
			switch {
			case diff > 0:
				changeLine = fmt.Sprintf("24h change: ↑ %+d (from %d → %d)", diff, prevVal, curVal)
			case diff < 0:
				changeLine = fmt.Sprintf("24h change: ↓ %d (from %d → %d)", -diff, prevVal, curVal)
			default:
				changeLine = fmt.Sprintf("24h change: no change (still %d)", curVal)
			}
		}
	}

	// Contextual note so it's actually useful
	note := explainFearGreed(curVal, current.ValueClassification)

	var b strings.Builder
	b.WriteString("📊 *Crypto Fear & Greed Index*\n")
	b.WriteString(fmt.Sprintf("Current: %d/100 — %s\n", curVal, current.ValueClassification))
	if changeLine != "" {
		b.WriteString(changeLine + "\n")
	}
	b.WriteString(fmt.Sprintf("Last update: %s\n\n", lastUpdate))

	if note != "" {
		b.WriteString(note + "\n\n")
	}

	b.WriteString("Quick guide:\n")
	b.WriteString("• 0–24   : Extreme Fear (capitulation / panic selling)\n")
	b.WriteString("• 25–49  : Fear (market still cautious / risk-off)\n")
	b.WriteString("• 50–54  : Neutral (no clear dominance of bulls or bears)\n")
	b.WriteString("• 55–74  : Greed (optimism rising, risk of FOMO entries)\n")
	b.WriteString("• 75–100 : Extreme Greed (euphoria, market overheated and correction risk is high)\n\n")
	b.WriteString("Source: alternative.me")

	return b.String(), nil
}

// explainFearGreed returns a short interpretation line based on the index value.
func explainFearGreed(value int, classification string) string {
	class := strings.ToLower(classification)

	switch {
	case strings.Contains(class, "extreme fear") || value <= 24:
		return "Interpretation: the market is in *Extreme Fear*. Many participants are scared and selling at low prices. This often aligns with long-term accumulation zones, but volatility and downside risk are still high."
	case strings.Contains(class, "fear") || (value >= 25 && value <= 49):
		return "Interpretation: the market is in a *Fear* regime. Sentiment is negative and volatility can spike. Aggressive traders sometimes start looking for entries here."
	case strings.Contains(class, "neutral") || (value >= 50 && value <= 54):
		return "Interpretation: the market is *Neutral*. Neither bulls nor bears clearly dominate. Good phase to observe where the next major trend might build."
	case strings.Contains(class, "greed") && value < 75:
		return "Interpretation: the market is in *Greed* mode. Optimism is increasing and prices have already moved up. Stick to your risk management and avoid chasing green candles blindly."
	case strings.Contains(class, "extreme greed") || value >= 75:
		return "Interpretation: the market is in *Extreme Greed*. Euphoria is high and assets can be overbought. Great time to be disciplined, rebalance, or secure profits rather than FOMO in."
	default:
		return ""
	}
}
