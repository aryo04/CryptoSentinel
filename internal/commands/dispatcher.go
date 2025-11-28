package commands

import (
	"context"
	"strings"

	"teneo-agent-demo1/internal/clients"
	"teneo-agent-demo1/internal/services"
)

type Router struct {
	clients *clients.Clients
	alerts  *services.AlertService
}

func NewRouter(cl *clients.Clients, alertSvc *services.AlertService) *Router {
	return &Router{
		clients: cl,
		alerts:  alertSvc,
	}
}

func (r *Router) Handle(ctx context.Context, task string) (string, error) {
	task = strings.TrimSpace(strings.TrimPrefix(task, "/"))
	if task == "" {
		return HelpText(), nil
	}
	parts := strings.Fields(task)
	cmd := strings.ToLower(parts[0])
	args := parts[1:]

	switch cmd {

	// Core market data
	case "price":
		return CmdPrice(ctx, r.clients, args)
	case "compare":
		return CmdCompare(ctx, r.clients, args)
	case "digest":
		return CmdDigest(ctx, r.clients, args)

	// DeFi TVL
	case "tvl":
		return CmdTVL(ctx, r.clients, args)
	case "tvlchain":
		return CmdTVLChain(ctx, r.clients, args)
	case "tvlprotocols":
		return CmdTVLProtocols(ctx, r.clients, args)

	// Market lists
	case "top":
		return CmdTop(ctx, r.clients, args)
	case "gainers":
		return CmdGainers(ctx, r.clients, args)
	case "losers":
		return CmdLosers(ctx, r.clients, args)

	// Sentiment
	case "feargreed":
		return CmdFearGreed(ctx, r.clients, args)

	// Conversions
	case "convert":
		return CmdConvert(ctx, r.clients, args)

	// CEX gainers
	case "gainers_cex":
		return CmdGainersCEX(ctx, r.clients, args)
	case "gainers_compare":
		return CmdGainersCompare(ctx, r.clients, args)

	// Gas tracker (Owlracle)
	case "gas":
		return CmdGas(ctx, args)

	// DEX price (Dexscreener)
	case "dexprice":
		return CmdDexPrice(ctx, args)

	// Portfolio (DeBank)
	case "portfolio":
		return CmdPortfolio(ctx, args)

	// Alerts
	case "alert":
		return CmdAlertAdd(r.alerts, args)
	case "alert_list":
		return CmdAlertList(r.alerts)
	case "alert_remove":
		return CmdAlertRemove(r.alerts, args)
	case "alert_clear":
		return CmdAlertClear(r.alerts)

	// News / updates
	case "news":
		return CmdNews(ctx, args)

	// Help
	case "help":
		return HelpText(), nil

	default:
		return "Unknown command. Type: help", nil
	}
}
