package model

import "github.com/shopspring/decimal"

type Budget struct {
	ID           string          `json:"id"`
	Category     string          `json:"category"`
	MonthlyLimit decimal.Decimal `json:"monthly_limit"`
	Spent        decimal.Decimal `json:"spent"`
}
