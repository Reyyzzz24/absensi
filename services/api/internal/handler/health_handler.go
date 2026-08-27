package handler

import (
	"database/sql"
	"net/http"
)

// HealthHandler is infra-only: it proves the API process is up and can reach
// Postgres. No business logic -- used to verify the Dockerized dev stack
// boots cleanly (docs/PROGRESS.md, dev-environment bring-up ahead of Phase 5).
type HealthHandler struct {
	sqlDB *sql.DB
}

func NewHealthHandler(sqlDB *sql.DB) HealthHandler {
	return HealthHandler{sqlDB: sqlDB}
}

func (h HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	dbStatus := "ok"
	status := "ok"
	httpStatus := http.StatusOK

	if err := h.sqlDB.PingContext(r.Context()); err != nil {
		dbStatus = "unreachable"
		status = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	writeJSON(w, httpStatus, map[string]string{
		"status": status,
		"db":     dbStatus,
	})
}
