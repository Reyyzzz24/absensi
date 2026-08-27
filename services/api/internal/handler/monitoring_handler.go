package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/eprisi/absensi-next/services/api/internal/platform/storage"
	"github.com/eprisi/absensi-next/services/api/internal/usecase/monitoring"
)

type MonitoringHandler struct {
	service monitoring.Service
}

func NewMonitoringHandler(service monitoring.Service) MonitoringHandler {
	return MonitoringHandler{service: service}
}

// List returns every attendance record for the given date (default today),
// across all employees -- admin-only, RBAC enforced by route middleware.
func (h MonitoringHandler) List(w http.ResponseWriter, r *http.Request) {
	dateParam := r.URL.Query().Get("date")
	date := time.Now()
	if dateParam != "" {
		parsed, err := time.Parse("2006-01-02", dateParam)
		if err != nil {
			writeError(w, http.StatusBadRequest, "date must be YYYY-MM-DD")
			return
		}
		date = parsed
	}
	y, m, d := date.Date()
	date = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)

	list, err := h.service.ListByDate(r.Context(), date)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list attendance")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h MonitoringHandler) Photo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	kind := r.URL.Query().Get("type")

	bytes, err := h.service.Photo(r.Context(), id, kind)
	if err != nil {
		switch {
		case errors.Is(err, monitoring.ErrInvalidKind):
			writeError(w, http.StatusBadRequest, "type must be 'in' or 'out'")
		case errors.Is(err, monitoring.ErrNotFound):
			writeError(w, http.StatusNotFound, "attendance record not found")
		case errors.Is(err, monitoring.ErrNoPhoto):
			writeError(w, http.StatusNotFound, "no photo recorded for this event")
		case errors.Is(err, storage.ErrInvalidPath):
			writeError(w, http.StatusInternalServerError, "invalid photo path")
		default:
			writeError(w, http.StatusInternalServerError, "failed to load photo")
		}
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(bytes)
}
