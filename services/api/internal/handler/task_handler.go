package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	appmiddleware "github.com/eprisi/absensi-next/services/api/internal/middleware"
	"github.com/eprisi/absensi-next/services/api/internal/usecase/task"
)

type TaskHandler struct {
	service task.Service
}

func NewTaskHandler(service task.Service) TaskHandler {
	return TaskHandler{service: service}
}

type taskRequest struct {
	Title    string     `json:"title"`
	Detail   string     `json:"detail"`
	StartsAt time.Time  `json:"starts_at"`
	EndsAt   *time.Time `json:"ends_at"`
	Status   string     `json:"status"`
}

func currentEmployeeID(r *http.Request) (int64, bool) {
	claims, ok := appmiddleware.ClaimsFromContext(r.Context())
	if !ok {
		return 0, false
	}
	id, err := strconv.ParseInt(claims.Subject, 10, 64)
	return id, err == nil
}

func (h TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	employeeID, ok := currentEmployeeID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	var req taskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" || req.StartsAt.IsZero() {
		writeError(w, http.StatusUnprocessableEntity, "title and starts_at are required")
		return
	}

	t, err := h.service.Create(r.Context(), employeeID, task.TaskInput{
		Title:    req.Title,
		Detail:   req.Detail,
		StartsAt: req.StartsAt,
		EndsAt:   req.EndsAt,
		Status:   req.Status,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create task")
		return
	}

	writeJSON(w, http.StatusCreated, t)
}

func (h TaskHandler) ListOwn(w http.ResponseWriter, r *http.Request) {
	employeeID, ok := currentEmployeeID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	list, err := h.service.ListOwn(r.Context(), employeeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list tasks")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// Update enforces ownership via task.Service -- fixes the legacy IDOR where
// edit_task/update_task never checked whether the task belonged to the
// caller (A5/D-5). A task that exists but belongs to someone else returns
// the same 404 as a task that doesn't exist at all.
func (h TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	employeeID, ok := currentEmployeeID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req taskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" || req.StartsAt.IsZero() {
		writeError(w, http.StatusUnprocessableEntity, "title and starts_at are required")
		return
	}

	t, err := h.service.Update(r.Context(), id, employeeID, task.TaskInput{
		Title:    req.Title,
		Detail:   req.Detail,
		StartsAt: req.StartsAt,
		EndsAt:   req.EndsAt,
		Status:   req.Status,
	})
	if err != nil {
		if errors.Is(err, task.ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update task")
		return
	}

	writeJSON(w, http.StatusOK, t)
}
