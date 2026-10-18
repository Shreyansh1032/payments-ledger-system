package service_test

import (
	"context"
	"testing"

	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
	"github.com/Shreyansh1032/payments-ledger-system/internal/testutil"
)

type fakeEmailSender struct {
	lastLink string
}

func (f *fakeEmailSender) SendVerificationEmail(_ context.Context, _, link string) error {
	f.lastLink = link
	return nil
}

func TestEmailVerificationService_SetAndVerify(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	sender := &fakeEmailSender{}
	verificationService := service.NewEmailVerificationService(pool, sender)

	user, _ := authService.SignUp(ctx, "email_user1", "E", "U", "password123")

	if err := verificationService.SetEmail(ctx, user.ID, "email_user1@example.com", "http://localhost:8080"); err != nil {
		t.Fatalf("set email failed: %v", err)
	}
	if sender.lastLink == "" {
		t.Fatal("expected a verification link to have been sent")
	}

	token := sender.lastLink[len("http://localhost:8080/verify-email?token="):]
	if err := verificationService.Verify(ctx, token); err != nil {
		t.Fatalf("verify failed: %v", err)
	}

	var verified bool
	if err := pool.QueryRow(ctx, `SELECT email_verified FROM users WHERE id=$1`, user.ID).Scan(&verified); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if !verified {
		t.Error("expected email_verified to be true")
	}
}

func TestEmailVerificationService_RejectsInvalidToken(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	verificationService := service.NewEmailVerificationService(pool, &fakeEmailSender{})

	err := verificationService.Verify(ctx, "not-a-real-token")
	if err != service.ErrInvalidVerificationToken {
		t.Errorf("expected ErrInvalidVerificationToken, got %v", err)
	}
}

func TestEmailVerificationService_RejectsDuplicateEmail(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	verificationService := service.NewEmailVerificationService(pool, &fakeEmailSender{})

	userA, _ := authService.SignUp(ctx, "email_user_a", "A", "A", "password123")
	userB, _ := authService.SignUp(ctx, "email_user_b", "B", "B", "password123")

	if err := verificationService.SetEmail(ctx, userA.ID, "shared@example.com", "http://localhost:8080"); err != nil {
		t.Fatalf("set email for A failed: %v", err)
	}

	err := verificationService.SetEmail(ctx, userB.ID, "shared@example.com", "http://localhost:8080")
	if err != service.ErrEmailAlreadyInUse {
		t.Errorf("expected ErrEmailAlreadyInUse, got %v", err)
	}
}
