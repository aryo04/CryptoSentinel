package models

type CoinMarket struct {
	ID                string
	Symbol            string
	Name              string
	CurrentPrice      float64
	MarketCap         float64
	TotalVolume       float64
	PriceChangePct24h float64
	PriceChangePct7d  float64
}

type CexGainer struct {
	Exchange  string
	Symbol    string
	Price     float64
	ChangePct float64
}
