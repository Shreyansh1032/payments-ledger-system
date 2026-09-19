package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")

const refreshTokenTTL = 7 * 24 * time.Hour

type RefreshTokenService struct {
	pool *pgxpool.Pool
}

func NewRefreshTokenService(pool *pgxpool.Pool) *RefreshTokenService {
	return &RefreshTokenService{pool: pool}
}

func generateRandomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *RefreshTokenService) Issue(ctx context.Context, userID string) (string, error) {
	token, err := generateRandomToken()
	if err != nil {
		return "", err
	}

	_, err = s.pool.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, hashToken(token), time.Now().Add(refreshTokenTTL),
	)
	if err != nil {
		return "", err
	}

	return token, nil
}

// Rotate validates a refresh token, revokes it, and issues a new one --
// old token can never be used again once this succeeds.
func (s *RefreshTokenService) Rotate(ctx context.Context, token string) (userID, username, newRefreshToken string, err error) {
	hash := hashToken(token)

	var id string
	var expiresAt time.Time
	var revokedAt *time.Time

	err = s.pool.QueryRow(ctx,
		`SELECT rt.id, rt.user_id, rt.expires_at, rt.revoked_at, u.username
		 FROM refresh_tokens rt JOIN users u ON u.id = rt.user_id
		 WHERE rt.token_hash=$1`,
		hash,
	).Scan(&id, &userID, &expiresAt, &revokedAt, &username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", "", ErrInvalidRefreshToken
		}
		return "", "", "", err
	}

	if revokedAt != nil || time.Now().After(expiresAt) {
		return "", "", "", ErrInvalidRefreshToken
	}

	if _, err = s.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked_at = now() WHERE id=$1`, id); err != nil {
		return "", "", "", err
	}

	newRefreshToken, err = s.Issue(ctx, userID)
	if err != nil {
		return "", "", "", err
	}

	return userID, username, newRefreshToken, nil
}

func (s *RefreshTokenService) Revoke(ctx context.Context, token string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at = now() WHERE token_hash=$1 AND revoked_at IS NULL`,
		hashToken(token),
	)
	return err
}
