package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Shreyansh1032/payments-ledger-system/internal/middleware"
	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
)

type TOTPHandler struct {
	totpService *service.TOTPService
}

func NewTOTPHandler(s *service.TOTPService) *TOTPHandler {
	return &TOTPHandler{totpService: s}
}

func (h *TOTPHandler) BeginSetup(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	username, _ := r.Context().Value(middleware.UsernameKey).(string)

	key, err := h.totpService.BeginSetup(r.Context(), userID, username)
	if err != nil {
		if errors.Is(err, service.ErrTOTPAlreadyEnabled) {
			writeError(w, http.StatusConflict, "two-factor authentication is already enabled")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not start 2FA setup")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"secret":      key.Secret(),
		"otpauth_url": key.URL(),
	})
}

type totpCodeRequest struct {
	Code string `json:"code"`
}

func (h *TOTPHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req totpCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.totpService.Confirm(r.Context(), userID, req.Code); err != nil {
		writeError(w, http.StatusBadRequest, "invalid code")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "two-factor authentication enabled"})
}

func (h *TOTPHandler) Disable(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req totpCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.totpService.Disable(r.Context(), userID, req.Code); err != nil {
		writeError(w, http.StatusBadRequest, "invalid code")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "two-factor authentication disabled"})
}
