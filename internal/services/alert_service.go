package services

import (
	"context"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"teneo-agent-demo1/internal/clients"
	"teneo-agent-demo1/internal/models"
)

type AlertService struct {
	mu      sync.Mutex
	alerts  []*models.PriceAlert
	clients *clients.Clients
}

func NewAlertService(c *clients.Clients) *AlertService {
	return &AlertService{
		clients: c,
		alerts:  []*models.PriceAlert{},
	}
}

func (s *AlertService) Add(symbol, op, valStr string) (string, error) {
	op = strings.TrimSpace(op)
	if op == "==" || op == "=" {
		return "Invalid operator. Use one of: >, <, >=, <=", nil
	}
	if op != ">" && op != "<" && op != ">=" && op != "<=" {
		return "Operator must be one of >, <, >=, <=", nil
	}

	v, err := strconv.ParseFloat(strings.TrimSpace(valStr), 64)
	if err != nil {
		return "Invalid value: " + valStr, nil
	}

	alert := &models.PriceAlert{
		Symbol: strings.ToLower(strings.TrimSpace(symbol)),
		Op:     op,
		Value:  v,
		AddedAt: time.Now(),
	}
	s.mu.Lock()
	s.alerts = append(s.alerts, alert)
	idx := len(s.alerts) - 1
	s.mu.Unlock()

	return "Alert #" + strconv.Itoa(idx) + " added: " +
		strings.ToUpper(alert.Symbol) + " " + alert.Op + " " + F(alert.Value) + " (USD)", nil
}

func (s *AlertService) List() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.alerts) == 0 {
		return "No active alerts."
	}
	var b strings.Builder
	b.WriteString("Active alerts:\n")
	for i, a := range s.alerts {
		status := "open"
		if a.Triggered {
			status = "triggered"
		}
		b.WriteString(
			strconv.Itoa(i) + ") " + strings.ToUpper(a.Symbol) +
				" " + a.Op + " " + F(a.Value) + " [" + status + "]\n",
		)
	}
	return b.String()
}

func (s *AlertService) Remove(index int) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if index < 0 || index >= len(s.alerts) {
		return "Invalid alert index."
	}
	s.alerts = append(s.alerts[:index], s.alerts[index+1:]...)
	return "Alert removed."
}

func (s *AlertService) Clear() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.alerts = []*models.PriceAlert{}
	return "All alerts cleared."
}

func (s *AlertService) StartChecker(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.checkAlerts(ctx)
			}
		}
	}()
}

func (s *AlertService) checkAlerts(ctx context.Context) {
	s.mu.Lock()
	alertsCopy := make([]*models.PriceAlert, len(s.alerts))
	copy(alertsCopy, s.alerts)
	s.mu.Unlock()

	for _, a := range alertsCopy {
		if a.Triggered {
			continue
		}
		id, err := s.clients.ResolveCoinID(ctx, a.Symbol)
		if err != nil {
			log.Printf("[alert] resolve error %s: %v", a.Symbol, err)
			continue
		}
		mkt, err := s.clients.FetchCoinMarket(ctx, id)
		if err != nil {
			log.Printf("[alert] market error %s: %v", a.Symbol, err)
			continue
		}
		price := mkt.CurrentPrice

		trigger := false
		switch a.Op {
		case ">":
			trigger = price > a.Value
		case "<":
			trigger = price < a.Value
		case ">=":
			trigger = price >= a.Value
		case "<=":
			trigger = price <= a.Value
		}

		if trigger {
			s.mu.Lock()
			a.Triggered = true
			s.mu.Unlock()
			log.Printf("[ALERT] %s %s %s (current: $%s) — triggered at %s",
				strings.ToUpper(a.Symbol), a.Op, F(a.Value), F(price), time.Now().Format(time.RFC3339))
		}
	}
}
