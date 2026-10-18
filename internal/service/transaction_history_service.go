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

func (s *TransactionHistoryService) query(ctx context.Context, userID string, limit int) ([]model.TransactionHistoryEntry, error) {
	sql := `
		SELECT
			t.id,
			CASE WHEN fa.user_id = $1 THEN 'sent' ELSE 'received' END AS direction,
			CASE WHEN fa.user_id = $1 THEN ru.username ELSE su.username END AS counterparty,
			t.amount,
			t.category,
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
	`
	args := []interface{}{userID}
	if limit > 0 {
		sql += ` LIMIT $2`
		args = append(args, limit)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []model.TransactionHistoryEntry{}
	for rows.Next() {
		var e model.TransactionHistoryEntry
		if err := rows.Scan(&e.ID, &e.Direction, &e.Counterparty, &e.Amount, &e.Category, &e.Status, &e.FailureReason, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}

	return entries, rows.Err()
}

func (s *TransactionHistoryService) GetHistory(ctx context.Context, userID string) ([]model.TransactionHistoryEntry, error) {
	return s.query(ctx, userID, 50)
}

// GetHistoryForExport returns the full history with no row limit, for CSV export.
func (s *TransactionHistoryService) GetHistoryForExport(ctx context.Context, userID string) ([]model.TransactionHistoryEntry, error) {
	return s.query(ctx, userID, 0)
}
