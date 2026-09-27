package service_test

import (
	"context"
	"testing"

	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
	"github.com/Shreyansh1032/payments-ledger-system/internal/testutil"
)

func TestFavoriteService_AddListRemove(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	favoriteService := service.NewFavoriteService(pool)

	user, _ := authService.SignUp(ctx, "fav_user", "F", "U", "password123")
	if _, err := authService.SignUp(ctx, "fav_target", "T", "A", "password123"); err != nil {
		t.Fatalf("signup fav_target: %v", err)
	}

	if err := favoriteService.Add(ctx, user.ID, "fav_target"); err != nil {
		t.Fatalf("add favorite failed: %v", err)
	}

	favorites, err := favoriteService.List(ctx, user.ID)
	if err != nil {
		t.Fatalf("list favorites failed: %v", err)
	}
	if len(favorites) != 1 || favorites[0].Username != "fav_target" {
		t.Fatalf("expected 1 favorite fav_target, got %+v", favorites)
	}

	if err := favoriteService.Remove(ctx, user.ID, "fav_target"); err != nil {
		t.Fatalf("remove favorite failed: %v", err)
	}

	favorites, err = favoriteService.List(ctx, user.ID)
	if err != nil {
		t.Fatalf("list favorites failed: %v", err)
	}
	if len(favorites) != 0 {
		t.Errorf("expected 0 favorites after remove, got %d", len(favorites))
	}
}

func TestFavoriteService_CannotFavoriteSelf(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	favoriteService := service.NewFavoriteService(pool)

	user, _ := authService.SignUp(ctx, "fav_self", "S", "S", "password123")

	err := favoriteService.Add(ctx, user.ID, "fav_self")
	if err != service.ErrCannotFavoriteSelf {
		t.Errorf("expected ErrCannotFavoriteSelf, got %v", err)
	}
}

func TestFavoriteService_AddIsIdempotent(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	favoriteService := service.NewFavoriteService(pool)

	user, _ := authService.SignUp(ctx, "fav_dup", "D", "U", "password123")
	if _, err := authService.SignUp(ctx, "fav_dup_target", "T", "A", "password123"); err != nil {
		t.Fatalf("signup: %v", err)
	}

	if err := favoriteService.Add(ctx, user.ID, "fav_dup_target"); err != nil {
		t.Fatalf("first add failed: %v", err)
	}
	if err := favoriteService.Add(ctx, user.ID, "fav_dup_target"); err != nil {
		t.Fatalf("second add (duplicate) should not error, got: %v", err)
	}

	favorites, _ := favoriteService.List(ctx, user.ID)
	if len(favorites) != 1 {
		t.Errorf("expected exactly 1 favorite after duplicate add, got %d", len(favorites))
	}
}
