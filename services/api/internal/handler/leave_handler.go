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
	"github.com/eprisi/absensi-next/services/api/internal/platform/authtoken"
	"github.com/eprisi/absensi-next/services/api/internal/usecase/leave"
	"github.com/eprisi/absensi-next/services/api/internal/usecase/notification"
)

type LeaveHandler struct {
	service       leave.Service
	notifications notification.Service
}

func NewLeaveHandler(service leave.Service, notifications notification.Service) LeaveHandler {
	return LeaveHandler{service: service, notifications: notifications}
}

type submitLeaveRequest struct {
	Type      string `json:"type"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Reason    string `json:"reason"`
}

func (h LeaveHandler) Submit(w http.ResponseWriter, r *http.Request) {
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

	var req submitLeaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	leaveType := domain.LeaveType(req.Type)
	if leaveType != domain.LeaveTypeIzin && leaveType != domain.LeaveTypeSakit {
		writeError(w, http.StatusUnprocessableEntity, "type must be 'izin' or 'sakit'")
		return
	}
	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "start_date must be YYYY-MM-DD")
		return
	}
	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "end_date must be YYYY-MM-DD")
		return
	}

	lr, err := h.service.Submit(r.Context(), leave.SubmitInput{
		EmployeeID: employeeID,
		Type:       leaveType,
		StartDate:  start,
		EndDate:    end,
		Reason:     req.Reason,
	})
	if err != nil {
		if errors.Is(err, leave.ErrInvalidDateRange) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to submit leave request")
		return
	}

	writeJSON(w, http.StatusCreated, lr)
}

func (h LeaveHandler) ListOwn(w http.ResponseWriter, r *http.Request) {
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

	list, err := h.service.ListOwn(r.Context(), employeeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list leave requests")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h LeaveHandler) ListForAdmin(w http.ResponseWriter, r *http.Request) {
	status := domain.LeaveStatus(r.URL.Query().Get("status"))
	list, err := h.service.ListForAdmin(r.Context(), status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list leave requests")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

type reviewRequest struct {
	Decision string `json:"decision"`
}

func (h LeaveHandler) Review(w http.ResponseWriter, r *http.Request) {
	claims, ok := appmiddleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	reviewerID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid token subject")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req reviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	lr, err := h.service.Review(r.Context(), id, reviewerID, req.Decision)
	if err != nil {
		switch {
		case errors.Is(err, leave.ErrNotFound):
			writeError(w, http.StatusNotFound, "leave request not found")
		case errors.Is(err, leave.ErrAlreadyReviewed):
			writeError(w, http.StatusConflict, "leave request already reviewed")
		case errors.Is(err, leave.ErrInvalidDecision):
			writeError(w, http.StatusUnprocessableEntity, "decision must be 'approved' or 'rejected'")
		default:
			writeError(w, http.StatusInternalServerError, "failed to review leave request")
		}
		return
	}

	// Single real notification trigger required for the B6 approval workflow
	// (D-15): the employee who submitted the request gets notified the
	// moment its status changes. Best-effort -- a notification failure must
	// never roll back or fail the review itself, which already succeeded.
	title := "Pengajuan izin/sakit disetujui"
	if lr.Status == domain.LeaveStatusRejected {
		title = "Pengajuan izin/sakit ditolak"
	}
	_ = h.notifications.Notify(r.Context(), authtoken.AudienceEmployee, lr.EmployeeID,
		notification.TypeLeaveStatusChange, title, "", "/leave")

	writeJSON(w, http.StatusOK, lr)
}
