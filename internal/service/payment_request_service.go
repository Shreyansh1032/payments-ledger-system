package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/Shreyansh1032/payments-ledger-system/internal/model"
)

var (
	ErrPaymentRequestNotFound   = errors.New("payment request not found")
	ErrPaymentRequestNotPending = errors.New("payment request is not pending")
	ErrCannotRequestFromSelf    = errors.New("cannot request money from yourself")
	ErrNoParticipants           = errors.New("at least one participant is required")
)

type PaymentRequestService struct {
	pool            *pgxpool.Pool
	transferService *TransferService
}

func NewPaymentRequestService(pool *pgxpool.Pool, transferService *TransferService) *PaymentRequestService {
	return &PaymentRequestService{pool: pool, transferService: transferService}
}

func (s *PaymentRequestService) Create(ctx context.Context, requesterID, payerUsername, note string, amount decimal.Decimal) (*model.PaymentRequest, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalidAmount
	}

	var payerID string
	err := s.pool.QueryRow(ctx, `SELECT id FROM users WHERE username=$1`, payerUsername).Scan(&payerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRecipientNotFound
		}
		return nil, err
	}

	if payerID == requesterID {
		return nil, ErrCannotRequestFromSelf
	}

	req := &model.PaymentRequest{Amount: amount, Note: note, PayerUsername: payerUsername}
	err = s.pool.QueryRow(ctx, `
		INSERT INTO payment_requests (requester_id, payer_id, amount, note)
		VALUES ($1, $2, $3, NULLIF($4, ''))
		RETURNING id, status, created_at
	`, requesterID, payerID, amount, note).Scan(&req.ID, &req.Status, &req.CreatedAt)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// CreateSplit divides total evenly and requests each participant's share.
// Not wrapped in a single DB transaction: if a later participant is
// invalid, earlier requests in the same split will already exist.
func (s *PaymentRequestService) CreateSplit(ctx context.Context, creatorID string, participantUsernames []string, total decimal.Decimal, note string) ([]model.PaymentRequest, error) {
	if len(participantUsernames) == 0 {
		return nil, ErrNoParticipants
	}
	if total.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalidAmount
	}

	n := int64(len(participantUsernames))
	share := total.Div(decimal.NewFromInt(n)).Round(2)
	remainder := total.Sub(share.Mul(decimal.NewFromInt(n)))

	requests := make([]model.PaymentRequest, 0, n)
	for i, username := range participantUsernames {
		amount := share
		if i == len(participantUsernames)-1 {
			amount = amount.Add(remainder)
		}
		req, err := s.Create(ctx, creatorID, username, note, amount)
		if err != nil {
			return nil, err
		}
		requests = append(requests, *req)
	}
	return requests, nil
}

func (s *PaymentRequestService) ListIncoming(ctx context.Context, payerID string) ([]model.PaymentRequest, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT pr.id, u.username, pr.amount, COALESCE(pr.note, ''), pr.status, pr.created_at
		FROM payment_requests pr
		JOIN users u ON u.id = pr.requester_id
		WHERE pr.payer_id = $1
		ORDER BY pr.created_at DESC
	`, payerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := []model.PaymentRequest{}
	for rows.Next() {
		var r model.PaymentRequest
		if err := rows.Scan(&r.ID, &r.RequesterUsername, &r.Amount, &r.Note, &r.Status, &r.CreatedAt); err != nil {
			return nil, err
		}
		requests = append(requests, r)
	}
	return requests, rows.Err()
}

func (s *PaymentRequestService) ListOutgoing(ctx context.Context, requesterID string) ([]model.PaymentRequest, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT pr.id, u.username, pr.amount, COALESCE(pr.note, ''), pr.status, pr.created_at
		FROM payment_requests pr
		JOIN users u ON u.id = pr.payer_id
		WHERE pr.requester_id = $1
		ORDER BY pr.created_at DESC
	`, requesterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := []model.PaymentRequest{}
	for rows.Next() {
		var r model.PaymentRequest
		if err := rows.Scan(&r.ID, &r.PayerUsername, &r.Amount, &r.Note, &r.Status, &r.CreatedAt); err != nil {
			return nil, err
		}
		requests = append(requests, r)
	}
	return requests, rows.Err()
}

func (s *PaymentRequestService) Approve(ctx context.Context, requestID, payerID string) error {
	var amount decimal.Decimal
	var status, requesterUsername, reqPayerID string
	err := s.pool.QueryRow(ctx, `
		SELECT pr.amount, pr.status, pr.payer_id, u.username
		FROM payment_requests pr
		JOIN users u ON u.id = pr.requester_id
		WHERE pr.id = $1
	`, requestID).Scan(&amount, &status, &reqPayerID, &requesterUsername)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrPaymentRequestNotFound
		}
		return err
	}

	if reqPayerID != payerID {
		return ErrPaymentRequestNotFound
	}
	if status != "pending" {
		return ErrPaymentRequestNotPending
	}

	idempotencyKey := fmt.Sprintf("payment-request-%s", requestID)
	result, err := s.transferService.Transfer(ctx, idempotencyKey, payerID, requesterUsername, amount)
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(ctx, `
		UPDATE payment_requests
		SET status = 'approved', transaction_id = $1, resolved_at = now()
		WHERE id = $2
	`, result.TransactionID, requestID)
	return err
}

func (s *PaymentRequestService) Decline(ctx context.Context, requestID, payerID string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE payment_requests
		SET status = 'declined', resolved_at = now()
		WHERE id = $1 AND payer_id = $2 AND status = 'pending'
	`, requestID, payerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrPaymentRequestNotFound
	}
	return nil
}
