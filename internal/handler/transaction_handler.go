package handler

import (
	"net/http"

	"github.com/Shreyansh1032/payments-ledger-system/internal/middleware"
	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
)

type TransactionHandler struct {
	historyService *service.TransactionHistoryService
}

func NewTransactionHandler(historyService *service.TransactionHistoryService) *TransactionHandler {
	return &TransactionHandler{historyService: historyService}
}

func (h *TransactionHandler) History(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	entries, err := h.historyService.GetHistory(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not fetch transaction history")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"transactions": entries})
}
