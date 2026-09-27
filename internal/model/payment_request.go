package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type PaymentRequest struct {
	ID                string          `json:"id"`
	RequesterUsername string          `json:"requester_username,omitempty"`
	PayerUsername     string          `json:"payer_username,omitempty"`
	Amount            decimal.Decimal `json:"amount"`
	Note              string          `json:"note,omitempty"`
	Status            string          `json:"status"`
	CreatedAt         time.Time       `json:"created_at"`
}
