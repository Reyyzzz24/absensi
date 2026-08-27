// Package testutil provides a real Postgres connection for integration
// tests. Rather than testcontainers (the dev container has no Docker socket
// -- see docs/PROGRESS.md Phase 4 notes), tests run against a dedicated
// "absensi_test" database on the same Postgres instance the dev stack
// already uses (DATABASE_URL env var, set by docker-compose.yml).
//
// IMPORTANT: every package using NewDB shares that one "absensi_test"
// database, and NewDB drops+recreates it on every call. `go test ./...`
// runs different packages' test binaries in parallel by default, which
// races multiple packages' DROP/CREATE DATABASE against each other and
// produces spurious failures that have nothing to do with the code under
// test. Always run with `go test -p 1 ./...` (or scope to one package at a
// time) when integration tests are involved.
package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gorm.io/gorm"

	"github.com/eprisi/absensi-next/services/api/internal/platform/db"
)

// baseDSN returns the dev DATABASE_URL with the database name swapped to
// "postgres" (always present) so we can issue CREATE/DROP DATABASE.
func baseDSN(t *testing.T) (adminDSN, testDSN string) {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set -- skipping integration test (run inside the api container)")
	}
	idx := strings.LastIndex(dsn, "/")
	if idx == -1 {
		t.Fatalf("unexpected DATABASE_URL shape: %s", dsn)
	}
	prefix := dsn[:idx]
	suffix := ""
	if q := strings.Index(dsn[idx:], "?"); q != -1 {
		suffix = dsn[idx:][q:]
	}
	return prefix + "/postgres" + suffix, prefix + "/absensi_test" + suffix
}

// NewDB creates a fresh "absensi_test" database, runs all migrations against
// it, and returns a connected *gorm.DB. Each call drops and recreates the
// database first, so tests get a clean schema but must not run in parallel
// against the same Postgres instance (they share one test DB, sequentially).
func NewDB(t *testing.T) *gorm.DB {
	t.Helper()
	adminDSN, testDSN := baseDSN(t)

	admin, err := db.Connect(adminDSN)
	if err != nil {
		t.Fatalf("connect to admin db: %v", err)
	}
	sqlDB, err := admin.DB()
	if err != nil {
		t.Fatalf("get sql.DB: %v", err)
	}
	defer sqlDB.Close()

	if _, err := sqlDB.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'absensi_test' AND pid <> pg_backend_pid()`); err != nil {
		t.Fatalf("terminate existing connections: %v", err)
	}
	if _, err := sqlDB.Exec(`DROP DATABASE IF EXISTS absensi_test`); err != nil {
		t.Fatalf("drop test db: %v", err)
	}
	if _, err := sqlDB.Exec(`CREATE DATABASE absensi_test`); err != nil {
		t.Fatalf("create test db: %v", err)
	}

	migrationsPath := findMigrationsPath(t)
	if err := db.Migrate(testDSN, migrationsPath); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	gdb, err := db.Connect(testDSN)
	if err != nil {
		t.Fatalf("connect to test db: %v", err)
	}
	return gdb
}

// findMigrationsPath locates services/api/migrations relative to the
// package under test, walking up from the working directory (which is the
// test's own package dir, several levels below services/api).
func findMigrationsPath(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for i := 0; i < 10; i++ {
		candidate := filepath.Join(dir, "migrations")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		dir = filepath.Dir(dir)
	}
	t.Fatal("could not locate migrations directory")
	return ""
}
