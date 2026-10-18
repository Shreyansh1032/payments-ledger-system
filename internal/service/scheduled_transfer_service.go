package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/Shreyansh1032/payments-ledger-system/internal/model"
)

var (
	ErrInvalidFrequency          = errors.New("frequency must be daily, weekly, or monthly")
	ErrScheduledTransferNotFound = errors.New("scheduled transfer not found")
)

type ScheduledTransferService struct {
	pool            *pgxpool.Pool
	transferService *TransferService
}

func NewScheduledTransferService(pool *pgxpool.Pool, transferService *TransferService) *ScheduledTransferService {
	return &ScheduledTransferService{pool: pool, transferService: transferService}
}

func nextOccurrence(frequency string, from time.Time) time.Time {
	switch frequency {
	case "daily":
		return from.AddDate(0, 0, 1)
	case "weekly":
		return from.AddDate(0, 0, 7)
	default: // monthly
		return from.AddDate(0, 1, 0)
	}
}

func (s *ScheduledTransferService) Create(ctx context.Context, userID, toUsername, frequency string, amount decimal.Decimal) (*model.ScheduledTransfer, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalidAmount
	}
	if frequency != "daily" && frequency != "weekly" && frequency != "monthly" {
		return nil, ErrInvalidFrequency
	}

	var exists bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE username=$1)`, toUsername).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrRecipientNotFound
	}

	st := &model.ScheduledTransfer{ToUsername: toUsername, Amount: amount, Frequency: frequency, Active: true}
	nextRun := nextOccurrence(frequency, time.Now())
	err := s.pool.QueryRow(ctx, `
		INSERT INTO scheduled_transfers (user_id, to_username, amount, frequency, next_run_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, next_run_at, created_at
	`, userID, toUsername, amount, frequency, nextRun).Scan(&st.ID, &st.NextRunAt, &st.CreatedAt)
	if err != nil {
		return nil, err
	}
	return st, nil
}

func (s *ScheduledTransferService) List(ctx context.Context, userID string) ([]model.ScheduledTransfer, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, to_username, amount, frequency, next_run_at, active, last_run_at, last_run_status, created_at
		FROM scheduled_transfers
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []model.ScheduledTransfer{}
	for rows.Next() {
		var st model.ScheduledTransfer
		if err := rows.Scan(&st.ID, &st.ToUsername, &st.Amount, &st.Frequency, &st.NextRunAt, &st.Active, &st.LastRunAt, &st.LastRunStatus, &st.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, st)
	}
	return results, rows.Err()
}

func (s *ScheduledTransferService) Cancel(ctx context.Context, userID, id string) error {
	tag, err := s.pool.Exec(ctx, `UPDATE scheduled_transfers SET active = false WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrScheduledTransferNotFound
	}
	return nil
}

// RunDue executes every scheduled transfer whose next_run_at has passed.
// Each run's idempotency key is derived from the schedule id and the exact
// due timestamp, so re-running this (e.g. after a crash) never double-charges --
// it reuses the same idempotent Transfer path everything else goes through.
// On failure the schedule still advances to its next period rather than
// retrying every tick, so one insufficient-balance month doesn't spam retries.
func (s *ScheduledTransferService) RunDue(ctx context.Context) error {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, to_username, amount, frequency, next_run_at
		FROM scheduled_transfers
		WHERE active = true AND next_run_at <= now()
	`)
	if err != nil {
		return err
	}

	type due struct {
		id, userID, toUsername, frequency string
		amount                            decimal.Decimal
		nextRunAt                         time.Time
	}
	var dueList []due
	for rows.Next() {
		var d due
		if err := rows.Scan(&d.id, &d.userID, &d.toUsername, &d.amount, &d.frequency, &d.nextRunAt); err != nil {
			rows.Close()
			return err
		}
		dueList = append(dueList, d)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	for _, d := range dueList {
		idempotencyKey := fmt.Sprintf("scheduled-%s-%d", d.id, d.nextRunAt.Unix())
		_, transferErr := s.transferService.Transfer(ctx, idempotencyKey, d.userID, d.toUsername, d.amount)

		status := "completed"
		if transferErr != nil {
			status = "failed"
		}
		newNextRun := nextOccurrence(d.frequency, d.nextRunAt)
		if _, err := s.pool.Exec(ctx, `
			UPDATE scheduled_transfers
			SET next_run_at = $1, last_run_at = now(), last_run_status = $2
			WHERE id = $3
		`, newNextRun, status, d.id); err != nil {
			return err
		}
	}

	return nil
}
