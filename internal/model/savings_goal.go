package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type SavingsGoal struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	TargetAmount  decimal.Decimal `json:"target_amount"`
	CurrentAmount decimal.Decimal `json:"current_amount"`
	CreatedAt     time.Time       `json:"created_at"`
}
