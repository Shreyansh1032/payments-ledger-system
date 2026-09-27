package service

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/Shreyansh1032/payments-ledger-system/internal/model"
)

type InsightsService struct {
	pool *pgxpool.Pool
}

func NewInsightsService(pool *pgxpool.Pool) *InsightsService {
	return &InsightsService{pool: pool}
}

func (s *InsightsService) GetMonthlyInsights(ctx context.Context, userID string) (*model.MonthlyInsights, error) {
	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	insights := &model.MonthlyInsights{
		TotalSent:     decimal.Zero,
		TotalReceived: decimal.Zero,
		TopRecipients: []model.RecipientTotal{},
	}

	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(t.amount), 0)
		FROM transactions t
		JOIN accounts a ON a.id = t.from_account_id
		WHERE a.user_id = $1 AND t.status = 'completed' AND t.created_at >= $2
	`, userID, monthStart).Scan(&insights.TotalSent)
	if err != nil {
		return nil, err
	}

	err = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(t.amount), 0)
		FROM transactions t
		JOIN accounts a ON a.id = t.to_account_id
		WHERE a.user_id = $1 AND t.status = 'completed' AND t.created_at >= $2
	`, userID, monthStart).Scan(&insights.TotalReceived)
	if err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT u.username, SUM(t.amount) as total
		FROM transactions t
		JOIN accounts fa ON fa.id = t.from_account_id
		JOIN accounts ta ON ta.id = t.to_account_id
		JOIN users u ON u.id = ta.user_id
		WHERE fa.user_id = $1 AND t.status = 'completed' AND t.created_at >= $2
		GROUP BY u.username
		ORDER BY total DESC
		LIMIT 5
	`, userID, monthStart)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var rt model.RecipientTotal
		if err := rows.Scan(&rt.Username, &rt.Total); err != nil {
			return nil, err
		}
		insights.TopRecipients = append(insights.TopRecipients, rt)
	}

	return insights, rows.Err()
}
