package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/Shreyansh1032/payments-ledger-system/internal/model"
)

var ErrSavingsGoalNotFound = errors.New("savings goal not found")

type SavingsGoalService struct {
	pool *pgxpool.Pool
}

func NewSavingsGoalService(pool *pgxpool.Pool) *SavingsGoalService {
	return &SavingsGoalService{pool: pool}
}

func (s *SavingsGoalService) Create(ctx context.Context, userID, name string, target decimal.Decimal) (*model.SavingsGoal, error) {
	if target.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalidAmount
	}
	g := &model.SavingsGoal{Name: name, TargetAmount: target, CurrentAmount: decimal.Zero}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO savings_goals (user_id, name, target_amount)
		VALUES ($1, $2, $3)
		RETURNING id, current_amount, created_at
	`, userID, name, target).Scan(&g.ID, &g.CurrentAmount, &g.CreatedAt)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (s *SavingsGoalService) List(ctx context.Context, userID string) ([]model.SavingsGoal, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, target_amount, current_amount, created_at
		FROM savings_goals WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	goals := []model.SavingsGoal{}
	for rows.Next() {
		var g model.SavingsGoal
		if err := rows.Scan(&g.ID, &g.Name, &g.TargetAmount, &g.CurrentAmount, &g.CreatedAt); err != nil {
			return nil, err
		}
		goals = append(goals, g)
	}
	return goals, rows.Err()
}

// AddContribution logs progress toward a goal. This does not move money out
// of the user's spendable balance -- goals track a savings target
// alongside the real ledger, they don't reserve funds inside it.
func (s *SavingsGoalService) AddContribution(ctx context.Context, userID, goalID string, amount decimal.Decimal) (*model.SavingsGoal, bool, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, false, ErrInvalidAmount
	}

	var g model.SavingsGoal
	err := s.pool.QueryRow(ctx, `
		UPDATE savings_goals
		SET current_amount = current_amount + $1
		WHERE id = $2 AND user_id = $3
		RETURNING id, name, target_amount, current_amount, created_at
	`, amount, goalID, userID).Scan(&g.ID, &g.Name, &g.TargetAmount, &g.CurrentAmount, &g.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, ErrSavingsGoalNotFound
		}
		return nil, false, err
	}

	justReached := g.CurrentAmount.GreaterThanOrEqual(g.TargetAmount) && g.CurrentAmount.Sub(amount).LessThan(g.TargetAmount)
	return &g, justReached, nil
}

func (s *SavingsGoalService) NotifyGoalReached(ctx context.Context, userID, goalName string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO notifications (user_id, type, message) VALUES ($1, 'goal_reached', $2)`,
		userID, fmt.Sprintf("You reached your savings goal: %s", goalName),
	)
	return err
}

func (s *SavingsGoalService) Delete(ctx context.Context, userID, goalID string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM savings_goals WHERE id = $1 AND user_id = $2`, goalID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrSavingsGoalNotFound
	}
	return nil
}
