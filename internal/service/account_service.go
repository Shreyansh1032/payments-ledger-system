package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/Shreyansh1032/payments-ledger-system/internal/model"
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

func (s *AccountService) GetProfile(ctx context.Context, userID string) (*model.User, error) {
	var u model.User
	err := s.pool.QueryRow(ctx,
		`SELECT id, username, first_name, last_name, created_at FROM users WHERE id=$1`,
		userID,
	).Scan(&u.ID, &u.Username, &u.FirstName, &u.LastName, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
