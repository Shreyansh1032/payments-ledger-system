package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

const SystemTreasuryUsername = "system_treasury"

// Deposit adds funds to a user's own account, sourced from the system
// treasury account. It reuses the same locked, idempotent, double-entry
// pattern as Transfer -- the only difference is the sender is the system
// account, so the balance check on the sender is skipped.
func (s *TransferService) Deposit(ctx context.Context, idempotencyKey, userID string, amount decimal.Decimal) (*TransferResult, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalidAmount
	}

	var existingID, existingStatus string
	err := s.pool.QueryRow(ctx,
		`SELECT id, status FROM transactions WHERE idempotency_key=$1`,
		idempotencyKey,
	).Scan(&existingID, &existingStatus)
	if err == nil {
		return &TransferResult{TransactionID: existingID, Status: existingStatus}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var systemAccountID string
	if err := tx.QueryRow(ctx,
		`SELECT accounts.id FROM accounts JOIN users ON users.id = accounts.user_id WHERE users.username=$1`,
		SystemTreasuryUsername,
	).Scan(&systemAccountID); err != nil {
		return nil, err
	}

	var toAccountID string
	if err := tx.QueryRow(ctx, `SELECT id FROM accounts WHERE user_id=$1`, userID).Scan(&toAccountID); err != nil {
		return nil, err
	}

	firstID, secondID := systemAccountID, toAccountID
	if secondID < firstID {
		firstID, secondID = secondID, firstID
	}
	if _, err := tx.Exec(ctx, `SELECT id FROM accounts WHERE id=$1 FOR UPDATE`, firstID); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `SELECT id FROM accounts WHERE id=$1 FOR UPDATE`, secondID); err != nil {
		return nil, err
	}

	var txnID string
	err = tx.QueryRow(ctx,
		`INSERT INTO transactions (idempotency_key, from_account_id, to_account_id, amount, status)
		 VALUES ($1, $2, $3, $4, 'completed')
		 RETURNING id`,
		idempotencyKey, systemAccountID, toAccountID, amount,
	).Scan(&txnID)
	if err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx,
		`UPDATE accounts SET balance = balance - $1, updated_at = now() WHERE id=$2`,
		amount, systemAccountID,
	); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx,
		`UPDATE accounts SET balance = balance + $1, updated_at = now() WHERE id=$2`,
		amount, toAccountID,
	); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO ledger_entries (transaction_id, account_id, entry_type, amount) VALUES ($1, $2, 'debit', $3)`,
		txnID, systemAccountID, amount,
	); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO ledger_entries (transaction_id, account_id, entry_type, amount) VALUES ($1, $2, 'credit', $3)`,
		txnID, toAccountID, amount,
	); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `UPDATE transactions SET completed_at = now() WHERE id=$1`, txnID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &TransferResult{TransactionID: txnID, Status: "completed"}, nil
}
