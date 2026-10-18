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
	"github.com/rs/zerolog/log"
)

var (
	ErrEmailAlreadyInUse        = errors.New("email already in use")
	ErrInvalidVerificationToken = errors.New("invalid or expired verification link")
)

const verificationTokenTTL = 24 * time.Hour

// EmailSender abstracts how a verification link actually reaches the user.
// LogEmailSender is the local-dev stand-in -- it writes the link to the
// server log instead of sending real email. A real deployment swaps this
// for an SMTP- or provider-backed implementation without touching anything
// that calls it.
type EmailSender interface {
	SendVerificationEmail(ctx context.Context, toEmail, link string) error
}

type LogEmailSender struct{}

func (LogEmailSender) SendVerificationEmail(_ context.Context, toEmail, link string) error {
	log.Info().Str("to", toEmail).Str("link", link).Msg("verification email (dev mode: logged, not sent)")
	return nil
}

type EmailVerificationService struct {
	pool   *pgxpool.Pool
	sender EmailSender
}

func NewEmailVerificationService(pool *pgxpool.Pool, sender EmailSender) *EmailVerificationService {
	return &EmailVerificationService{pool: pool, sender: sender}
}

func generateVerificationToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashVerificationToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *EmailVerificationService) SetEmail(ctx context.Context, userID, email, appBaseURL string) error {
	var existingOwner string
	err := s.pool.QueryRow(ctx, `SELECT id FROM users WHERE email=$1 AND id != $2`, email, userID).Scan(&existingOwner)
	if err == nil {
		return ErrEmailAlreadyInUse
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	if _, err := s.pool.Exec(ctx, `UPDATE users SET email=$1, email_verified=false WHERE id=$2`, email, userID); err != nil {
		return err
	}

	token, err := generateVerificationToken()
	if err != nil {
		return err
	}
	if _, err := s.pool.Exec(ctx, `
		INSERT INTO email_verifications (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, hashVerificationToken(token), time.Now().Add(verificationTokenTTL)); err != nil {
		return err
	}

	link := appBaseURL + "/verify-email?token=" + token
	return s.sender.SendVerificationEmail(ctx, email, link)
}

func (s *EmailVerificationService) Verify(ctx context.Context, token string) error {
	hash := hashVerificationToken(token)

	var id, userID string
	var expiresAt time.Time
	var usedAt *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, expires_at, used_at FROM email_verifications WHERE token_hash=$1
	`, hash).Scan(&id, &userID, &expiresAt, &usedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidVerificationToken
		}
		return err
	}
	if usedAt != nil || time.Now().After(expiresAt) {
		return ErrInvalidVerificationToken
	}

	if _, err := s.pool.Exec(ctx, `UPDATE email_verifications SET used_at=now() WHERE id=$1`, id); err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `UPDATE users SET email_verified=true WHERE id=$1`, userID)
	return err
}
