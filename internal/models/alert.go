package models

import "time"

type PriceAlert struct {
	Symbol    string
	Op        string
	Value     float64
	AddedAt   time.Time
	Triggered bool
}
