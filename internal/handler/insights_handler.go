package handler

import (
	"net/http"

	"github.com/Shreyansh1032/payments-ledger-system/internal/middleware"
	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
)

type InsightsHandler struct {
	insightsService *service.InsightsService
}

func NewInsightsHandler(insightsService *service.InsightsService) *InsightsHandler {
	return &InsightsHandler{insightsService: insightsService}
}

func (h *InsightsHandler) Monthly(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	insights, err := h.insightsService.GetMonthlyInsights(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not fetch insights")
		return
	}

	writeJSON(w, http.StatusOK, insights)
}
