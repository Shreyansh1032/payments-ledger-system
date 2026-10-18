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

type SavingsGoalHandler struct {
	goalService *service.SavingsGoalService
}

func NewSavingsGoalHandler(s *service.SavingsGoalService) *SavingsGoalHandler {
	return &SavingsGoalHandler{goalService: s}
}

type createGoalRequest struct {
	Name         string `json:"name"`
	TargetAmount string `json:"target_amount"`
}

func (h *SavingsGoalHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		writeError(w, http.StatusBadRequest, "name and target_amount are required")
		return
	}

	target, err := decimal.NewFromString(req.TargetAmount)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid target amount")
		return
	}

	goal, err := h.goalService.Create(r.Context(), userID, req.Name, target)
	if err != nil {
		if errors.Is(err, service.ErrInvalidAmount) {
			writeError(w, http.StatusBadRequest, "target amount must be positive")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not create goal")
		return
	}

	writeJSON(w, http.StatusCreated, goal)
}

func (h *SavingsGoalHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	goals, err := h.goalService.List(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not fetch goals")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"goals": goals})
}

type contributeRequest struct {
	Amount string `json:"amount"`
}

func (h *SavingsGoalHandler) Contribute(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	goalID := chi.URLParam(r, "id")

	var req contributeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid amount")
		return
	}

	goal, justReached, err := h.goalService.AddContribution(r.Context(), userID, goalID, amount)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSavingsGoalNotFound):
			writeError(w, http.StatusNotFound, "goal not found")
		case errors.Is(err, service.ErrInvalidAmount):
			writeError(w, http.StatusBadRequest, "amount must be positive")
		default:
			writeError(w, http.StatusInternalServerError, "could not add contribution")
		}
		return
	}

	if justReached {
		_ = h.goalService.NotifyGoalReached(r.Context(), userID, goal.Name)
	}

	writeJSON(w, http.StatusOK, goal)
}

func (h *SavingsGoalHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	goalID := chi.URLParam(r, "id")
	if err := h.goalService.Delete(r.Context(), userID, goalID); err != nil {
		if errors.Is(err, service.ErrSavingsGoalNotFound) {
			writeError(w, http.StatusNotFound, "goal not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not delete goal")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
