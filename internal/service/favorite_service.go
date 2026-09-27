package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Shreyansh1032/payments-ledger-system/internal/model"
)

var ErrCannotFavoriteSelf = errors.New("cannot favorite yourself")

type FavoriteService struct {
	pool *pgxpool.Pool
}

func NewFavoriteService(pool *pgxpool.Pool) *FavoriteService {
	return &FavoriteService{pool: pool}
}

func (s *FavoriteService) Add(ctx context.Context, userID, favoriteUsername string) error {
	var favoriteUserID string
	err := s.pool.QueryRow(ctx, `SELECT id FROM users WHERE username=$1`, favoriteUsername).Scan(&favoriteUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRecipientNotFound
		}
		return err
	}

	if favoriteUserID == userID {
		return ErrCannotFavoriteSelf
	}

	_, err = s.pool.Exec(ctx,
		`INSERT INTO favorites (user_id, favorite_user_id) VALUES ($1, $2)
		 ON CONFLICT (user_id, favorite_user_id) DO NOTHING`,
		userID, favoriteUserID,
	)
	return err
}

func (s *FavoriteService) Remove(ctx context.Context, userID, favoriteUsername string) error {
	_, err := s.pool.Exec(ctx, `
		DELETE FROM favorites
		WHERE user_id = $1
		  AND favorite_user_id = (SELECT id FROM users WHERE username = $2)
	`, userID, favoriteUsername)
	return err
}

func (s *FavoriteService) List(ctx context.Context, userID string) ([]model.UserSearchResult, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT u.username, u.first_name, u.last_name
		FROM favorites f
		JOIN users u ON u.id = f.favorite_user_id
		WHERE f.user_id = $1
		ORDER BY f.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []model.UserSearchResult{}
	for rows.Next() {
		var r model.UserSearchResult
		if err := rows.Scan(&r.Username, &r.FirstName, &r.LastName); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}
