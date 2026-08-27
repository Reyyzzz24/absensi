package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
	appmiddleware "github.com/eprisi/absensi-next/services/api/internal/middleware"
	"github.com/eprisi/absensi-next/services/api/internal/usecase/holiday"
)

type HolidayHandler struct {
	service holiday.Service
}

func NewHolidayHandler(service holiday.Service) HolidayHandler {
	return HolidayHandler{service: service}
}

// --- Combined calendar (all three sources) -- for the admin "Hari Libur"
// page's calendar view, and any future employee weekly-strip integration. ---

type calendarDay struct {
	Date          string `json:"date"`
	IsHoliday     bool   `json:"is_holiday"`
	Source        string `json:"source"`
	Label         string `json:"label,omitempty"`
	IsCutiBersama bool   `json:"is_cuti_bersama,omitempty"`
}

func (h HolidayHandler) Calendar(w http.ResponseWriter, r *http.Request) {
	year, err := strconv.Atoi(r.URL.Query().Get("year"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "year is required (YYYY)")
		return
	}
	month, err := strconv.Atoi(r.URL.Query().Get("month"))
	if err != nil || month < 1 || month > 12 {
		writeError(w, http.StatusBadRequest, "month is required (1-12)")
		return
	}

	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, -1)

	statuses, err := h.service.ResolveRange(r.Context(), start, end)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to resolve calendar")
		return
	}

	days := make([]calendarDay, 0, end.Day())
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		status := statuses[key]
		days = append(days, calendarDay{
			Date:          key,
			IsHoliday:     status.IsHoliday,
			Source:        string(status.Source),
			Label:         status.Label,
			IsCutiBersama: status.IsCutiBersama,
		})
	}
	writeJSON(w, http.StatusOK, days)
}

// --- National holidays (cached, admin-readable, superadmin can sync) ---

func (h HolidayHandler) ListNational(w http.ResponseWriter, r *http.Request) {
	year, err := strconv.Atoi(r.URL.Query().Get("year"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "year is required (YYYY)")
		return
	}
	list, err := h.service.ListNationalHolidays(r.Context(), year)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list national holidays")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// SyncNational is admin-triggered only -- the resolver never calls the
// external source itself (D-25), so this is the sole place a network
// fetch to the calendar source happens, and only on demand.
func (h HolidayHandler) SyncNational(w http.ResponseWriter, r *http.Request) {
	year, err := strconv.Atoi(r.URL.Query().Get("year"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "year is required (YYYY)")
		return
	}
	n, err := h.service.SyncNational(r.Context(), year)
	if err != nil {
		// Graceful degradation: the existing cache is untouched by a failed
		// sync (holiday.Service.SyncNational), so callers keep working --
		// this just reports the sync itself didn't complete this time.
		writeError(w, http.StatusBadGateway, "failed to sync from holiday calendar source: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"synced": n})
}

// --- Company manual holidays (single date or range, superadmin-managed) ---

type companyHolidayRequest struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Note      string `json:"note"`
}

func parseCompanyHolidayRequest(r *http.Request) (holiday.CompanyHolidayInput, error) {
	var req companyHolidayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return holiday.CompanyHolidayInput{}, errors.New("invalid request body")
	}
	if req.Name == "" || req.StartDate == "" {
		return holiday.CompanyHolidayInput{}, errors.New("name and start_date are required")
	}
	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return holiday.CompanyHolidayInput{}, errors.New("start_date must be YYYY-MM-DD")
	}
	end := start
	if req.EndDate != "" {
		end, err = time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return holiday.CompanyHolidayInput{}, errors.New("end_date must be YYYY-MM-DD")
		}
	}
	if end.Before(start) {
		return holiday.CompanyHolidayInput{}, errors.New("end_date must not be before start_date")
	}
	holidayType := domain.CompanyHolidayTypeLibur
	if req.Type == string(domain.CompanyHolidayTypeCutiBersama) {
		holidayType = domain.CompanyHolidayTypeCutiBersama
	}
	return holiday.CompanyHolidayInput{StartDate: start, EndDate: end, Name: req.Name, Type: holidayType, Note: req.Note}, nil
}

func (h HolidayHandler) ListCompany(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.ListCompanyHolidays(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list company holidays")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h HolidayHandler) CreateCompany(w http.ResponseWriter, r *http.Request) {
	in, err := parseCompanyHolidayRequest(r)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	claims, ok := appmiddleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	createdBy, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid token subject")
		return
	}
	in.CreatedBy = createdBy
	created, err := h.service.CreateCompanyHoliday(r.Context(), in)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create company holiday")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h HolidayHandler) UpdateCompany(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	in, err := parseCompanyHolidayRequest(r)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	updated, err := h.service.UpdateCompanyHoliday(r.Context(), id, in)
	if err != nil {
		if errors.Is(err, holiday.ErrNotFound) {
			writeError(w, http.StatusNotFound, "company holiday not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update company holiday")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h HolidayHandler) DeleteCompany(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.service.DeleteCompanyHoliday(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete company holiday")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
