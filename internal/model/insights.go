package model

import "github.com/shopspring/decimal"

type RecipientTotal struct {
	Username string          `json:"username"`
	Total    decimal.Decimal `json:"total"`
}

type MonthlyInsights struct {
	TotalSent     decimal.Decimal  `json:"total_sent"`
	TotalReceived decimal.Decimal  `json:"total_received"`
	TopRecipients []RecipientTotal `json:"top_recipients"`
}
