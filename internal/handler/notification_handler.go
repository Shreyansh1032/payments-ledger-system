package handler

import (
	"net/http"

	"github.com/Shreyansh1032/payments-ledger-system/internal/middleware"
	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
)

type NotificationHandler struct {
	notificationService *service.NotificationService
}

func NewNotificationHandler(s *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationService: s}
}

func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	notifications, err := h.notificationService.List(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not fetch notifications")
		return
	}

	unread, err := h.notificationService.UnreadCount(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not fetch notifications")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"notifications": notifications, "unread_count": unread})
}

func (h *NotificationHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.notificationService.MarkAllRead(r.Context(), userID); err != nil {
		writeError(w, http.StatusInternalServerError, "could not update notifications")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}
