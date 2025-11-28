package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"teneo-agent-demo1/internal/clients"
	"teneo-agent-demo1/internal/services"
)

// CmdConvert melakukan konversi antara crypto & fiat
// Usage:
//   convert [amount] [from] [to]
//   convert 2 sol usdt
//   convert 1000 usd eth
func CmdConvert(ctx context.Context, cl *clients.Clients, args []string) (string, error) {
	if len(args) < 3 {
		return "Usage: convert [amount] [from] [to]\nExamples:\n  convert 0.5 btc usd\n  convert 1000 usd eth\n  convert 1 eth btc", nil
	}

	amountStr, from, to := args[0], args[1], args[2]

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		return fmt.Sprintf("Invalid amount: %s", amountStr), nil
	}

	fromLower := strings.ToLower(from)
	toLower := strings.ToLower(to)

	fromIsFiat := isFiat(fromLower)
	toIsFiat := isFiat(toLower)

	// ============ crypto → crypto ============
	if !fromIsFiat && !toIsFiat {
		fromID, err := resolveCoinID(ctx, cl, fromLower)
		if err != nil {
			return fmt.Sprintf("Could not resolve asset '%s': %v", from, err), nil
		}
		toID, err := resolveCoinID(ctx, cl, toLower)
		if err != nil {
			return fmt.Sprintf("Could not resolve asset '%s': %v", to, err), nil
		}

		fromUSD, err := fetchUSDPrice(ctx, cl, fromID)
		if err != nil {
			return fmt.Sprintf("Failed to fetch USD price for %s: %v", from, err), nil
		}
		toUSD, err := fetchUSDPrice(ctx, cl, toID)
		if err != nil {
			return fmt.Sprintf("Failed to fetch USD price for %s: %v", to, err), nil
		}
		if toUSD <= 0 {
			return fmt.Sprintf("Invalid reference price for %s", to), nil
		}

		result := amount * fromUSD / toUSD

		return fmt.Sprintf(
			"Conversion (crypto → crypto):\n%.8f %s ≈ %.8f %s\n1 %s ≈ %.4f %s\n1 %s ≈ %.4f %s\nSource: CoinGecko",
			amount, strings.ToUpper(from),
			result, strings.ToUpper(to),
			strings.ToUpper(from), fromUSD/toUSD, strings.ToUpper(to),
			strings.ToUpper(to), toUSD/fromUSD, strings.ToUpper(from),
		), nil
	}

	// ============ crypto → fiat ============
	if !fromIsFiat && toIsFiat {
		fromID, err := resolveCoinID(ctx, cl, fromLower)
		if err != nil {
			return fmt.Sprintf("Could not resolve asset '%s': %v", from, err), nil
		}
		price, err := fetchSimplePrice(ctx, cl, fromID, toLower)
		if err != nil {
			return fmt.Sprintf("Failed to fetch rate %s→%s: %v", from, to, err), nil
		}
		result := amount * price
		return fmt.Sprintf(
			"Conversion (crypto → fiat):\n%.8f %s ≈ %s %s\nRate: 1 %s ≈ %s %s\nSource: CoinGecko",
			amount, strings.ToUpper(from),
			services.F(result), strings.ToUpper(to),
			strings.ToUpper(from), services.F(price), strings.ToUpper(to),
		), nil
	}

	// ============ fiat → crypto ============
	if fromIsFiat && !toIsFiat {
		toID, err := resolveCoinID(ctx, cl, toLower)
		if err != nil {
			return fmt.Sprintf("Could not resolve asset '%s': %v", to, err), nil
		}
		price, err := fetchSimplePrice(ctx, cl, toID, fromLower)
		if err != nil {
			return fmt.Sprintf("Failed to fetch rate %s→%s: %v", from, to, err), nil
		}
		if price <= 0 {
			return "Rate is zero or invalid.", nil
		}
		result := amount / price
		return fmt.Sprintf(
			"Conversion (fiat → crypto):\n%s %s ≈ %.8f %s\nRate: 1 %s ≈ %s %s\nSource: CoinGecko",
			services.F(amount), strings.ToUpper(from),
			result, strings.ToUpper(to),
			strings.ToUpper(to), services.F(price), strings.ToUpper(from),
		), nil
	}

	// fiat → fiat tidak didukung
	return "Fiat-to-fiat conversion is not supported in this version. Try converting via a crypto asset, e.g.: convert 100 usd btc", nil
}

// ================== helper buat convert ==================

func isFiat(code string) bool {
	switch strings.ToLower(code) {
	case "usd", "eur", "idr", "jpy", "gbp", "aud", "cad", "chf", "cny", "inr", "sgd", "myr", "php", "krw", "brl":
		return true
	default:
		return false
	}
}

func resolveCoinID(ctx context.Context, cl *clients.Clients, query string) (string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", fmt.Errorf("empty symbol")
	}
	endpoint := fmt.Sprintf("https://api.coingecko.com/api/v3/search?query=%s", url.QueryEscape(query))

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")

	resp, err := cl.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("coingecko search status: %d", resp.StatusCode)
	}

	var sr struct {
		Coins []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Symbol string `json:"symbol"`
		} `json:"coins"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return "", err
	}
	if len(sr.Coins) == 0 {
		return "", fmt.Errorf("no match on CoinGecko")
	}

	q := strings.ToLower(query)
	for _, c := range sr.Coins {
		if strings.ToLower(c.Symbol) == q {
			return c.ID, nil
		}
	}
	return sr.Coins[0].ID, nil
}

func fetchSimplePrice(ctx context.Context, cl *clients.Clients, id, vs string) (float64, error) {
	vs = strings.ToLower(vs)
	endpoint := fmt.Sprintf(
		"https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=%s",
		url.QueryEscape(id),
		url.QueryEscape(vs),
	)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	req.Header.Set("User-Agent", "CryptoSentinelAI/1.0")

	resp, err := cl.HTTP.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("simple price status: %d", resp.StatusCode)
	}

	var m map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return 0, err
	}
	row, ok := m[id]
	if !ok {
		return 0, fmt.Errorf("no price for id %s", id)
	}
	price, ok := row[vs]
	if !ok {
		return 0, fmt.Errorf("no %s price for id %s", vs, id)
	}
	return price, nil
}

func fetchUSDPrice(ctx context.Context, cl *clients.Clients, id string) (float64, error) {
	return fetchSimplePrice(ctx, cl, id, "usd")
}
