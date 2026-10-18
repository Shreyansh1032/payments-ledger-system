package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Shreyansh1032/payments-ledger-system/internal/middleware"
	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
)

type EmailVerificationHandler struct {
	verificationService *service.EmailVerificationService
	appBaseURL          string
}

func NewEmailVerificationHandler(s *service.EmailVerificationService, appBaseURL string) *EmailVerificationHandler {
	return &EmailVerificationHandler{verificationService: s, appBaseURL: appBaseURL}
}

type setEmailRequest struct {
	Email string `json:"email"`
}

func (h *EmailVerificationHandler) SetEmail(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req setEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		writeError(w, http.StatusBadRequest, "email is required")
		return
	}

	if err := h.verificationService.SetEmail(r.Context(), userID, req.Email, h.appBaseURL); err != nil {
		if errors.Is(err, service.ErrEmailAlreadyInUse) {
			writeError(w, http.StatusConflict, "email already in use")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not update email")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "verification email sent"})
}

type verifyEmailRequest struct {
	Token string `json:"token"`
}

func (h *EmailVerificationHandler) Verify(w http.ResponseWriter, r *http.Request) {
	var req verifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Token == "" {
		writeError(w, http.StatusBadRequest, "token is required")
		return
	}

	if err := h.verificationService.Verify(r.Context(), req.Token); err != nil {
		if errors.Is(err, service.ErrInvalidVerificationToken) {
			writeError(w, http.StatusBadRequest, "invalid or expired verification link")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not verify email")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "email verified"})
}
