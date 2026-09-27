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

type PaymentRequestHandler struct {
	requestService *service.PaymentRequestService
}

func NewPaymentRequestHandler(requestService *service.PaymentRequestService) *PaymentRequestHandler {
	return &PaymentRequestHandler{requestService: requestService}
}

type createRequestBody struct {
	PayerUsername string `json:"payer_username"`
	Amount        string `json:"amount"`
	Note          string `json:"note"`
}

func (h *PaymentRequestHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid amount")
		return
	}

	result, err := h.requestService.Create(r.Context(), userID, req.PayerUsername, req.Note, amount)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRecipientNotFound):
			writeError(w, http.StatusNotFound, "user not found")
		case errors.Is(err, service.ErrCannotRequestFromSelf):
			writeError(w, http.StatusBadRequest, "cannot request money from yourself")
		case errors.Is(err, service.ErrInvalidAmount):
			writeError(w, http.StatusBadRequest, "amount must be positive")
		default:
			writeError(w, http.StatusInternalServerError, "could not create request")
		}
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

type createSplitBody struct {
	Participants []string `json:"participants"`
	TotalAmount  string   `json:"total_amount"`
	Note         string   `json:"note"`
}

func (h *PaymentRequestHandler) CreateSplit(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createSplitBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	total, err := decimal.NewFromString(req.TotalAmount)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid amount")
		return
	}

	results, err := h.requestService.CreateSplit(r.Context(), userID, req.Participants, total, req.Note)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNoParticipants):
			writeError(w, http.StatusBadRequest, "at least one participant is required")
		case errors.Is(err, service.ErrRecipientNotFound):
			writeError(w, http.StatusNotFound, "one of the participants was not found")
		case errors.Is(err, service.ErrCannotRequestFromSelf):
			writeError(w, http.StatusBadRequest, "cannot include yourself as a participant")
		case errors.Is(err, service.ErrInvalidAmount):
			writeError(w, http.StatusBadRequest, "amount must be positive")
		default:
			writeError(w, http.StatusInternalServerError, "could not create split")
		}
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"requests": results})
}

func (h *PaymentRequestHandler) ListIncoming(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	requests, err := h.requestService.ListIncoming(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not fetch requests")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"requests": requests})
}

func (h *PaymentRequestHandler) ListOutgoing(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	requests, err := h.requestService.ListOutgoing(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not fetch requests")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"requests": requests})
}

func (h *PaymentRequestHandler) Approve(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	requestID := chi.URLParam(r, "id")
	if err := h.requestService.Approve(r.Context(), requestID, userID); err != nil {
		switch {
		case errors.Is(err, service.ErrPaymentRequestNotFound):
			writeError(w, http.StatusNotFound, "request not found")
		case errors.Is(err, service.ErrPaymentRequestNotPending):
			writeError(w, http.StatusBadRequest, "request already resolved")
		case errors.Is(err, service.ErrInsufficientBalance):
			writeError(w, http.StatusBadRequest, "insufficient balance")
		default:
			writeError(w, http.StatusInternalServerError, "could not approve request")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "approved"})
}

func (h *PaymentRequestHandler) Decline(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	requestID := chi.URLParam(r, "id")
	if err := h.requestService.Decline(r.Context(), requestID, userID); err != nil {
		if errors.Is(err, service.ErrPaymentRequestNotFound) {
			writeError(w, http.StatusNotFound, "request not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not decline request")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "declined"})
}
