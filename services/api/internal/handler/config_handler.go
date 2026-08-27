package handler

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	appmiddleware "github.com/eprisi/absensi-next/services/api/internal/middleware"
	"github.com/eprisi/absensi-next/services/api/internal/platform/storage"
	"github.com/eprisi/absensi-next/services/api/internal/usecase/config"
)

type ConfigHandler struct {
	service config.Service
	logos   storage.LocalStore
}

func NewConfigHandler(service config.Service, logos storage.LocalStore) ConfigHandler {
	return ConfigHandler{service: service, logos: logos}
}

// --- Company settings ---
// Single-row config (no multi-tenant concept in this app), admin-only read,
// superadmin-only write -- same RBAC tier as other config domains (D-7).

func (h ConfigHandler) GetCompanySettings(w http.ResponseWriter, r *http.Request) {
	c, err := h.service.GetCompanySettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load company settings")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

type companySettingsRequest struct {
	Name string `json:"name"`
}

func (h ConfigHandler) UpdateCompanySettings(w http.ResponseWriter, r *http.Request) {
	var req companySettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	c, err := h.service.UpdateCompanySettings(r.Context(), config.CompanySettingsInput{Name: req.Name})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update company settings")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

type uploadLogoRequest struct {
	Photo string `json:"photo"`
}

func (h ConfigHandler) UploadCompanyLogo(w http.ResponseWriter, r *http.Request) {
	var req uploadLogoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	photoBytes, err := base64.StdEncoding.DecodeString(req.Photo)
	if err != nil || len(photoBytes) == 0 {
		writeError(w, http.StatusUnprocessableEntity, "photo must be valid base64 image data")
		return
	}
	if len(photoBytes) > 2*1024*1024 {
		writeError(w, http.StatusUnprocessableEntity, "photo exceeds 2MB limit")
		return
	}
	contentType := http.DetectContentType(photoBytes)
	if contentType != "image/png" && contentType != "image/jpeg" && contentType != "image/webp" {
		writeError(w, http.StatusUnprocessableEntity, "photo must be PNG, JPEG, or WebP")
		return
	}

	relPath, err := h.logos.SaveAvatar("company", 1, photoBytes)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save logo")
		return
	}
	c, err := h.service.UpdateCompanySettings(r.Context(), config.CompanySettingsInput{
		Name:     mustCurrentCompanyName(r, h),
		LogoPath: &relPath,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update company settings")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

// mustCurrentCompanyName re-reads the existing name so a logo-only upload
// doesn't require the client to resend it (UpdateCompanySettings always
// takes a full Name because it's also used by the name-only PUT above).
func mustCurrentCompanyName(r *http.Request, h ConfigHandler) string {
	existing, err := h.service.GetCompanySettings(r.Context())
	if err != nil || existing == nil {
		return ""
	}
	return existing.Name
}

func (h ConfigHandler) CompanyLogo(w http.ResponseWriter, r *http.Request) {
	c, err := h.service.GetCompanySettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load company settings")
		return
	}
	if c.LogoPath == nil {
		writeError(w, http.StatusNotFound, "no logo set")
		return
	}
	bytes, err := h.logos.ReadPhoto(*c.LogoPath)
	if err != nil {
		if errors.Is(err, storage.ErrInvalidPath) {
			writeError(w, http.StatusInternalServerError, "invalid logo path")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load logo")
		return
	}
	w.Header().Set("Content-Type", http.DetectContentType(bytes))
	w.Write(bytes)
}

// --- Shifts ---

type shiftRequest struct {
	Code             string `json:"code"`
	Name             string `json:"name"`
	IsDayOff         bool   `json:"is_day_off"`
	StartTime        string `json:"start_time"`
	EndTime          string `json:"end_time"`
	LateGraceMinutes int    `json:"late_grace_minutes"`
}

func (h ConfigHandler) ListShifts(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.ListShifts(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list shifts")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h ConfigHandler) CreateShift(w http.ResponseWriter, r *http.Request) {
	var req shiftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" || req.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "code and name are required")
		return
	}
	grace := req.LateGraceMinutes
	if grace <= 0 {
		grace = 15 // parity default with legacy 09:15 threshold, D-10
	}

	shift, err := h.service.CreateShift(r.Context(), config.ShiftInput{
		Code:             req.Code,
		Name:             req.Name,
		IsDayOff:         req.IsDayOff,
		StartTime:        req.StartTime,
		EndTime:          req.EndTime,
		LateGraceMinutes: grace,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create shift")
		return
	}
	writeJSON(w, http.StatusCreated, shift)
}

func (h ConfigHandler) UpdateShift(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req shiftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "name is required")
		return
	}

	shift, err := h.service.UpdateShift(r.Context(), id, config.ShiftInput{
		Name:             req.Name,
		IsDayOff:         req.IsDayOff,
		StartTime:        req.StartTime,
		EndTime:          req.EndTime,
		LateGraceMinutes: req.LateGraceMinutes,
	})
	if err != nil {
		if errors.Is(err, config.ErrNotFound) {
			writeError(w, http.StatusNotFound, "shift not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update shift")
		return
	}
	writeJSON(w, http.StatusOK, shift)
}

// --- Office locations ---

type officeLocationRequest struct {
	Name         string  `json:"name"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	RadiusMeters int     `json:"radius_meters"`
	IsActive     bool    `json:"is_active"`
}

func (h ConfigHandler) ListOfficeLocations(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.ListOfficeLocations(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list office locations")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h ConfigHandler) CreateOfficeLocation(w http.ResponseWriter, r *http.Request) {
	var req officeLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "name is required")
		return
	}

	loc, err := h.service.CreateOfficeLocation(r.Context(), config.OfficeLocationInput{
		Name:         req.Name,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		RadiusMeters: req.RadiusMeters,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create office location")
		return
	}
	writeJSON(w, http.StatusCreated, loc)
}

func (h ConfigHandler) UpdateOfficeLocation(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req officeLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "name is required")
		return
	}

	loc, err := h.service.UpdateOfficeLocation(r.Context(), id, config.OfficeLocationInput{
		Name:         req.Name,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		RadiusMeters: req.RadiusMeters,
		IsActive:     req.IsActive,
	})
	if err != nil {
		if errors.Is(err, config.ErrNotFound) {
			writeError(w, http.StatusNotFound, "office location not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update office location")
		return
	}
	writeJSON(w, http.StatusOK, loc)
}

// --- Shift schedules ---

type workScheduleRequest struct {
	EmployeeID int64  `json:"employee_id"`
	WorkDate   string `json:"work_date"` // YYYY-MM-DD
	ShiftID    int64  `json:"shift_id"`
}

func (h ConfigHandler) SetWorkSchedule(w http.ResponseWriter, r *http.Request) {
	var req workScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.EmployeeID == 0 || req.ShiftID == 0 {
		writeError(w, http.StatusUnprocessableEntity, "employee_id, work_date and shift_id are required")
		return
	}
	date, err := time.Parse("2006-01-02", req.WorkDate)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "work_date must be YYYY-MM-DD")
		return
	}

	if err := h.service.SetWorkSchedule(r.Context(), req.EmployeeID, date, req.ShiftID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to set work schedule")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

type weeklyShiftDefaultRequest struct {
	EmployeeID int64 `json:"employee_id"`
	DayOfWeek  int16 `json:"day_of_week"`
	ShiftID    int64 `json:"shift_id"`
}

func (h ConfigHandler) SetWeeklyShiftDefault(w http.ResponseWriter, r *http.Request) {
	var req weeklyShiftDefaultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.EmployeeID == 0 || req.ShiftID == 0 {
		writeError(w, http.StatusUnprocessableEntity, "employee_id, day_of_week and shift_id are required")
		return
	}
	if req.DayOfWeek < 0 || req.DayOfWeek > 6 {
		writeError(w, http.StatusUnprocessableEntity, "day_of_week must be 0-6")
		return
	}

	if err := h.service.SetWeeklyShiftDefault(r.Context(), req.EmployeeID, req.DayOfWeek, req.ShiftID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to set weekly shift default")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

// --- Field assignments ("dinas luar", D-21) ---

type fieldAssignmentRequest struct {
	EmployeeID int64  `json:"employee_id"`
	WorkDate   string `json:"work_date"`
	Note       string `json:"note"`
}

func (h ConfigHandler) CreateFieldAssignment(w http.ResponseWriter, r *http.Request) {
	claims, ok := appmiddleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	approvedBy, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid token subject")
		return
	}

	var req fieldAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.EmployeeID == 0 {
		writeError(w, http.StatusUnprocessableEntity, "employee_id and work_date are required")
		return
	}
	date, err := time.Parse("2006-01-02", req.WorkDate)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "work_date must be YYYY-MM-DD")
		return
	}

	fa, err := h.service.CreateFieldAssignment(r.Context(), config.FieldAssignmentInput{
		EmployeeID: req.EmployeeID,
		WorkDate:   date,
		Note:       req.Note,
		ApprovedBy: approvedBy,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create field assignment")
		return
	}
	writeJSON(w, http.StatusCreated, fa)
}

func (h ConfigHandler) ListFieldAssignments(w http.ResponseWriter, r *http.Request) {
	var employeeID int64
	if v := r.URL.Query().Get("employee_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid employee_id")
			return
		}
		employeeID = id
	}

	list, err := h.service.ListFieldAssignments(r.Context(), employeeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list field assignments")
		return
	}
	writeJSON(w, http.StatusOK, list)
}
