package handler

import (
	"net/http"
	"strconv"

	"github.com/eprisi/absensi-next/services/api/internal/usecase/recap"
)

type RecapHandler struct {
	service recap.Service
}

func NewRecapHandler(service recap.Service) RecapHandler {
	return RecapHandler{service: service}
}

func (h RecapHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	var employeeID int64
	if v := r.URL.Query().Get("employee_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid employee_id")
			return
		}
		employeeID = id
	}

	result, err := h.service.Generate(r.Context(), year, month, employeeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate recap")
		return
	}
	writeJSON(w, http.StatusOK, result)
}
