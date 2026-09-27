package handler

import (
	"net/http"

	"github.com/Shreyansh1032/payments-ledger-system/internal/middleware"
	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
)

type UserSearchHandler struct {
	searchService *service.UserSearchService
}

func NewUserSearchHandler(searchService *service.UserSearchService) *UserSearchHandler {
	return &UserSearchHandler{searchService: searchService}
}

func (h *UserSearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	query := r.URL.Query().Get("q")

	results, err := h.searchService.Search(r.Context(), userID, query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "search failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"users": results})
}
