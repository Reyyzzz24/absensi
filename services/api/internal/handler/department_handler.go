package handler

import (
	"encoding/json"
	"net/http"

	"github.com/eprisi/absensi-next/services/api/internal/usecase/department"
)

type DepartmentHandler struct {
	service department.Service
}

func NewDepartmentHandler(service department.Service) DepartmentHandler {
	return DepartmentHandler{service: service}
}

func (h DepartmentHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list departments")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

type departmentRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func (h DepartmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req departmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" || req.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "code and name are required")
		return
	}

	d, err := h.service.Create(r.Context(), department.DepartmentInput{Code: req.Code, Name: req.Name})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create department")
		return
	}
	writeJSON(w, http.StatusCreated, d)
}
