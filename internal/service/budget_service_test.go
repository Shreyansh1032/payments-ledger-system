package service_test

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
	"github.com/Shreyansh1032/payments-ledger-system/internal/testutil"
)

func TestBudgetService_TracksSpendAgainstLimit(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	transferService := service.NewTransferService(pool)
	budgetService := service.NewBudgetService(pool)
	categoryService := service.NewTransactionCategoryService(pool)

	alice, _ := authService.SignUp(ctx, "budget_alice", "A", "L", "password123")
	if _, err := authService.SignUp(ctx, "budget_bob", "B", "O", "password123"); err != nil {
		t.Fatalf("signup bob: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE accounts SET balance = 1000 WHERE user_id = $1`, alice.ID); err != nil {
		t.Fatalf("fund alice: %v", err)
	}

	if err := budgetService.SetLimit(ctx, alice.ID, "food", decimal.NewFromInt(200)); err != nil {
		t.Fatalf("set limit failed: %v", err)
	}

	result, err := transferService.Transfer(ctx, "budget-key-1", alice.ID, "budget_bob", decimal.NewFromInt(80))
	if err != nil {
		t.Fatalf("transfer failed: %v", err)
	}
	if err := categoryService.SetCategory(ctx, alice.ID, result.TransactionID, "food"); err != nil {
		t.Fatalf("set category failed: %v", err)
	}

	budgets, err := budgetService.List(ctx, alice.ID)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(budgets) != 1 {
		t.Fatalf("expected 1 budget, got %d", len(budgets))
	}
	if !budgets[0].Spent.Equal(decimal.NewFromInt(80)) {
		t.Errorf("expected spent 80, got %s", budgets[0].Spent.String())
	}
}

func TestBudgetService_SetLimitUpdatesExisting(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	budgetService := service.NewBudgetService(pool)

	user, _ := authService.SignUp(ctx, "budget_update_user", "B", "U", "password123")

	if err := budgetService.SetLimit(ctx, user.ID, "transport", decimal.NewFromInt(100)); err != nil {
		t.Fatalf("first set failed: %v", err)
	}
	if err := budgetService.SetLimit(ctx, user.ID, "transport", decimal.NewFromInt(150)); err != nil {
		t.Fatalf("second set failed: %v", err)
	}

	budgets, _ := budgetService.List(ctx, user.ID)
	if len(budgets) != 1 {
		t.Fatalf("expected limit update not a duplicate row, got %d budgets", len(budgets))
	}
	if !budgets[0].MonthlyLimit.Equal(decimal.NewFromInt(150)) {
		t.Errorf("expected updated limit 150, got %s", budgets[0].MonthlyLimit.String())
	}
}
