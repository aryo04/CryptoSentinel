package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type ProtocolTVL struct {
	Name  string
	Chain string
	TVL   float64
}

func (c *Clients) FetchProtocolTVL(ctx context.Context, protocol string) (*ProtocolTVL, error) {
	slug := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(protocol), " ", "-"))
	endpoint := fmt.Sprintf("https://api.llama.fi/protocol/%s", url.PathEscape(slug))

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("DefiLlama request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("DefiLlama status %d for '%s'", resp.StatusCode, slug)
	}

	var data struct {
		Name string `json:"name"`
		Chain string `json:"chain"`
		TVL []struct {
			Date              int64   `json:"date"`
			TotalLiquidityUSD float64 `json:"totalLiquidityUSD"`
		} `json:"tvl"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode error: %v", err)
	}
	if len(data.TVL) == 0 {
		return nil, fmt.Errorf("no TVL data for %s", protocol)
	}

	last := data.TVL[len(data.TVL)-1]
	return &ProtocolTVL{
		Name:  data.Name,
		Chain: data.Chain,
		TVL:   last.TotalLiquidityUSD,
	}, nil
}
