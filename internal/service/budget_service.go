package service

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/Shreyansh1032/payments-ledger-system/internal/model"
)

var ErrBudgetNotFound = errors.New("budget not found")

type BudgetService struct {
	pool *pgxpool.Pool
}

func NewBudgetService(pool *pgxpool.Pool) *BudgetService {
	return &BudgetService{pool: pool}
}

func (s *BudgetService) SetLimit(ctx context.Context, userID, category string, limit decimal.Decimal) error {
	if limit.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidAmount
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO budgets (user_id, category, monthly_limit)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, category) DO UPDATE SET monthly_limit = EXCLUDED.monthly_limit
	`, userID, category, limit)
	return err
}

func (s *BudgetService) List(ctx context.Context, userID string) ([]model.Budget, error) {
	monthStart := time.Now().UTC()
	monthStart = time.Date(monthStart.Year(), monthStart.Month(), 1, 0, 0, 0, 0, time.UTC)

	rows, err := s.pool.Query(ctx, `
		SELECT b.id, b.category, b.monthly_limit,
		       COALESCE((
		           SELECT SUM(t.amount)
		           FROM transactions t
		           JOIN accounts a ON a.id = t.from_account_id
		           WHERE a.user_id = $1 AND t.category = b.category
		             AND t.status = 'completed' AND t.created_at >= $2
		       ), 0) AS spent
		FROM budgets b
		WHERE b.user_id = $1
		ORDER BY b.category
	`, userID, monthStart)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	budgets := []model.Budget{}
	for rows.Next() {
		var b model.Budget
		if err := rows.Scan(&b.ID, &b.Category, &b.MonthlyLimit, &b.Spent); err != nil {
			return nil, err
		}
		budgets = append(budgets, b)
	}
	return budgets, rows.Err()
}

func (s *BudgetService) Delete(ctx context.Context, userID, budgetID string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM budgets WHERE id = $1 AND user_id = $2`, budgetID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrBudgetNotFound
	}
	return nil
}
