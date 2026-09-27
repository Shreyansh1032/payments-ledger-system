package service_test

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
	"github.com/Shreyansh1032/payments-ledger-system/internal/testutil"
)

func TestDeposit_IncreasesBalance(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	transferService := service.NewTransferService(pool)
	accountService := service.NewAccountService(pool)

	user, _ := authService.SignUp(ctx, "depositor1", "D", "One", "password123")

	result, err := transferService.Deposit(ctx, "deposit-key-1", user.ID, decimal.NewFromInt(500))
	if err != nil {
		t.Fatalf("deposit failed: %v", err)
	}
	if result.Status != "completed" {
		t.Errorf("expected completed, got %s", result.Status)
	}

	balance, _ := accountService.GetBalance(ctx, user.ID)
	if !balance.Equal(decimal.NewFromInt(500)) {
		t.Errorf("expected balance 500, got %s", balance.String())
	}
}

func TestDeposit_IdempotentReplay(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	transferService := service.NewTransferService(pool)
	accountService := service.NewAccountService(pool)

	user, _ := authService.SignUp(ctx, "depositor2", "D", "Two", "password123")

	key := "deposit-idempotency-key"
	r1, err := transferService.Deposit(ctx, key, user.ID, decimal.NewFromInt(200))
	if err != nil {
		t.Fatalf("first deposit failed: %v", err)
	}
	r2, err := transferService.Deposit(ctx, key, user.ID, decimal.NewFromInt(200))
	if err != nil {
		t.Fatalf("replayed deposit failed: %v", err)
	}
	if r1.TransactionID != r2.TransactionID {
		t.Errorf("replay should return same transaction id, got %s vs %s", r1.TransactionID, r2.TransactionID)
	}

	balance, _ := accountService.GetBalance(ctx, user.ID)
	if !balance.Equal(decimal.NewFromInt(200)) {
		t.Errorf("expected balance 200 after replay (must not double-deposit), got %s", balance.String())
	}
}

func TestDeposit_InvalidAmount(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	transferService := service.NewTransferService(pool)

	user, _ := authService.SignUp(ctx, "depositor3", "D", "Three", "password123")

	_, err := transferService.Deposit(ctx, "deposit-key-3", user.ID, decimal.NewFromInt(-50))
	if err != service.ErrInvalidAmount {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
}

// The core double-entry invariant: no matter how much money moves around
// (deposits, transfers), the sum of every account's balance in the system
// must always be exactly zero. The system treasury account's negative
// balance is what makes this hold after deposits create new money.
func TestDeposit_LedgerStaysBalanced(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	transferService := service.NewTransferService(pool)

	alice, _ := authService.SignUp(ctx, "ledger_alice", "A", "L", "password123")
	if _, err := authService.SignUp(ctx, "ledger_bob", "B", "L", "password123"); err != nil {
		t.Fatalf("signup bob: %v", err)
	}

	if _, err := transferService.Deposit(ctx, "ledger-deposit-1", alice.ID, decimal.NewFromInt(1000)); err != nil {
		t.Fatalf("deposit failed: %v", err)
	}
	if _, err := transferService.Transfer(ctx, "ledger-transfer-1", alice.ID, "ledger_bob", decimal.NewFromInt(300)); err != nil {
		t.Fatalf("transfer failed: %v", err)
	}

	var sum decimal.Decimal
	if err := pool.QueryRow(ctx, `SELECT COALESCE(SUM(balance), 0) FROM accounts`).Scan(&sum); err != nil {
		t.Fatalf("sum balances: %v", err)
	}
	if !sum.IsZero() {
		t.Errorf("expected all account balances to sum to zero, got %s", sum.String())
	}
}
