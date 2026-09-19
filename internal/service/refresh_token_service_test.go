package service_test

import (
	"context"
	"testing"

	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
	"github.com/Shreyansh1032/payments-ledger-system/internal/testutil"
)

func TestRefreshTokenService_IssueAndRotate(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	refreshTokenService := service.NewRefreshTokenService(pool)

	user, err := authService.SignUp(ctx, "rotateuser", "R", "U", "password123")
	if err != nil {
		t.Fatalf("signup failed: %v", err)
	}

	token, err := refreshTokenService.Issue(ctx, user.ID)
	if err != nil {
		t.Fatalf("issue failed: %v", err)
	}

	userID, username, newToken, err := refreshTokenService.Rotate(ctx, token)
	if err != nil {
		t.Fatalf("rotate failed: %v", err)
	}
	if userID != user.ID {
		t.Errorf("expected user id %s, got %s", user.ID, userID)
	}
	if username != "rotateuser" {
		t.Errorf("expected username rotateuser, got %s", username)
	}
	if newToken == token {
		t.Error("expected a new token, got the same one back")
	}
}

func TestRefreshTokenService_CannotReuseRotatedToken(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	refreshTokenService := service.NewRefreshTokenService(pool)

	user, _ := authService.SignUp(ctx, "reuseuser", "R", "U", "password123")
	token, _ := refreshTokenService.Issue(ctx, user.ID)

	if _, _, _, err := refreshTokenService.Rotate(ctx, token); err != nil {
		t.Fatalf("first rotate failed: %v", err)
	}

	_, _, _, err := refreshTokenService.Rotate(ctx, token)
	if err != service.ErrInvalidRefreshToken {
		t.Errorf("expected ErrInvalidRefreshToken on reuse, got %v", err)
	}
}

func TestRefreshTokenService_InvalidToken(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	refreshTokenService := service.NewRefreshTokenService(pool)

	_, _, _, err := refreshTokenService.Rotate(ctx, "totally-made-up-token")
	if err != service.ErrInvalidRefreshToken {
		t.Errorf("expected ErrInvalidRefreshToken, got %v", err)
	}
}

func TestRefreshTokenService_Revoke(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	refreshTokenService := service.NewRefreshTokenService(pool)

	user, _ := authService.SignUp(ctx, "revokeuser", "R", "U", "password123")
	token, _ := refreshTokenService.Issue(ctx, user.ID)

	if err := refreshTokenService.Revoke(ctx, token); err != nil {
		t.Fatalf("revoke failed: %v", err)
	}

	_, _, _, err := refreshTokenService.Rotate(ctx, token)
	if err != service.ErrInvalidRefreshToken {
		t.Errorf("expected ErrInvalidRefreshToken after revoke, got %v", err)
	}
}
