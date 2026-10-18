package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"

	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
	"github.com/Shreyansh1032/payments-ledger-system/internal/testutil"
)

func TestTOTPService_SetupConfirmAndValidate(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	totpService := service.NewTOTPService(pool)

	user, _ := authService.SignUp(ctx, "totp_user1", "T", "U", "password123")

	key, err := totpService.BeginSetup(ctx, user.ID, "totp_user1")
	if err != nil {
		t.Fatalf("begin setup failed: %v", err)
	}

	code, err := totp.GenerateCode(key.Secret(), time.Now())
	if err != nil {
		t.Fatalf("generate code failed: %v", err)
	}

	if err := totpService.Confirm(ctx, user.ID, code); err != nil {
		t.Fatalf("confirm failed: %v", err)
	}

	enabled, err := totpService.IsEnabled(ctx, user.ID)
	if err != nil {
		t.Fatalf("is enabled failed: %v", err)
	}
	if !enabled {
		t.Error("expected 2FA to be enabled after confirm")
	}

	validCode, _ := totp.GenerateCode(key.Secret(), time.Now())
	valid, err := totpService.ValidateLogin(ctx, user.ID, validCode)
	if err != nil {
		t.Fatalf("validate login failed: %v", err)
	}
	if !valid {
		t.Error("expected a freshly generated code to validate")
	}

	valid, err = totpService.ValidateLogin(ctx, user.ID, "000000")
	if err != nil {
		t.Fatalf("validate login failed: %v", err)
	}
	if valid {
		t.Error("expected an arbitrary wrong code to be rejected")
	}
}

func TestTOTPService_CannotSetupTwice(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	totpService := service.NewTOTPService(pool)

	user, _ := authService.SignUp(ctx, "totp_user2", "T", "U", "password123")

	key, _ := totpService.BeginSetup(ctx, user.ID, "totp_user2")
	code, _ := totp.GenerateCode(key.Secret(), time.Now())
	if err := totpService.Confirm(ctx, user.ID, code); err != nil {
		t.Fatalf("confirm failed: %v", err)
	}

	_, err := totpService.BeginSetup(ctx, user.ID, "totp_user2")
	if err != service.ErrTOTPAlreadyEnabled {
		t.Errorf("expected ErrTOTPAlreadyEnabled, got %v", err)
	}
}

func TestTOTPService_DisableRequiresValidCode(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	totpService := service.NewTOTPService(pool)

	user, _ := authService.SignUp(ctx, "totp_user3", "T", "U", "password123")
	key, _ := totpService.BeginSetup(ctx, user.ID, "totp_user3")
	code, _ := totp.GenerateCode(key.Secret(), time.Now())
	if err := totpService.Confirm(ctx, user.ID, code); err != nil {
		t.Fatalf("confirm failed: %v", err)
	}

	if err := totpService.Disable(ctx, user.ID, "000000"); err != service.ErrInvalidTOTPCode {
		t.Errorf("expected ErrInvalidTOTPCode for wrong code, got %v", err)
	}

	validCode, _ := totp.GenerateCode(key.Secret(), time.Now())
	if err := totpService.Disable(ctx, user.ID, validCode); err != nil {
		t.Fatalf("disable with valid code failed: %v", err)
	}

	enabled, _ := totpService.IsEnabled(ctx, user.ID)
	if enabled {
		t.Error("expected 2FA disabled")
	}
}
