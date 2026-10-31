package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/Shreyansh1032/payments-ledger-system/internal/model"
)

var (
	ErrGoogleTokenInvalid     = errors.New("invalid google token")
	ErrGoogleEmailNotVerified = errors.New("google account email is not verified")
)

type GoogleClaims struct {
	Subject       string
	Email         string
	EmailVerified bool
	GivenName     string
	FamilyName    string
}

type GoogleTokenVerifier interface {
	Verify(ctx context.Context, idToken string) (*GoogleClaims, error)
}

type GoogleAuthService struct {
	pool     *pgxpool.Pool
	verifier GoogleTokenVerifier
}

func NewGoogleAuthService(pool *pgxpool.Pool, verifier GoogleTokenVerifier) *GoogleAuthService {
	return &GoogleAuthService{pool: pool, verifier: verifier}
}

var usernameSanitizer = regexp.MustCompile(`[^a-z0-9]`)

func (s *GoogleAuthService) generateUsername(ctx context.Context, email string) (string, error) {
	local := strings.ToLower(strings.SplitN(email, "@", 2)[0])
	base := usernameSanitizer.ReplaceAllString(local, "")
	if base == "" {
		base = "user"
	}

	candidate := base
	for i := 0; ; i++ {
		if i > 0 {
			candidate = base + strconv.Itoa(i)
		}
		var exists bool
		if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE username=$1)`, candidate).Scan(&exists); err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
}

func randomPasswordHash() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(hex.EncodeToString(b)), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s *GoogleAuthService) SignIn(ctx context.Context, idToken string) (*model.User, bool, error) {
	claims, err := s.verifier.Verify(ctx, idToken)
	if err != nil {
		return nil, false, ErrGoogleTokenInvalid
	}
	if !claims.EmailVerified {
		return nil, false, ErrGoogleEmailNotVerified
	}

	var u model.User
	err = s.pool.QueryRow(ctx,
		`SELECT id, username, first_name, last_name, created_at FROM users WHERE google_id=$1`,
		claims.Subject,
	).Scan(&u.ID, &u.Username, &u.FirstName, &u.LastName, &u.CreatedAt)
	if err == nil {
		return &u, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, false, err
	}

	err = s.pool.QueryRow(ctx,
		`SELECT id, username, first_name, last_name, created_at FROM users WHERE email=$1 AND email_verified=true`,
		claims.Email,
	).Scan(&u.ID, &u.Username, &u.FirstName, &u.LastName, &u.CreatedAt)
	if err == nil {
		if _, err := s.pool.Exec(ctx, `UPDATE users SET google_id=$1 WHERE id=$2`, claims.Subject, u.ID); err != nil {
			return nil, false, err
		}
		return &u, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, false, err
	}

	username, err := s.generateUsername(ctx, claims.Email)
	if err != nil {
		return nil, false, err
	}
	passwordHash, err := randomPasswordHash()
	if err != nil {
		return nil, false, err
	}

	firstName, lastName := claims.GivenName, claims.FamilyName
	if firstName == "" {
		firstName = username
	}
	if lastName == "" {
		lastName = "-"
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, false, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx,
		`INSERT INTO users (username, first_name, last_name, password_hash, email, email_verified, google_id)
		 VALUES ($1, $2, $3, $4, $5, true, $6)
		 RETURNING id, username, first_name, last_name, created_at`,
		username, firstName, lastName, passwordHash, claims.Email, claims.Subject,
	).Scan(&u.ID, &u.Username, &u.FirstName, &u.LastName, &u.CreatedAt)
	if err != nil {
		return nil, false, err
	}

	if _, err := tx.Exec(ctx, `INSERT INTO accounts (user_id, balance) VALUES ($1, 0)`, u.ID); err != nil {
		return nil, false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, false, err
	}

	return &u, true, nil
}
