package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type ScheduledTransfer struct {
	ID            string          `json:"id"`
	ToUsername    string          `json:"to_username"`
	Amount        decimal.Decimal `json:"amount"`
	Frequency     string          `json:"frequency"`
	NextRunAt     time.Time       `json:"next_run_at"`
	Active        bool            `json:"active"`
	LastRunAt     *time.Time      `json:"last_run_at,omitempty"`
	LastRunStatus *string         `json:"last_run_status,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}
