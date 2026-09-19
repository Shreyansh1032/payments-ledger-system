package service_test

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
	"github.com/Shreyansh1032/payments-ledger-system/internal/testutil"
)

func TestTransactionHistoryService_ReturnsSentAndReceived(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	transferService := service.NewTransferService(pool)
	historyService := service.NewTransactionHistoryService(pool)

	alice, _ := authService.SignUp(ctx, "alice_hist", "A", "H", "password123")
	bob, _ := authService.SignUp(ctx, "bob_hist", "B", "H", "password123")

	if _, err := pool.Exec(ctx, `UPDATE accounts SET balance = 1000 WHERE user_id = $1`, alice.ID); err != nil {
		t.Fatalf("fund alice: %v", err)
	}

	if _, err := transferService.Transfer(ctx, "hist-key-1", alice.ID, "bob_hist", decimal.NewFromInt(150)); err != nil {
		t.Fatalf("transfer failed: %v", err)
	}

	aliceHistory, err := historyService.GetHistory(ctx, alice.ID)
	if err != nil {
		t.Fatalf("get alice history: %v", err)
	}
	if len(aliceHistory) != 1 {
		t.Fatalf("expected 1 entry for alice, got %d", len(aliceHistory))
	}
	if aliceHistory[0].Direction != "sent" {
		t.Errorf("expected sent, got %s", aliceHistory[0].Direction)
	}
	if aliceHistory[0].Counterparty != "bob_hist" {
		t.Errorf("expected counterparty bob_hist, got %s", aliceHistory[0].Counterparty)
	}

	bobHistory, err := historyService.GetHistory(ctx, bob.ID)
	if err != nil {
		t.Fatalf("get bob history: %v", err)
	}
	if len(bobHistory) != 1 {
		t.Fatalf("expected 1 entry for bob, got %d", len(bobHistory))
	}
	if bobHistory[0].Direction != "received" {
		t.Errorf("expected received, got %s", bobHistory[0].Direction)
	}
	if bobHistory[0].Counterparty != "alice_hist" {
		t.Errorf("expected counterparty alice_hist, got %s", bobHistory[0].Counterparty)
	}
}

func TestTransactionHistoryService_EmptyForNewUser(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	historyService := service.NewTransactionHistoryService(pool)

	user, _ := authService.SignUp(ctx, "newuser_hist", "N", "H", "password123")

	history, err := historyService.GetHistory(ctx, user.ID)
	if err != nil {
		t.Fatalf("get history: %v", err)
	}
	if len(history) != 0 {
		t.Errorf("expected empty history, got %d entries", len(history))
	}
}
