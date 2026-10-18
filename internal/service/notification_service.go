package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Shreyansh1032/payments-ledger-system/internal/model"
)

type NotificationService struct {
	pool *pgxpool.Pool
}

func NewNotificationService(pool *pgxpool.Pool) *NotificationService {
	return &NotificationService{pool: pool}
}

func (s *NotificationService) List(ctx context.Context, userID string) ([]model.Notification, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, type, message, read, created_at
		FROM notifications WHERE user_id = $1
		ORDER BY created_at DESC LIMIT 50
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := []model.Notification{}
	for rows.Next() {
		var n model.Notification
		if err := rows.Scan(&n.ID, &n.Type, &n.Message, &n.Read, &n.CreatedAt); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, rows.Err()
}

func (s *NotificationService) UnreadCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id=$1 AND read=false`, userID).Scan(&count)
	return count, err
}

func (s *NotificationService) MarkAllRead(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx, `UPDATE notifications SET read=true WHERE user_id=$1 AND read=false`, userID)
	return err
}
