package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type AccountService struct {
	pool *pgxpool.Pool
}

func NewAccountService(pool *pgxpool.Pool) *AccountService {
	return &AccountService{pool: pool}
}

func (s *AccountService) GetBalance(ctx context.Context, userID string) (decimal.Decimal, error) {
	var balance decimal.Decimal
	err := s.pool.QueryRow(ctx,
		`SELECT balance FROM accounts WHERE user_id=$1`,
		userID,
	).Scan(&balance)
	if err != nil {
		return decimal.Decimal{}, err
	}
	return balance, nil
}
