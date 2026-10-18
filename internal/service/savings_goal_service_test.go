package service_test

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
	"github.com/Shreyansh1032/payments-ledger-system/internal/testutil"
)

func TestSavingsGoalService_CreateAndContribute(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	goalService := service.NewSavingsGoalService(pool)

	user, _ := authService.SignUp(ctx, "goal_user1", "G", "U", "password123")

	goal, err := goalService.Create(ctx, user.ID, "New laptop", decimal.NewFromInt(1000))
	if err != nil {
		t.Fatalf("create goal failed: %v", err)
	}

	updated, justReached, err := goalService.AddContribution(ctx, user.ID, goal.ID, decimal.NewFromInt(300))
	if err != nil {
		t.Fatalf("contribute failed: %v", err)
	}
	if justReached {
		t.Error("should not be reached yet")
	}
	if !updated.CurrentAmount.Equal(decimal.NewFromInt(300)) {
		t.Errorf("expected current amount 300, got %s", updated.CurrentAmount.String())
	}
}

func TestSavingsGoalService_DetectsGoalReached(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	goalService := service.NewSavingsGoalService(pool)

	user, _ := authService.SignUp(ctx, "goal_user2", "G", "U", "password123")
	goal, _ := goalService.Create(ctx, user.ID, "Trip", decimal.NewFromInt(500))

	_, justReached, err := goalService.AddContribution(ctx, user.ID, goal.ID, decimal.NewFromInt(500))
	if err != nil {
		t.Fatalf("contribute failed: %v", err)
	}
	if !justReached {
		t.Error("expected justReached true when contribution meets the target exactly")
	}

	_, justReachedAgain, err := goalService.AddContribution(ctx, user.ID, goal.ID, decimal.NewFromInt(50))
	if err != nil {
		t.Fatalf("second contribute failed: %v", err)
	}
	if justReachedAgain {
		t.Error("should not report justReached again once already past the target")
	}
}

func TestSavingsGoalService_DeleteAndNotFound(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	goalService := service.NewSavingsGoalService(pool)

	user, _ := authService.SignUp(ctx, "goal_user3", "G", "U", "password123")
	goal, _ := goalService.Create(ctx, user.ID, "Emergency fund", decimal.NewFromInt(200))

	if err := goalService.Delete(ctx, user.ID, goal.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	_, _, err := goalService.AddContribution(ctx, user.ID, goal.ID, decimal.NewFromInt(10))
	if err != service.ErrSavingsGoalNotFound {
		t.Errorf("expected ErrSavingsGoalNotFound after delete, got %v", err)
	}
}
