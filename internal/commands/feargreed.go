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

// CmdFearGreed mengambil Crypto Fear & Greed Index dari alternative.me
// Usage: feargreed
func CmdFearGreed(ctx context.Context, cl *clients.Clients, _ []string) (string, error) {
	// Ambil 2 poin data (sekarang + 24h sebelumnya) supaya bisa tampilkan perubahan
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

	// parse timestamp → waktu manusiawi (UTC)
	var lastUpdate string
	if ts, err := strconv.ParseInt(current.Timestamp, 10, 64); err == nil {
		lastUpdate = time.Unix(ts, 0).UTC().Format("2006-01-02 15:04 MST")
	} else {
		lastUpdate = current.Timestamp
	}

	// hitung perubahan vs 24h lalu (kalau datanya ada)
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

	// sedikit note kontekstual biar lebih “berguna”
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
	b.WriteString("• 0–24   : Extreme Fear (kapitulasi / panic selling)\n")
	b.WriteString("• 25–49  : Fear (sentimen masih negatif)\n")
	b.WriteString("• 50–54  : Neutral (market seimbang)\n")
	b.WriteString("• 55–74  : Greed (optimisme mulai tinggi)\n")
	b.WriteString("• 75–100 : Extreme Greed (market panas, rawan koreksi)\n\n")
	b.WriteString("Source: alternative.me")

	return b.String(), nil
}

// explainFearGreed memberi penjelasan singkat berdasarkan nilai index.
func explainFearGreed(value int, classification string) string {
	class := strings.ToLower(classification)

	switch {
	case strings.Contains(class, "extreme fear") || value <= 24:
		return "Interpretation: market sedang dalam zona *Extreme Fear*. Biasanya banyak pelaku pasar yang takut dan menjual di harga rendah. Ini bisa jadi area akumulasi jangka panjang, tapi risiko masih tinggi."
	case strings.Contains(class, "fear") || (value >= 25 && value <= 49):
		return "Interpretation: market berada di zona *Fear*. Sentimen masih negatif, volatilitas bisa cukup tinggi. Trader agresif kadang mulai mencari entry di fase ini."
	case strings.Contains(class, "neutral") || (value >= 50 && value <= 54):
		return "Interpretation: market berada di zona *Neutral*. Tidak terlalu bullish atau bearish. Cocok untuk observasi tren berikutnya dan menyiapkan rencana."
	case strings.Contains(class, "greed") && value < 75:
		return "Interpretation: market berada di zona *Greed*. Optimisme meningkat, harga sudah naik lumayan. Perlu disiplin risk management dan hindari FOMO."
	case strings.Contains(class, "extreme greed") || value >= 75:
		return "Interpretation: market berada di zona *Extreme Greed*. Euforia tinggi dan market cenderung overbought. Waspada potensi koreksi tajam."
	default:
		return ""
	}
}
