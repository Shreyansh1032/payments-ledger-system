package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

var (
	ErrTOTPAlreadyEnabled = errors.New("two-factor authentication is already enabled")
	ErrInvalidTOTPCode    = errors.New("invalid code")
	ErrTOTPNotEnabled     = errors.New("two-factor authentication is not enabled")
)

type TOTPService struct {
	pool *pgxpool.Pool
}

func NewTOTPService(pool *pgxpool.Pool) *TOTPService {
	return &TOTPService{pool: pool}
}

// BeginSetup generates a new secret and stores it pending confirmation --
// 2FA isn't enabled until Confirm proves the user actually scanned it.
func (s *TOTPService) BeginSetup(ctx context.Context, userID, username string) (*otp.Key, error) {
	var enabled bool
	if err := s.pool.QueryRow(ctx, `SELECT totp_enabled FROM users WHERE id=$1`, userID).Scan(&enabled); err != nil {
		return nil, err
	}
	if enabled {
		return nil, ErrTOTPAlreadyEnabled
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "PayWave",
		AccountName: username,
	})
	if err != nil {
		return nil, err
	}

	if _, err := s.pool.Exec(ctx, `UPDATE users SET totp_secret=$1 WHERE id=$2`, key.Secret(), userID); err != nil {
		return nil, err
	}

	return key, nil
}

func (s *TOTPService) Confirm(ctx context.Context, userID, code string) error {
	var secret string
	if err := s.pool.QueryRow(ctx, `SELECT totp_secret FROM users WHERE id=$1`, userID).Scan(&secret); err != nil {
		return err
	}
	if secret == "" || !totp.Validate(code, secret) {
		return ErrInvalidTOTPCode
	}

	_, err := s.pool.Exec(ctx, `UPDATE users SET totp_enabled=true WHERE id=$1`, userID)
	return err
}

func (s *TOTPService) Disable(ctx context.Context, userID, code string) error {
	var enabled bool
	var secret string
	if err := s.pool.QueryRow(ctx, `SELECT totp_enabled, totp_secret FROM users WHERE id=$1`, userID).Scan(&enabled, &secret); err != nil {
		return err
	}
	if !enabled {
		return ErrTOTPNotEnabled
	}
	if !totp.Validate(code, secret) {
		return ErrInvalidTOTPCode
	}

	_, err := s.pool.Exec(ctx, `UPDATE users SET totp_enabled=false, totp_secret=NULL WHERE id=$1`, userID)
	return err
}

// ValidateLogin checks a code against an already-enabled user's secret --
// used by the login flow's second factor.
func (s *TOTPService) ValidateLogin(ctx context.Context, userID, code string) (bool, error) {
	var secret string
	if err := s.pool.QueryRow(ctx, `SELECT totp_secret FROM users WHERE id=$1`, userID).Scan(&secret); err != nil {
		return false, err
	}
	return totp.Validate(code, secret), nil
}

func (s *TOTPService) IsEnabled(ctx context.Context, userID string) (bool, error) {
	var enabled bool
	err := s.pool.QueryRow(ctx, `SELECT totp_enabled FROM users WHERE id=$1`, userID).Scan(&enabled)
	return enabled, err
}
