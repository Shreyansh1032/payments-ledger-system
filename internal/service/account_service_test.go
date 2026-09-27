package service_test

import (
	"context"
	"testing"

	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
	"github.com/Shreyansh1032/payments-ledger-system/internal/testutil"
)

func TestAccountService_GetProfile(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	accountService := service.NewAccountService(pool)

	user, _ := authService.SignUp(ctx, "profileuser", "Profile", "User", "password123")

	profile, err := accountService.GetProfile(ctx, user.ID)
	if err != nil {
		t.Fatalf("get profile failed: %v", err)
	}
	if profile.Username != "profileuser" {
		t.Errorf("expected username profileuser, got %s", profile.Username)
	}
	if profile.CreatedAt.IsZero() {
		t.Error("expected created_at to be set")
	}
}
