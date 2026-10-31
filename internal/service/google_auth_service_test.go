package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
	"github.com/Shreyansh1032/payments-ledger-system/internal/testutil"
)

type fakeGoogleVerifier struct {
	claims *service.GoogleClaims
	err    error
}

func (f *fakeGoogleVerifier) Verify(_ context.Context, _ string) (*service.GoogleClaims, error) {
	return f.claims, f.err
}

func TestGoogleAuthService_CreatesNewUserOnFirstSignIn(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	verifier := &fakeGoogleVerifier{claims: &service.GoogleClaims{
		Subject: "google-sub-1", Email: "newgoogleuser@example.com",
		EmailVerified: true, GivenName: "New", FamilyName: "User",
	}}
	googleAuthService := service.NewGoogleAuthService(pool, verifier)

	user, isNew, err := googleAuthService.SignIn(ctx, "fake-token")
	if err != nil {
		t.Fatalf("sign in failed: %v", err)
	}
	if !isNew {
		t.Error("expected a brand new user to be reported as new")
	}
	if user.FirstName != "New" {
		t.Errorf("expected first name New, got %s", user.FirstName)
	}
}

func TestGoogleAuthService_ReturnsSameUserOnRepeatSignIn(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	verifier := &fakeGoogleVerifier{claims: &service.GoogleClaims{
		Subject: "google-sub-2", Email: "repeatuser@example.com",
		EmailVerified: true, GivenName: "Repeat", FamilyName: "User",
	}}
	googleAuthService := service.NewGoogleAuthService(pool, verifier)

	first, isNew1, err := googleAuthService.SignIn(ctx, "fake-token")
	if err != nil {
		t.Fatalf("first sign in failed: %v", err)
	}
	if !isNew1 {
		t.Error("expected first sign-in to be new")
	}

	second, isNew2, err := googleAuthService.SignIn(ctx, "fake-token")
	if err != nil {
		t.Fatalf("second sign in failed: %v", err)
	}
	if isNew2 {
		t.Error("expected second sign-in to not be new")
	}
	if first.ID != second.ID {
		t.Errorf("expected same user id on repeat sign-in, got %s vs %s", first.ID, second.ID)
	}
}

func TestGoogleAuthService_LinksToExistingVerifiedEmail(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	existingUser, err := authService.SignUp(ctx, "existing_linker", "Existing", "Linker", "password123")
	if err != nil {
		t.Fatalf("signup failed: %v", err)
	}

	if _, err := pool.Exec(ctx, `UPDATE users SET email=$1, email_verified=true WHERE id=$2`, "linkme@example.com", existingUser.ID); err != nil {
		t.Fatalf("set verified email: %v", err)
	}

	verifier := &fakeGoogleVerifier{claims: &service.GoogleClaims{
		Subject: "google-sub-3", Email: "linkme@example.com",
		EmailVerified: true, GivenName: "Ignored", FamilyName: "Ignored",
	}}
	googleAuthService := service.NewGoogleAuthService(pool, verifier)

	user, isNew, err := googleAuthService.SignIn(ctx, "fake-token")
	if err != nil {
		t.Fatalf("sign in failed: %v", err)
	}
	if isNew {
		t.Error("expected linking to an existing user, not a new one")
	}
	if user.ID != existingUser.ID {
		t.Errorf("expected linked to existing user %s, got %s", existingUser.ID, user.ID)
	}
	if user.Username != "existing_linker" {
		t.Errorf("expected existing username preserved, got %s", user.Username)
	}
}

func TestGoogleAuthService_RejectsUnverifiedEmail(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	verifier := &fakeGoogleVerifier{claims: &service.GoogleClaims{
		Subject: "google-sub-4", Email: "unverified@example.com", EmailVerified: false,
	}}
	googleAuthService := service.NewGoogleAuthService(pool, verifier)

	_, _, err := googleAuthService.SignIn(ctx, "fake-token")
	if err != service.ErrGoogleEmailNotVerified {
		t.Errorf("expected ErrGoogleEmailNotVerified, got %v", err)
	}
}

func TestGoogleAuthService_RejectsInvalidToken(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	verifier := &fakeGoogleVerifier{err: errors.New("bad token")}
	googleAuthService := service.NewGoogleAuthService(pool, verifier)

	_, _, err := googleAuthService.SignIn(ctx, "fake-token")
	if err != service.ErrGoogleTokenInvalid {
		t.Errorf("expected ErrGoogleTokenInvalid, got %v", err)
	}
}
