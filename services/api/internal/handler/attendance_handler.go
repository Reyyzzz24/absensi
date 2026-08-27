package handler

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	appmiddleware "github.com/eprisi/absensi-next/services/api/internal/middleware"
	"github.com/eprisi/absensi-next/services/api/internal/platform/storage"
	"github.com/eprisi/absensi-next/services/api/internal/usecase/attendance"
)

type AttendanceHandler struct {
	service attendance.Service
	photos  storage.LocalStore
}

func NewAttendanceHandler(service attendance.Service, photos storage.LocalStore) AttendanceHandler {
	return AttendanceHandler{service: service, photos: photos}
}

type checkInRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Photo     string  `json:"photo"` // base64, validated below (legacy accepted malformed data unvalidated -- LOGIC_SPEC.md §11)
	IsWFH     bool    `json:"is_wfh"`
}

func (h AttendanceHandler) CheckIn(w http.ResponseWriter, r *http.Request) {
	claims, ok := appmiddleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	employeeID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid token subject")
		return
	}

	var req checkInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Photo == "" {
		writeError(w, http.StatusUnprocessableEntity, "photo is required")
		return
	}
	photoBytes, err := base64.StdEncoding.DecodeString(req.Photo)
	if err != nil || len(photoBytes) == 0 {
		writeError(w, http.StatusUnprocessableEntity, "photo must be valid base64 image data")
		return
	}
	if req.Latitude < -90 || req.Latitude > 90 || req.Longitude < -180 || req.Longitude > 180 {
		writeError(w, http.StatusUnprocessableEntity, "invalid coordinates")
		return
	}

	// Saved before the geofence/idempotency checks run inside the usecase --
	// a rejected attempt (outside radius, cooldown, etc.) leaves an orphaned
	// file on disk. Acceptable for local disk in dev; revisit if/when this
	// moves to object storage with lifecycle rules (Phase 5, not decided yet).
	photoPath, err := h.photos.SaveCheckInPhoto(employeeID, photoBytes)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to store photo")
		return
	}

	result, err := h.service.CheckInOrOut(r.Context(), attendance.CheckInInput{
		EmployeeID: employeeID,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		PhotoPath:  photoPath,
		IsWFH:      req.IsWFH,
	})
	if err != nil {
		switch {
		case errors.Is(err, attendance.ErrOutsideGeofence):
			writeError(w, http.StatusUnprocessableEntity, "outside allowed check-in radius")
		case errors.Is(err, attendance.ErrAlreadyCheckedOut):
			writeError(w, http.StatusConflict, "already checked out for this shift")
		case errors.Is(err, attendance.ErrCooldownActive):
			writeError(w, http.StatusConflict, "too soon after previous checkout to start a new cycle")
		default:
			writeError(w, http.StatusInternalServerError, "check-in failed")
		}
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// Today returns null when the employee has no attendance recorded yet
// today (and no still-open overnight cycle from yesterday) -- that's a
// valid "haven't checked in yet" state, not an error.
func (h AttendanceHandler) Today(w http.ResponseWriter, r *http.Request) {
	claims, ok := appmiddleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	employeeID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid token subject")
		return
	}

	result, err := h.service.Today(r.Context(), employeeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load today's attendance")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// Geofence returns the active office locations so the employee's check-in
// page can render them on a map (marker + radius circle) and show a
// display-only distance hint before submitting. Read-only, no admin-only
// fields -- safe for the employee audience.
func (h AttendanceHandler) Geofence(w http.ResponseWriter, r *http.Request) {
	locations, err := h.service.ActiveGeofences(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load geofence config")
		return
	}
	writeJSON(w, http.StatusOK, locations)
}
