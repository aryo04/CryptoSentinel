package agent

import (
	"context"

	"teneo-agent-demo1/internal/commands"
	"teneo-agent-demo1/internal/clients"
	"teneo-agent-demo1/internal/services"
)

type CryptoSentinelAgent struct {
	router    *commands.Router
	alertSvc  *services.AlertService
	apiClient *clients.Clients
}

func NewCryptoSentinelAgent(apiClients *clients.Clients, alertSvc *services.AlertService) *CryptoSentinelAgent {
	router := commands.NewRouter(apiClients, alertSvc)

	return &CryptoSentinelAgent{
		router:    router,
		alertSvc:  alertSvc,
		apiClient: apiClients,
	}
}

func (a *CryptoSentinelAgent) ProcessTask(ctx context.Context, task string) (string, error) {
	return a.router.Handle(ctx, task)
}

func (a *CryptoSentinelAgent) StartBackgroundJobs(ctx context.Context) {
	a.alertSvc.StartChecker(ctx)
}
