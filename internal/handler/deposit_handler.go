package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/shopspring/decimal"

	"github.com/Shreyansh1032/payments-ledger-system/internal/middleware"
	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
)

type DepositHandler struct {
	transferService *service.TransferService
}

func NewDepositHandler(transferService *service.TransferService) *DepositHandler {
	return &DepositHandler{transferService: transferService}
}

type depositRequest struct {
	Amount string `json:"amount"`
}

func (h *DepositHandler) Deposit(w http.ResponseWriter, r *http.Request) {
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

	var req depositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid amount")
		return
	}

	result, err := h.transferService.Deposit(r.Context(), idempotencyKey, userID, amount)
	if err != nil {
		if errors.Is(err, service.ErrInvalidAmount) {
			writeError(w, http.StatusBadRequest, "amount must be positive")
			return
		}
		writeError(w, http.StatusInternalServerError, "deposit failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"transaction_id": result.TransactionID,
		"status":         result.Status,
	})
}
