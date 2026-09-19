package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type TransactionHistoryEntry struct {
	ID            string          `json:"id"`
	Direction     string          `json:"direction"` // "sent" or "received"
	Counterparty  string          `json:"counterparty"`
	Amount        decimal.Decimal `json:"amount"`
	Status        string          `json:"status"`
	FailureReason *string         `json:"failure_reason,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}
