package clients

import (
	"net/http"
	"os"
)

type Clients struct {
	HTTP      *http.Client
	CMCAPIKey string
}

func NewClients(httpClient *http.Client) *Clients {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	return &Clients{
		HTTP:      httpClient,
		CMCAPIKey: os.Getenv("CMC_API_KEY"),
	}
}
