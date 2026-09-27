package service_test

import (
	"context"
	"testing"

	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
	"github.com/Shreyansh1032/payments-ledger-system/internal/testutil"
)

func TestAuthService_SignUp(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	accountService := service.NewAccountService(pool)

	user, err := authService.SignUp(ctx, "alice", "Alice", "A", "password123")
	if err != nil {
		t.Fatalf("signup failed: %v", err)
	}
	if user.Username != "alice" {
		t.Errorf("expected username alice, got %s", user.Username)
	}

	balance, err := accountService.GetBalance(ctx, user.ID)
	if err != nil {
		t.Fatalf("get balance failed: %v", err)
	}
	if !balance.IsZero() {
		t.Errorf("expected 0 balance for new account, got %s", balance.String())
	}
}

func TestAuthService_SignUp_DuplicateUsername(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)

	if _, err := authService.SignUp(ctx, "bob", "Bob", "B", "password123"); err != nil {
		t.Fatalf("first signup failed: %v", err)
	}

	_, err := authService.SignUp(ctx, "bob", "Bob", "B", "password123")
	if err != service.ErrUsernameTaken {
		t.Errorf("expected ErrUsernameTaken, got %v", err)
	}
}

func TestAuthService_Login(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)

	if _, err := authService.SignUp(ctx, "carol", "Carol", "C", "password123"); err != nil {
		t.Fatalf("signup failed: %v", err)
	}

	user, err := authService.Login(ctx, "carol", "password123")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if user.Username != "carol" {
		t.Errorf("expected username carol, got %s", user.Username)
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)

	if _, err := authService.SignUp(ctx, "dave", "Dave", "D", "password123"); err != nil {
		t.Fatalf("signup failed: %v", err)
	}

	_, err := authService.Login(ctx, "dave", "wrongpassword")
	if err != service.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_ChangePassword_Success(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)

	user, _ := authService.SignUp(ctx, "pwchange", "P", "C", "oldpassword")

	if err := authService.ChangePassword(ctx, user.ID, "oldpassword", "newpassword123"); err != nil {
		t.Fatalf("change password failed: %v", err)
	}

	if _, err := authService.Login(ctx, "pwchange", "oldpassword"); err != service.ErrInvalidCredentials {
		t.Errorf("expected old password to be rejected, got %v", err)
	}
	if _, err := authService.Login(ctx, "pwchange", "newpassword123"); err != nil {
		t.Errorf("expected new password to work, got %v", err)
	}
}

func TestAuthService_ChangePassword_WrongCurrentPassword(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)

	if _, err := authService.SignUp(ctx, "pwwrong", "P", "W", "correctpassword"); err != nil {
		t.Fatalf("signup failed: %v", err)
	}

	user, _ := authService.Login(ctx, "pwwrong", "correctpassword")
	err := authService.ChangePassword(ctx, user.ID, "wrongpassword", "newpassword123")
	if err != service.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}
