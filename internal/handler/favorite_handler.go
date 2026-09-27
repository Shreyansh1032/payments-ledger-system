package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Shreyansh1032/payments-ledger-system/internal/middleware"
	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
)

type FavoriteHandler struct {
	favoriteService *service.FavoriteService
}

func NewFavoriteHandler(favoriteService *service.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{favoriteService: favoriteService}
}

type favoriteRequest struct {
	Username string `json:"username"`
}

func (h *FavoriteHandler) Add(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req favoriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Username == "" {
		writeError(w, http.StatusBadRequest, "username is required")
		return
	}

	if err := h.favoriteService.Add(r.Context(), userID, req.Username); err != nil {
		switch {
		case errors.Is(err, service.ErrRecipientNotFound):
			writeError(w, http.StatusNotFound, "user not found")
		case errors.Is(err, service.ErrCannotFavoriteSelf):
			writeError(w, http.StatusBadRequest, "cannot favorite yourself")
		default:
			writeError(w, http.StatusInternalServerError, "could not add favorite")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "added"})
}

func (h *FavoriteHandler) Remove(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	username := chi.URLParam(r, "username")
	if err := h.favoriteService.Remove(r.Context(), userID, username); err != nil {
		writeError(w, http.StatusInternalServerError, "could not remove favorite")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

func (h *FavoriteHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	favorites, err := h.favoriteService.List(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not fetch favorites")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"favorites": favorites})
}
