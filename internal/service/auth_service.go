package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/Shreyansh1032/payments-ledger-system/internal/model"
)

var ErrUsernameTaken = errors.New("username already taken")
var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	pool *pgxpool.Pool
}

func NewAuthService(pool *pgxpool.Pool) *AuthService {
	return &AuthService{pool: pool}
}

func (s *AuthService) SignUp(ctx context.Context, username, firstName, lastName, password string) (*model.User, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE username=$1)`, username).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUsernameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var user model.User
	err = tx.QueryRow(ctx,
		`INSERT INTO users (username, first_name, last_name, password_hash)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, username, first_name, last_name, created_at`,
		username, firstName, lastName, string(hash),
	).Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO accounts (user_id, balance) VALUES ($1, 0)`,
		user.ID,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*model.User, error) {
	var user model.User
	err := s.pool.QueryRow(ctx,
		`SELECT id, username, first_name, last_name, password_hash, created_at
		 FROM users WHERE username=$1`,
		username,
	).Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return &user, nil
}
