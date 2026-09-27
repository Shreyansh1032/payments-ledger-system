package service_test

import (
	"context"
	"testing"

	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
	"github.com/Shreyansh1032/payments-ledger-system/internal/testutil"
)

func TestUserSearchService_FindsMatchingUsersAlphabetically(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	searchService := service.NewUserSearchService(pool)

	requester, _ := authService.SignUp(ctx, "searcher", "S", "R", "password123")
	if _, err := authService.SignUp(ctx, "priyasharma", "Priya", "Sharma", "password123"); err != nil {
		t.Fatalf("signup priyasharma: %v", err)
	}
	if _, err := authService.SignUp(ctx, "priyankasingh", "Priyanka", "Singh", "password123"); err != nil {
		t.Fatalf("signup priyankasingh: %v", err)
	}
	if _, err := authService.SignUp(ctx, "rohit", "Rohit", "Kumar", "password123"); err != nil {
		t.Fatalf("signup rohit: %v", err)
	}

	results, err := searchService.Search(ctx, requester.ID, "priy")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 matches for 'priy', got %d", len(results))
	}
	if results[0].Username != "priyankasingh" || results[1].Username != "priyasharma" {
		t.Errorf("expected alphabetical [priyankasingh, priyasharma], got [%s, %s]", results[0].Username, results[1].Username)
	}
}

func TestUserSearchService_ExcludesSelfAndSystemAccount(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	searchService := service.NewUserSearchService(pool)

	requester, _ := authService.SignUp(ctx, "selftest", "S", "T", "password123")

	results, err := searchService.Search(ctx, requester.ID, "s")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	for _, r := range results {
		if r.Username == "selftest" {
			t.Error("search should not include the requesting user")
		}
		if r.Username == "system_treasury" {
			t.Error("search should not include the system treasury account")
		}
	}
}

func TestUserSearchService_EmptyQueryReturnsEmpty(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	searchService := service.NewUserSearchService(pool)

	requester, _ := authService.SignUp(ctx, "emptyquerytest", "E", "Q", "password123")

	results, err := searchService.Search(ctx, requester.ID, "")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty results for empty query, got %d", len(results))
	}
}
