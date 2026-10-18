package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
	"github.com/Shreyansh1032/payments-ledger-system/internal/testutil"
)

func TestScheduledTransferService_CreateValidatesFrequency(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	transferService := service.NewTransferService(pool)
	scheduledService := service.NewScheduledTransferService(pool, transferService)

	user, _ := authService.SignUp(ctx, "sched_user1", "S", "U", "password123")
	if _, err := authService.SignUp(ctx, "sched_target1", "T", "A", "password123"); err != nil {
		t.Fatalf("signup target: %v", err)
	}

	_, err := scheduledService.Create(ctx, user.ID, "sched_target1", "yearly", decimal.NewFromInt(100))
	if err != service.ErrInvalidFrequency {
		t.Errorf("expected ErrInvalidFrequency, got %v", err)
	}
}

func TestScheduledTransferService_CreateAndList(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	transferService := service.NewTransferService(pool)
	scheduledService := service.NewScheduledTransferService(pool, transferService)

	user, _ := authService.SignUp(ctx, "sched_user2", "S", "U", "password123")
	if _, err := authService.SignUp(ctx, "sched_target2", "T", "A", "password123"); err != nil {
		t.Fatalf("signup target: %v", err)
	}

	st, err := scheduledService.Create(ctx, user.ID, "sched_target2", "monthly", decimal.NewFromInt(500))
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if !st.NextRunAt.After(time.Now()) {
		t.Error("expected next_run_at to be in the future")
	}

	list, err := scheduledService.List(ctx, user.ID)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 scheduled transfer, got %d", len(list))
	}
}

func TestScheduledTransferService_RunDueExecutesAndReschedules(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	transferService := service.NewTransferService(pool)
	scheduledService := service.NewScheduledTransferService(pool, transferService)
	accountService := service.NewAccountService(pool)

	sender, _ := authService.SignUp(ctx, "sched_sender", "S", "D", "password123")
	if _, err := authService.SignUp(ctx, "sched_receiver", "R", "C", "password123"); err != nil {
		t.Fatalf("signup receiver: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE accounts SET balance = 1000 WHERE user_id = $1`, sender.ID); err != nil {
		t.Fatalf("fund sender: %v", err)
	}

	st, err := scheduledService.Create(ctx, sender.ID, "sched_receiver", "daily", decimal.NewFromInt(100))
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	if _, err := pool.Exec(ctx, `UPDATE scheduled_transfers SET next_run_at = now() - interval '1 minute' WHERE id = $1`, st.ID); err != nil {
		t.Fatalf("force due: %v", err)
	}

	if err := scheduledService.RunDue(ctx); err != nil {
		t.Fatalf("RunDue failed: %v", err)
	}

	senderBalance, _ := accountService.GetBalance(ctx, sender.ID)
	if !senderBalance.Equal(decimal.NewFromInt(900)) {
		t.Errorf("expected sender balance 900 after scheduled run, got %s", senderBalance.String())
	}

	list, _ := scheduledService.List(ctx, sender.ID)
	if list[0].LastRunStatus == nil || *list[0].LastRunStatus != "completed" {
		t.Errorf("expected last_run_status completed, got %+v", list[0].LastRunStatus)
	}
	if !list[0].NextRunAt.After(time.Now()) {
		t.Error("expected next_run_at to be rescheduled into the future")
	}

	if err := scheduledService.RunDue(ctx); err != nil {
		t.Fatalf("second RunDue failed: %v", err)
	}
	senderBalance, _ = accountService.GetBalance(ctx, sender.ID)
	if !senderBalance.Equal(decimal.NewFromInt(900)) {
		t.Errorf("expected balance unchanged at 900 after non-due RunDue, got %s", senderBalance.String())
	}
}

func TestScheduledTransferService_Cancel(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.Truncate(t, pool)
	ctx := context.Background()

	authService := service.NewAuthService(pool)
	transferService := service.NewTransferService(pool)
	scheduledService := service.NewScheduledTransferService(pool, transferService)

	user, _ := authService.SignUp(ctx, "sched_cancel_user", "S", "U", "password123")
	if _, err := authService.SignUp(ctx, "sched_cancel_target", "T", "A", "password123"); err != nil {
		t.Fatalf("signup target: %v", err)
	}

	st, _ := scheduledService.Create(ctx, user.ID, "sched_cancel_target", "weekly", decimal.NewFromInt(50))

	if err := scheduledService.Cancel(ctx, user.ID, st.ID); err != nil {
		t.Fatalf("cancel failed: %v", err)
	}

	if _, err := pool.Exec(ctx, `UPDATE scheduled_transfers SET next_run_at = now() - interval '1 minute' WHERE id = $1`, st.ID); err != nil {
		t.Fatalf("force due: %v", err)
	}
	if err := scheduledService.RunDue(ctx); err != nil {
		t.Fatalf("RunDue failed: %v", err)
	}

	list, _ := scheduledService.List(ctx, user.ID)
	if list[0].LastRunStatus != nil {
		t.Error("cancelled schedule should never execute")
	}
}
