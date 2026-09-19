package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

var (
	ErrRecipientNotFound   = errors.New("recipient not found")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrSelfTransfer        = errors.New("cannot transfer to yourself")
	ErrInvalidAmount       = errors.New("amount must be positive")
)

type TransferResult struct {
	TransactionID string
	Status        string
}

type TransferService struct {
	pool *pgxpool.Pool
}

func NewTransferService(pool *pgxpool.Pool) *TransferService {
	return &TransferService{pool: pool}
}

func (s *TransferService) Transfer(ctx context.Context, idempotencyKey, fromUserID, toUsername string, amount decimal.Decimal) (*TransferResult, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalidAmount
	}

	// idempotent replay: same key -> return the original result, don't re-process
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

	var fromAccountID string
	if err := tx.QueryRow(ctx, `SELECT id FROM accounts WHERE user_id=$1`, fromUserID).Scan(&fromAccountID); err != nil {
		return nil, err
	}

	var toAccountID, toUserID string
	err = tx.QueryRow(ctx,
		`SELECT accounts.id, users.id FROM accounts JOIN users ON users.id = accounts.user_id WHERE users.username=$1`,
		toUsername,
	).Scan(&toAccountID, &toUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRecipientNotFound
		}
		return nil, err
	}

	if fromUserID == toUserID {
		return nil, ErrSelfTransfer
	}

	// lock both accounts in a consistent order (sorted by id), regardless of
	// who's sending and who's receiving -- this is what prevents deadlocks
	// when two transfers happen concurrently in opposite directions
	firstID, secondID := fromAccountID, toAccountID
	if secondID < firstID {
		firstID, secondID = secondID, firstID
	}
	if _, err := tx.Exec(ctx, `SELECT id FROM accounts WHERE id=$1 FOR UPDATE`, firstID); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `SELECT id FROM accounts WHERE id=$1 FOR UPDATE`, secondID); err != nil {
		return nil, err
	}

	var fromBalance decimal.Decimal
	if err := tx.QueryRow(ctx, `SELECT balance FROM accounts WHERE id=$1`, fromAccountID).Scan(&fromBalance); err != nil {
		return nil, err
	}

	if fromBalance.LessThan(amount) {
		var txnID string
		err := tx.QueryRow(ctx,
			`INSERT INTO transactions (idempotency_key, from_account_id, to_account_id, amount, status, failure_reason)
			 VALUES ($1, $2, $3, $4, 'failed', 'insufficient balance')
			 RETURNING id`,
			idempotencyKey, fromAccountID, toAccountID, amount,
		).Scan(&txnID)
		if err != nil {
			return nil, err
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return nil, ErrInsufficientBalance
	}

	var txnID string
	err = tx.QueryRow(ctx,
		`INSERT INTO transactions (idempotency_key, from_account_id, to_account_id, amount, status)
		 VALUES ($1, $2, $3, $4, 'completed')
		 RETURNING id`,
		idempotencyKey, fromAccountID, toAccountID, amount,
	).Scan(&txnID)
	if err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx,
		`UPDATE accounts SET balance = balance - $1, updated_at = now() WHERE id=$2`,
		amount, fromAccountID,
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
		txnID, fromAccountID, amount,
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
