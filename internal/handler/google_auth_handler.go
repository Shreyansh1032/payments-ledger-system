package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
)

type GoogleAuthHandler struct {
	googleAuthService   *service.GoogleAuthService
	refreshTokenService *service.RefreshTokenService
	jwtSecret           string
}

func NewGoogleAuthHandler(g *service.GoogleAuthService, r *service.RefreshTokenService, jwtSecret string) *GoogleAuthHandler {
	return &GoogleAuthHandler{googleAuthService: g, refreshTokenService: r, jwtSecret: jwtSecret}
}

type googleSignInRequest struct {
	IDToken string `json:"id_token"`
}

func (h *GoogleAuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	var req googleSignInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.IDToken == "" {
		writeError(w, http.StatusBadRequest, "id_token is required")
		return
	}

	user, isNewUser, err := h.googleAuthService.SignIn(r.Context(), req.IDToken)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrGoogleTokenInvalid):
			writeError(w, http.StatusUnauthorized, "invalid google token")
		case errors.Is(err, service.ErrGoogleEmailNotVerified):
			writeError(w, http.StatusBadRequest, "google account email is not verified")
		default:
			writeError(w, http.StatusInternalServerError, "could not sign in with google")
		}
		return
	}

	accessToken, err := service.GenerateAccessToken(h.jwtSecret, user.ID, user.Username)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not generate token")
		return
	}
	refreshToken, err := h.refreshTokenService.Issue(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not generate refresh token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"username":      user.Username,
		"first_name":    user.FirstName,
		"is_new_user":   isNewUser,
	})
}
