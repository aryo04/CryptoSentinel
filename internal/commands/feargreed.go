package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"teneo-agent-demo1/internal/clients"
)

// CmdFearGreed mengambil Crypto Fear & Greed Index dari alternative.me
// Usage: feargreed
func CmdFearGreed(ctx context.Context, cl *clients.Clients, _ []string) (string, error) {
	const endpoint = "https://api.alternative.me/fng/?limit=1"

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

	d := data.Data[0]
	return fmt.Sprintf(
		"Crypto Fear & Greed Index:\nValue: %s/100\nSentiment: %s\nSource: alternative.me",
		d.Value,
		d.ValueClassification,
	), nil
}
