package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/eprisi/absensi-next/services/api/internal/usecase/employee"
)

type EmployeeHandler struct {
	service employee.Service
}

func NewEmployeeHandler(service employee.Service) EmployeeHandler {
	return EmployeeHandler{service: service}
}

func (h EmployeeHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list employees")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

type employeeRequest struct {
	NIK          string `json:"nik"`
	FullName     string `json:"full_name"`
	Password     string `json:"password"`
	DepartmentID *int64 `json:"department_id"`
	Position     string `json:"position"`
	Phone        string `json:"phone"`
}

func (h EmployeeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req employeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil ||
		req.NIK == "" || req.FullName == "" || req.Password == "" {
		writeError(w, http.StatusUnprocessableEntity, "nik, full_name and password are required")
		return
	}

	e, err := h.service.Create(r.Context(), employee.CreateInput{
		NIK:          req.NIK,
		FullName:     req.FullName,
		Password:     req.Password,
		DepartmentID: req.DepartmentID,
		Position:     req.Position,
		Phone:        req.Phone,
	})
	if err != nil {
		if errors.Is(err, employee.ErrNIKTaken) {
			writeError(w, http.StatusConflict, "nik already registered")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create employee")
		return
	}
	writeJSON(w, http.StatusCreated, e)
}

type employeeUpdateRequest struct {
	FullName     string `json:"full_name"`
	Password     string `json:"password"`
	DepartmentID *int64 `json:"department_id"`
	Position     string `json:"position"`
	Phone        string `json:"phone"`
	IsActive     bool   `json:"is_active"`
}

func (h EmployeeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req employeeUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.FullName == "" {
		writeError(w, http.StatusUnprocessableEntity, "full_name is required")
		return
	}

	e, err := h.service.Update(r.Context(), id, employee.UpdateInput{
		FullName:     req.FullName,
		Password:     req.Password,
		DepartmentID: req.DepartmentID,
		Position:     req.Position,
		Phone:        req.Phone,
		IsActive:     req.IsActive,
	})
	if err != nil {
		if errors.Is(err, employee.ErrNotFound) {
			writeError(w, http.StatusNotFound, "employee not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update employee")
		return
	}
	writeJSON(w, http.StatusOK, e)
}
