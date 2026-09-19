package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/shopspring/decimal"

	"github.com/Shreyansh1032/payments-ledger-system/internal/middleware"
	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
)

type TransferHandler struct {
	transferService *service.TransferService
}

func NewTransferHandler(transferService *service.TransferService) *TransferHandler {
	return &TransferHandler{transferService: transferService}
}

type transferRequest struct {
	ToUsername string `json:"to_username"`
	Amount     string `json:"amount"`
}

func (h *TransferHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idempotencyKey := r.Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		writeError(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	var req transferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid amount")
		return
	}

	result, err := h.transferService.Transfer(r.Context(), idempotencyKey, userID, req.ToUsername, amount)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRecipientNotFound):
			middleware.TransfersTotal.WithLabelValues("failed_recipient_not_found").Inc()
			writeError(w, http.StatusNotFound, "recipient not found")
		case errors.Is(err, service.ErrInsufficientBalance):
			middleware.TransfersTotal.WithLabelValues("failed_insufficient_balance").Inc()
			writeError(w, http.StatusBadRequest, "insufficient balance")
		case errors.Is(err, service.ErrSelfTransfer):
			middleware.TransfersTotal.WithLabelValues("failed_self_transfer").Inc()
			writeError(w, http.StatusBadRequest, "cannot transfer to yourself")
		case errors.Is(err, service.ErrInvalidAmount):
			middleware.TransfersTotal.WithLabelValues("failed_invalid_amount").Inc()
			writeError(w, http.StatusBadRequest, "amount must be positive")
		default:
			middleware.TransfersTotal.WithLabelValues("failed_internal").Inc()
			writeError(w, http.StatusInternalServerError, "transfer failed")
		}
		return
	}

	middleware.TransfersTotal.WithLabelValues(result.Status).Inc()

	writeJSON(w, http.StatusOK, map[string]string{
		"transaction_id": result.TransactionID,
		"status":         result.Status,
	})
}
