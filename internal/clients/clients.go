package clients

import "net/http"

type Clients struct {
	HTTP      *http.Client
	CoinGecko *CoinGeckoClient
	DefiLlama *DefiLlamaClient
	CEX       *CEXClient
	Whale     *WhaleClient
}

func NewClients(httpClient *http.Client) *Clients {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	return &Clients{
		HTTP:      httpClient,
		CoinGecko: NewCoinGeckoClient(httpClient),
		DefiLlama: NewDefiLlamaClient(httpClient),
		CEX:       NewCEXClient(httpClient),
		Whale:     NewWhaleClient(httpClient), // ← telah ditambahkan sesuai permintaan
	}
}
