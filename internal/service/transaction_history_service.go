package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Shreyansh1032/payments-ledger-system/internal/model"
)

type TransactionHistoryService struct {
	pool *pgxpool.Pool
}

func NewTransactionHistoryService(pool *pgxpool.Pool) *TransactionHistoryService {
	return &TransactionHistoryService{pool: pool}
}

func (s *TransactionHistoryService) GetHistory(ctx context.Context, userID string) ([]model.TransactionHistoryEntry, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			t.id,
			CASE WHEN fa.user_id = $1 THEN 'sent' ELSE 'received' END AS direction,
			CASE WHEN fa.user_id = $1 THEN ru.username ELSE su.username END AS counterparty,
			t.amount,
			t.status,
			t.failure_reason,
			t.created_at
		FROM transactions t
		JOIN accounts fa ON fa.id = t.from_account_id
		JOIN accounts ta ON ta.id = t.to_account_id
		JOIN users su ON su.id = fa.user_id
		JOIN users ru ON ru.id = ta.user_id
		WHERE fa.user_id = $1 OR ta.user_id = $1
		ORDER BY t.created_at DESC
		LIMIT 50
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []model.TransactionHistoryEntry{}
	for rows.Next() {
		var e model.TransactionHistoryEntry
		if err := rows.Scan(&e.ID, &e.Direction, &e.Counterparty, &e.Amount, &e.Status, &e.FailureReason, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}

	return entries, rows.Err()
}
