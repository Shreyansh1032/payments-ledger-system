package testutil

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
)

func NewTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	_ = godotenv.Load("../../.env")

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Fatal("TEST_DATABASE_URL not set")
	}

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxdecimal.Register(conn.TypeMap())
		return nil
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}

	t.Cleanup(pool.Close)
	return pool
}

func Truncate(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	_, err := pool.Exec(ctx,
		`TRUNCATE TABLE ledger_entries, transactions, accounts, users RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}

	// Truncation wipes the system treasury account seeded by migration
	// 000005 along with everything else -- reseed it so Deposit tests
	// (and anything else that assumes it exists) keep working.
	_, err = pool.Exec(ctx,
		`INSERT INTO users (username, first_name, last_name, password_hash)
		 VALUES ('system_treasury', 'System', 'Treasury', 'no-login-system-account')`)
	if err != nil {
		t.Fatalf("reseed system user: %v", err)
	}
	_, err = pool.Exec(ctx,
		`INSERT INTO accounts (user_id, balance, is_system)
		 SELECT id, 0, true FROM users WHERE username = 'system_treasury'`)
	if err != nil {
		t.Fatalf("reseed system account: %v", err)
	}
}
