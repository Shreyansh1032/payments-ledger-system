package service_test

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
	"github.com/Shreyansh1032/payments-ledger-system/internal/testutil"
)

func TestPaymentRequestService_CreateAndApprove(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	transferService := service.NewTransferService(pool)
	requestService := service.NewPaymentRequestService(pool, transferService)
	accountService := service.NewAccountService(pool)

	requester, _ := authService.SignUp(ctx, "req_alice", "A", "L", "password123")
	payer, _ := authService.SignUp(ctx, "req_bob", "B", "O", "password123")

	if _, err := pool.Exec(ctx, `UPDATE accounts SET balance = 1000 WHERE user_id = $1`, payer.ID); err != nil {
		t.Fatalf("fund payer: %v", err)
	}

	req, err := requestService.Create(ctx, requester.ID, "req_bob", "dinner", decimal.NewFromInt(200))
	if err != nil {
		t.Fatalf("create request failed: %v", err)
	}
	if req.Status != "pending" {
		t.Errorf("expected pending, got %s", req.Status)
	}

	if err := requestService.Approve(ctx, req.ID, payer.ID); err != nil {
		t.Fatalf("approve failed: %v", err)
	}

	requesterBalance, _ := accountService.GetBalance(ctx, requester.ID)
	if !requesterBalance.Equal(decimal.NewFromInt(200)) {
		t.Errorf("expected requester balance 200, got %s", requesterBalance.String())
	}
	payerBalance, _ := accountService.GetBalance(ctx, payer.ID)
	if !payerBalance.Equal(decimal.NewFromInt(800)) {
		t.Errorf("expected payer balance 800, got %s", payerBalance.String())
	}
}

func TestPaymentRequestService_Decline(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	transferService := service.NewTransferService(pool)
	requestService := service.NewPaymentRequestService(pool, transferService)

	requester, _ := authService.SignUp(ctx, "decline_alice", "A", "L", "password123")
	payer, _ := authService.SignUp(ctx, "decline_bob", "B", "O", "password123")

	req, err := requestService.Create(ctx, requester.ID, "decline_bob", "", decimal.NewFromInt(100))
	if err != nil {
		t.Fatalf("create request failed: %v", err)
	}

	if err := requestService.Decline(ctx, req.ID, payer.ID); err != nil {
		t.Fatalf("decline failed: %v", err)
	}

	outgoing, err := requestService.ListOutgoing(ctx, requester.ID)
	if err != nil {
		t.Fatalf("list outgoing failed: %v", err)
	}
	if len(outgoing) != 1 || outgoing[0].Status != "declined" {
		t.Errorf("expected 1 declined request, got %+v", outgoing)
	}
}

func TestPaymentRequestService_CannotApproveTwice(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	transferService := service.NewTransferService(pool)
	requestService := service.NewPaymentRequestService(pool, transferService)

	requester, _ := authService.SignUp(ctx, "twice_alice", "A", "L", "password123")
	payer, _ := authService.SignUp(ctx, "twice_bob", "B", "O", "password123")
	if _, err := pool.Exec(ctx, `UPDATE accounts SET balance = 1000 WHERE user_id = $1`, payer.ID); err != nil {
		t.Fatalf("fund payer: %v", err)
	}

	req, _ := requestService.Create(ctx, requester.ID, "twice_bob", "", decimal.NewFromInt(50))

	if err := requestService.Approve(ctx, req.ID, payer.ID); err != nil {
		t.Fatalf("first approve failed: %v", err)
	}

	err := requestService.Approve(ctx, req.ID, payer.ID)
	if err != service.ErrPaymentRequestNotPending {
		t.Errorf("expected ErrPaymentRequestNotPending on second approve, got %v", err)
	}
}

func TestPaymentRequestService_CreateSplit_DividesEvenly(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	transferService := service.NewTransferService(pool)
	requestService := service.NewPaymentRequestService(pool, transferService)

	creator, _ := authService.SignUp(ctx, "split_creator", "C", "R", "password123")
	if _, err := authService.SignUp(ctx, "split_a", "A", "A", "password123"); err != nil {
		t.Fatalf("signup split_a: %v", err)
	}
	if _, err := authService.SignUp(ctx, "split_b", "B", "B", "password123"); err != nil {
		t.Fatalf("signup split_b: %v", err)
	}
	if _, err := authService.SignUp(ctx, "split_c", "C", "C", "password123"); err != nil {
		t.Fatalf("signup split_c: %v", err)
	}

	results, err := requestService.CreateSplit(ctx, creator.ID, []string{"split_a", "split_b", "split_c"}, decimal.NewFromInt(300), "trip")
	if err != nil {
		t.Fatalf("create split failed: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 requests, got %d", len(results))
	}

	total := decimal.Zero
	for _, r := range results {
		total = total.Add(r.Amount)
	}
	if !total.Equal(decimal.NewFromInt(300)) {
		t.Errorf("expected shares to sum to 300, got %s", total.String())
	}
}
