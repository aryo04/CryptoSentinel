package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var newsHTTPClient = &http.Client{Timeout: 15 * time.Second}

// Response minimal dari NewsData.io
type newsdataResponse struct {
	Status       string `json:"status"`
	TotalResults int    `json:"totalResults"`
	Results      []struct {
		Title       string `json:"title"`
		Link        string `json:"link"`
		Description string `json:"description"`
		SourceID    string `json:"source_id"`
		PubDate     string `json:"pubDate"`
	} `json:"results"`
}

// CmdNews
// - "news"        -> crypto news umum
// - "news solana" -> news dengan query "solana"
func CmdNews(ctx context.Context, args []string) (string, error) {
	apiKey := os.Getenv("NEWSDATA_API_KEY")
	if apiKey == "" {
		// Jangan bikin agent crash, kasih pesan ramah saja
		return "News service is not configured (missing NEWSDATA_API_KEY).", nil
	}

	// Query default: crypto market
	query := "crypto OR bitcoin OR ethereum"
	if len(args) > 0 {
		query = strings.TrimSpace(strings.Join(args, " "))
		if query == "" {
			query = "crypto OR bitcoin OR ethereum"
		}
	}

	params := url.Values{}
	params.Set("apikey", apiKey)
	params.Set("q", query)
	params.Set("language", "en")
	// kategori bisa disesuaikan, tapi business + crypto biasanya cukup
	params.Set("category", "business,technology")

	endpoint := "https://newsdata.io/api/1/news?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")

	resp, err := newsHTTPClient.Do(req)
	if err != nil {
		return "❌ Failed to reach NewsData.io. Please try again later.", nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("❌ News feed is currently unavailable (status %d). Please try again later.", resp.StatusCode), nil
	}

	var data newsdataResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "❌ Failed to decode news data from NewsData.io.", nil
	}
	if len(data.Results) == 0 {
		return "No recent crypto news found for your query.", nil
	}

	var b strings.Builder
	if len(args) == 0 {
		b.WriteString("📰 Crypto market headlines (NewsData.io)\n\n")
	} else {
		b.WriteString(fmt.Sprintf("📰 News for \"%s\" (NewsData.io)\n\n", query))
	}

	max := 5
	if len(data.Results) < max {
		max = len(data.Results)
	}

	for i := 0; i < max; i++ {
		r := data.Results[i]

		title := strings.TrimSpace(r.Title)
		if title == "" {
			title = "(no title)"
		}
		src := strings.TrimSpace(r.SourceID)
		if src == "" {
			src = "unknown source"
		}
		desc := strings.TrimSpace(r.Description)
		if desc != "" {
			// batasi biar ga kepanjangan
			if len(desc) > 220 {
				desc = desc[:217] + "..."
			}
		}

		fmt.Fprintf(&b, "%d) %s\n", i+1, title)
		fmt.Fprintf(&b, "   Source: %s\n", src)
		if desc != "" {
			fmt.Fprintf(&b, "   %s\n", desc)
		}
		if r.PubDate != "" {
			fmt.Fprintf(&b, "   Time: %s\n", r.PubDate)
		}
		if r.Link != "" {
			fmt.Fprintf(&b, "   Link: %s\n", r.Link)
		}
		b.WriteString("\n")
	}

	return b.String(), nil
}
