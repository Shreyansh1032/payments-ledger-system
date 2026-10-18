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

type BudgetHandler struct {
	budgetService *service.BudgetService
}

func NewBudgetHandler(s *service.BudgetService) *BudgetHandler {
	return &BudgetHandler{budgetService: s}
}

type setBudgetRequest struct {
	Category     string `json:"category"`
	MonthlyLimit string `json:"monthly_limit"`
}

func (h *BudgetHandler) SetLimit(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req setBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Category == "" {
		writeError(w, http.StatusBadRequest, "category and monthly_limit are required")
		return
	}

	limit, err := decimal.NewFromString(req.MonthlyLimit)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid monthly limit")
		return
	}

	if err := h.budgetService.SetLimit(r.Context(), userID, req.Category, limit); err != nil {
		if errors.Is(err, service.ErrInvalidAmount) {
			writeError(w, http.StatusBadRequest, "monthly limit must be positive")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not set budget")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (h *BudgetHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	budgets, err := h.budgetService.List(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not fetch budgets")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"budgets": budgets})
}

func (h *BudgetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := chi.URLParam(r, "id")
	if err := h.budgetService.Delete(r.Context(), userID, id); err != nil {
		writeError(w, http.StatusInternalServerError, "could not delete budget")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
