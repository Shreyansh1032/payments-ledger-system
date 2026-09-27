package service_test

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
	"github.com/Shreyansh1032/payments-ledger-system/internal/testutil"
)

func TestInsightsService_MonthlyTotals(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	transferService := service.NewTransferService(pool)
	insightsService := service.NewInsightsService(pool)

	alice, _ := authService.SignUp(ctx, "insights_alice", "A", "L", "password123")
	if _, err := authService.SignUp(ctx, "insights_bob", "B", "O", "password123"); err != nil {
		t.Fatalf("signup bob: %v", err)
	}

	if _, err := pool.Exec(ctx, `UPDATE accounts SET balance = 1000 WHERE user_id = $1`, alice.ID); err != nil {
		t.Fatalf("fund alice: %v", err)
	}

	if _, err := transferService.Transfer(ctx, "insights-key-1", alice.ID, "insights_bob", decimal.NewFromInt(150)); err != nil {
		t.Fatalf("transfer failed: %v", err)
	}

	insights, err := insightsService.GetMonthlyInsights(ctx, alice.ID)
	if err != nil {
		t.Fatalf("get insights failed: %v", err)
	}
	if !insights.TotalSent.Equal(decimal.NewFromInt(150)) {
		t.Errorf("expected total sent 150, got %s", insights.TotalSent.String())
	}
	if len(insights.TopRecipients) != 1 || insights.TopRecipients[0].Username != "insights_bob" {
		t.Errorf("expected top recipient insights_bob, got %+v", insights.TopRecipients)
	}
}
