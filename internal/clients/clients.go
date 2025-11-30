package clients

import (
	"net/http"
	"os"
)

type Clients struct {
	HTTP *http.Client

	// Existing
	CMCAPIKey string

	// 🔥 Newly added API keys
	BlockchairAPIKey string // untuk whale ETH & BSC (blockchair.com)
	// Tambahan ke depan bisa disini:
	// DebankAPIKey    string
	// OwlracleAPIKey  string
	// HeliusAPIKey    string
}

func NewClients(httpClient *http.Client) *Clients {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	return &Clients{
		HTTP: httpClient,

		// Existing
		CMCAPIKey: os.Getenv("CMC_API_KEY"),

		// New
		BlockchairAPIKey: os.Getenv("BLOCKCHAIR_API_KEY"),
	}
}
