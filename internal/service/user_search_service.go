package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Shreyansh1032/payments-ledger-system/internal/model"
)

type UserSearchService struct {
	pool *pgxpool.Pool
}

func NewUserSearchService(pool *pgxpool.Pool) *UserSearchService {
	return &UserSearchService{pool: pool}
}

// Search excludes the requester (can't send to yourself) and the system
// treasury account (not a real recipient) from results.
func (s *UserSearchService) Search(ctx context.Context, requestingUserID, query string) ([]model.UserSearchResult, error) {
	if query == "" {
		return []model.UserSearchResult{}, nil
	}

	rows, err := s.pool.Query(ctx, `
		SELECT u.username, u.first_name, u.last_name
		FROM users u
		JOIN accounts a ON a.user_id = u.id
		WHERE a.is_system = false
		  AND u.id != $1
		  AND u.username ILIKE '%' || $2 || '%'
		ORDER BY u.username ASC
		LIMIT 10
	`, requestingUserID, query)
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
