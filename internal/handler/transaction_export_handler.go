package handler

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Shreyansh1032/payments-ledger-system/internal/middleware"
	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
)

type TransactionExportHandler struct {
	historyService  *service.TransactionHistoryService
	categoryService *service.TransactionCategoryService
}

func NewTransactionExportHandler(h *service.TransactionHistoryService, c *service.TransactionCategoryService) *TransactionExportHandler {
	return &TransactionExportHandler{historyService: h, categoryService: c}
}

func (h *TransactionExportHandler) Export(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	entries, err := h.historyService.GetHistoryForExport(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not export transactions")
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="transactions.csv"`)

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"date", "direction", "counterparty", "amount", "category", "status"})
	for _, e := range entries {
		_ = cw.Write([]string{
			e.CreatedAt.Format("2006-01-02 15:04"),
			e.Direction,
			e.Counterparty,
			e.Amount.String(),
			e.Category,
			e.Status,
		})
	}
	cw.Flush()
}

type setCategoryRequest struct {
	Category string `json:"category"`
}

func (h *TransactionExportHandler) SetCategory(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := chi.URLParam(r, "id")

	var req setCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Category == "" {
		writeError(w, http.StatusBadRequest, "category is required")
		return
	}

	if err := h.categoryService.SetCategory(r.Context(), userID, id, req.Category); err != nil {
		if errors.Is(err, service.ErrTransactionNotFound) {
			writeError(w, http.StatusNotFound, "transaction not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not update category")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}
