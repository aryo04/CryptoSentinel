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

// CmdNews mengambil “status updates” dari CoinGecko sebagai pseudo-news.
// Tidak pakai API key, tapi endpoint kadang berubah / rate-limited.
func CmdNews(ctx context.Context, _ []string) (string, error) {
	// Versi tanpa filter category supaya lebih kompatibel
	url := "https://api.coingecko.com/api/v3/status_updates?per_page=6&page=1"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "❌ Failed to build news request.", nil
	}
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")

	resp, err := newsHTTPClient.Do(req)
	if err != nil {
		return fmt.Sprintf("❌ Failed to fetch news: %v", err), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Jangan lempar error ke SDK, cukup tampilkan ke user
		return fmt.Sprintf("❌ News feed is currently unavailable (status %d). Please try again later.", resp.StatusCode), nil
	}

	var data cgStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return fmt.Sprintf("❌ Failed to decode news response: %v", err), nil
	}
	if len(data.StatusUpdates) == 0 {
		return "No fresh updates from CoinGecko right now.", nil
	}

	var b strings.Builder
	b.WriteString("📰 Latest crypto market updates (CoinGecko)\n\n")

	max := 5
	if len(data.StatusUpdates) < max {
		max = len(data.StatusUpdates)
	}

	for i := 0; i < max; i++ {
		u := data.StatusUpdates[i]
		line := fmt.Sprintf(
			"%d) [%s / %s]\n   %s\n   Time: %s\n\n",
			i+1,
			u.Project.Name,
			strings.ToUpper(u.Project.Symbol),
			strings.TrimSpace(u.Description),
			u.CreatedAt, // biasanya ISO string
		)
		b.WriteString(line)
	}

	return b.String(), nil
}
