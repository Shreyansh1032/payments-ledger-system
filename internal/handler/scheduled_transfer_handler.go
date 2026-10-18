package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"

	"github.com/Shreyansh1032/payments-ledger-system/internal/middleware"
	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
)

type ScheduledTransferHandler struct {
	scheduledService *service.ScheduledTransferService
}

func NewScheduledTransferHandler(s *service.ScheduledTransferService) *ScheduledTransferHandler {
	return &ScheduledTransferHandler{scheduledService: s}
}

type createScheduledTransferRequest struct {
	ToUsername string `json:"to_username"`
	Amount     string `json:"amount"`
	Frequency  string `json:"frequency"`
}

func (h *ScheduledTransferHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createScheduledTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid amount")
		return
	}

	st, err := h.scheduledService.Create(r.Context(), userID, req.ToUsername, req.Frequency, amount)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRecipientNotFound):
			writeError(w, http.StatusNotFound, "recipient not found")
		case errors.Is(err, service.ErrInvalidFrequency):
			writeError(w, http.StatusBadRequest, "frequency must be daily, weekly, or monthly")
		case errors.Is(err, service.ErrInvalidAmount):
			writeError(w, http.StatusBadRequest, "amount must be positive")
		default:
			writeError(w, http.StatusInternalServerError, "could not create scheduled transfer")
		}
		return
	}

	writeJSON(w, http.StatusCreated, st)
}

func (h *ScheduledTransferHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	list, err := h.scheduledService.List(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not fetch scheduled transfers")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"scheduled_transfers": list})
}

func (h *ScheduledTransferHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := chi.URLParam(r, "id")
	if err := h.scheduledService.Cancel(r.Context(), userID, id); err != nil {
		if errors.Is(err, service.ErrScheduledTransferNotFound) {
			writeError(w, http.StatusNotFound, "scheduled transfer not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not cancel scheduled transfer")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}
