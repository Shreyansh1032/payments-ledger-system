package service_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
	"github.com/Shreyansh1032/payments-ledger-system/internal/testutil"
)

func fundAccount(t *testing.T, pool *pgxpool.Pool, userID string, amount string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`UPDATE accounts SET balance = $1 WHERE user_id = $2`, amount, userID)
	if err != nil {
		t.Fatalf("fund account: %v", err)
	}
}

func TestTransferService_SuccessfulTransfer(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	transferService := service.NewTransferService(pool)
	accountService := service.NewAccountService(pool)

	sender, err := authService.SignUp(ctx, "sender1", "S", "One", "password123")
	if err != nil {
		t.Fatalf("signup sender: %v", err)
	}
	if _, err := authService.SignUp(ctx, "receiver1", "R", "One", "password123"); err != nil {
		t.Fatalf("signup receiver: %v", err)
	}

	fundAccount(t, pool, sender.ID, "1000")

	result, err := transferService.Transfer(ctx, "test-key-1", sender.ID, "receiver1", decimal.NewFromInt(100))
	if err != nil {
		t.Fatalf("transfer failed: %v", err)
	}
	if result.Status != "completed" {
		t.Errorf("expected completed, got %s", result.Status)
	}

	senderBalance, _ := accountService.GetBalance(ctx, sender.ID)
	if !senderBalance.Equal(decimal.NewFromInt(900)) {
		t.Errorf("expected sender balance 900, got %s", senderBalance.String())
	}
}

func TestTransferService_InsufficientBalance(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	transferService := service.NewTransferService(pool)

	sender, _ := authService.SignUp(ctx, "sender2", "S", "Two", "password123")
	if _, err := authService.SignUp(ctx, "receiver2", "R", "Two", "password123"); err != nil {
		t.Fatalf("signup receiver: %v", err)
	}

	fundAccount(t, pool, sender.ID, "50")

	_, err := transferService.Transfer(ctx, "test-key-2", sender.ID, "receiver2", decimal.NewFromInt(100))
	if err != service.ErrInsufficientBalance {
		t.Errorf("expected ErrInsufficientBalance, got %v", err)
	}
}

func TestTransferService_RecipientNotFound(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	transferService := service.NewTransferService(pool)

	sender, _ := authService.SignUp(ctx, "sender3", "S", "Three", "password123")
	fundAccount(t, pool, sender.ID, "1000")

	_, err := transferService.Transfer(ctx, "test-key-3", sender.ID, "nonexistent", decimal.NewFromInt(100))
	if err != service.ErrRecipientNotFound {
		t.Errorf("expected ErrRecipientNotFound, got %v", err)
	}
}

func TestTransferService_SelfTransfer(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	transferService := service.NewTransferService(pool)

	sender, _ := authService.SignUp(ctx, "sender4", "S", "Four", "password123")
	fundAccount(t, pool, sender.ID, "1000")

	_, err := transferService.Transfer(ctx, "test-key-4", sender.ID, "sender4", decimal.NewFromInt(100))
	if err != service.ErrSelfTransfer {
		t.Errorf("expected ErrSelfTransfer, got %v", err)
	}
}

func TestTransferService_IdempotentReplay(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	transferService := service.NewTransferService(pool)
	accountService := service.NewAccountService(pool)

	sender, _ := authService.SignUp(ctx, "sender5", "S", "Five", "password123")
	if _, err := authService.SignUp(ctx, "receiver5", "R", "Five", "password123"); err != nil {
		t.Fatalf("signup receiver: %v", err)
	}
	fundAccount(t, pool, sender.ID, "1000")

	key := "idempotency-test-key"

	result1, err := transferService.Transfer(ctx, key, sender.ID, "receiver5", decimal.NewFromInt(100))
	if err != nil {
		t.Fatalf("first transfer failed: %v", err)
	}

	result2, err := transferService.Transfer(ctx, key, sender.ID, "receiver5", decimal.NewFromInt(100))
	if err != nil {
		t.Fatalf("replayed transfer failed: %v", err)
	}

	if result1.TransactionID != result2.TransactionID {
		t.Errorf("replay should return same transaction id, got %s vs %s", result1.TransactionID, result2.TransactionID)
	}

	balance, _ := accountService.GetBalance(ctx, sender.ID)
	if !balance.Equal(decimal.NewFromInt(900)) {
		t.Errorf("expected balance 900 after replay (must not double-charge), got %s", balance.String())
	}
}

// This is the automated version of the manual 10-curl concurrency test --
// proves the row locking prevents lost updates under real concurrent load.
func TestTransferService_ConcurrentTransfers(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	transferService := service.NewTransferService(pool)
	accountService := service.NewAccountService(pool)

	sender, _ := authService.SignUp(ctx, "sender6", "S", "Six", "password123")
	receiver, err := authService.SignUp(ctx, "receiver6", "R", "Six", "password123")
	if err != nil {
		t.Fatalf("signup receiver: %v", err)
	}
	fundAccount(t, pool, sender.ID, "1000")

	const n = 10
	var wg sync.WaitGroup
	errs := make([]error, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("concurrent-key-%d", i)
			_, err := transferService.Transfer(ctx, key, sender.ID, "receiver6", decimal.NewFromInt(50))
			errs[i] = err
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("concurrent transfer %d failed: %v", i, err)
		}
	}

	senderBalance, _ := accountService.GetBalance(ctx, sender.ID)
	if !senderBalance.Equal(decimal.NewFromInt(500)) {
		t.Errorf("expected sender balance 500 after 10 concurrent transfers of 50, got %s", senderBalance.String())
	}

	receiverBalance, _ := accountService.GetBalance(ctx, receiver.ID)
	if !receiverBalance.Equal(decimal.NewFromInt(500)) {
		t.Errorf("expected receiver balance 500, got %s", receiverBalance.String())
	}
}
