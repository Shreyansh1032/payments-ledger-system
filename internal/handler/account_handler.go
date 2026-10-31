package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Shreyansh1032/payments-ledger-system/internal/middleware"
	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
)

type AccountHandler struct {
	accountService *service.AccountService
	authService    *service.AuthService
	totpService    *service.TOTPService
}

func NewAccountHandler(accountService *service.AccountService, authService *service.AuthService, totpService *service.TOTPService) *AccountHandler {
	return &AccountHandler{accountService: accountService, authService: authService, totpService: totpService}
}

func (h *AccountHandler) Balance(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	balance, err := h.accountService.GetBalance(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not fetch balance")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"balance": balance.String()})
}

func (h *AccountHandler) Profile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	profile, err := h.accountService.GetProfile(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not fetch profile")
		return
	}

	totpEnabled, err := h.totpService.IsEnabled(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not fetch profile")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":           profile.ID,
		"username":     profile.Username,
		"first_name":   profile.FirstName,
		"last_name":    profile.LastName,
		"created_at":   profile.CreatedAt,
		"totp_enabled": totpEnabled,
	})
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (h *AccountHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.NewPassword) < 6 {
		writeError(w, http.StatusBadRequest, "new password must be at least 6 characters")
		return
	}

	if err := h.authService.ChangePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "current password is incorrect")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not change password")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "password updated"})
}
