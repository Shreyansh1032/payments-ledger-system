package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrTransactionNotFound = errors.New("transaction not found")

type TransactionCategoryService struct {
	pool *pgxpool.Pool
}

func NewTransactionCategoryService(pool *pgxpool.Pool) *TransactionCategoryService {
	return &TransactionCategoryService{pool: pool}
}

// SetCategory lets either party to a transaction label it -- the sender
// categorizing what they spent on, or the receiver categorizing income.
func (s *TransactionCategoryService) SetCategory(ctx context.Context, userID, transactionID, category string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE transactions t
		SET category = $1
		FROM accounts fa, accounts ta
		WHERE t.id = $2
		  AND fa.id = t.from_account_id AND ta.id = t.to_account_id
		  AND (fa.user_id = $3 OR ta.user_id = $3)
	`, category, transactionID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrTransactionNotFound
	}
	return nil
}
