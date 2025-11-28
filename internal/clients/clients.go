package clients

import "net/http"

type Clients struct {
	HTTP *http.Client
}

func NewClients(httpClient *http.Client) *Clients {
	return &Clients{HTTP: httpClient}
}
