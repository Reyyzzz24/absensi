package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/eprisi/absensi-next/services/api/internal/usecase/notification"
)

type NotificationHandler struct {
	service notification.Service
}

func NewNotificationHandler(service notification.Service) NotificationHandler {
	return NotificationHandler{service: service}
}

func (h NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	aud, id, ok := identity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	page := 1
	if v := r.URL.Query().Get("page"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			page = parsed
		}
	}
	const pageSize = 20
	list, err := h.service.List(r.Context(), authAudience(aud), id, pageSize, (page-1)*pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load notifications")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h NotificationHandler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	aud, id, ok := identity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	count, err := h.service.UnreadCount(r.Context(), authAudience(aud), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to count notifications")
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"unread_count": count})
}

func (h NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	aud, id, ok := identity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	notifID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.service.MarkRead(r.Context(), authAudience(aud), id, notifID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to mark as read")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h NotificationHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	aud, id, ok := identity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	if err := h.service.MarkAllRead(r.Context(), authAudience(aud), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to mark all as read")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h NotificationHandler) Preferences(w http.ResponseWriter, r *http.Request) {
	aud, id, ok := identity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	prefs, err := h.service.Preferences(r.Context(), authAudience(aud), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load preferences")
		return
	}
	writeJSON(w, http.StatusOK, prefs)
}

type setPreferenceRequest struct {
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
}

func (h NotificationHandler) SetPreference(w http.ResponseWriter, r *http.Request) {
	aud, id, ok := identity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	var req setPreferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Type == "" {
		writeError(w, http.StatusBadRequest, "type is required")
		return
	}
	err := h.service.SetPreference(r.Context(), authAudience(aud), id, req.Type, req.Enabled)
	if err != nil {
		if errors.Is(err, notification.ErrUnknownType) {
			writeError(w, http.StatusBadRequest, "unknown notification type")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to save preference")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
