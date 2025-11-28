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

// CmdNews mengambil "status_updates" dari CoinGecko.
// Kalau API lagi 404 / error, kita balas pesan ramah ke user, bukan error mentah.
func CmdNews(ctx context.Context, _ []string) (string, error) {
	url := "https://api.coingecko.com/api/v3/status_updates?category=general&per_page=6&page=1"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		// Ini betul-betul error internal (jarang terjadi)
		return "", err
	}
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")

	resp, err := newsHTTPClient.Do(req)
	if err != nil {
		return "⚠️ Failed to reach news provider. Please try again later.", nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "❌ News feed is currently unavailable (status 404). Please try again later.", nil
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("⚠️ News service returned status %d. Please try again later.", resp.StatusCode), nil
	}

	var data cgStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "⚠️ Failed to decode news response from provider.", nil
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
		desc := strings.TrimSpace(u.Description)
		if desc == "" {
			desc = "(no description)"
		}
		line := fmt.Sprintf(
			"%d) [%s / %s]\n   %s\n   Time: %s\n\n",
			i+1,
			u.Project.Name,
			strings.ToUpper(u.Project.Symbol),
			desc,
			u.CreatedAt,
		)
		b.WriteString(line)
	}

	return b.String(), nil
}
