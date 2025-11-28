package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var newsHTTPClient = &http.Client{Timeout: 15 * time.Second}

type cgStatusResponse struct {
	StatusUpdates []struct {
		Project struct {
			Name   string `json:"name"`
			Symbol string `json:"symbol"`
		} `json:"project"`
		Description string `json:"description"`
		Category    string `json:"category"`
		CreatedAt   string `json:"created_at"`
	} `json:"status_updates"`
}

func CmdNews(ctx context.Context, _ []string) (string, error) {
	// CoinGecko status updates → "semi-news" tapi lumayan dan tanpa API key
	url := "https://api.coingecko.com/api/v3/status_updates?category=general&per_page=6&page=1"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")

	resp, err := newsHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch news: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("news provider status: %d", resp.StatusCode)
	}

	var data cgStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("decode news error: %w", err)
	}
	if len(data.StatusUpdates) == 0 {
		return "No fresh updates from CoinGecko right now.", nil
	}

	var b strings.Builder
	b.WriteString("📰 **Latest crypto market updates (CoinGecko)**\n\n")

	max := 5
	if len(data.StatusUpdates) < max {
		max = len(data.StatusUpdates)
	}

	for i := 0; i < max; i++ {
		u := data.StatusUpdates[i]
		t := u.CreatedAt
		// created_at biasanya ISO string, biar aman tulis apa adanya
		line := fmt.Sprintf(
			"%d) [%s / %s]\n   %s\n   Time: %s\n\n",
			i+1,
			u.Project.Name,
			strings.ToUpper(u.Project.Symbol),
			strings.TrimSpace(u.Description),
			t,
		)
		b.WriteString(line)
	}

	return b.String(), nil
}
