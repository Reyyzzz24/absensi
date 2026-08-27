// cmd/seed populates a fresh dev database with just enough data to exercise
// the API end to end (admin login, shift/office config, employee check-in).
// Dev-only: never run this against a real/production database. Idempotent
// (safe to re-run) via ON CONFLICT DO NOTHING on each unique key.
//
// Coordinates/credentials here are placeholders for local testing only --
// NOT the real office location from docs/DECISIONS.md D-1 (that still needs
// the project owner's actual coordinates, tracked separately).
package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/eprisi/absensi-next/services/api/internal/config"
	"github.com/eprisi/absensi-next/services/api/internal/platform/db"
)

const devPassword = "password123"

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	gdb, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(devPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}

	if err := seed(gdb, string(hash)); err != nil {
		log.Fatalf("seed: %v", err)
	}

	fmt.Println("Seed complete. Dev credentials (password for both):", devPassword)
	fmt.Println("  Admin login:    POST /api/v1/auth/admin/login    { \"username\": \"admin\", \"password\": \"" + devPassword + "\" }")
	fmt.Println("  Employee login: POST /api/v1/auth/employee/login { \"nik\": \"0001\", \"password\": \"" + devPassword + "\" }")
	fmt.Println("  Seeded office location: -6.200000, 106.816666 (placeholder Jakarta coords, radius 100m)")
	fmt.Println("  Seeded shift: code SH01, 09:00-17:00, 15 min grace")
}

func seed(gdb *gorm.DB, passwordHash string) error {
	return gdb.Transaction(func(tx *gorm.DB) error {
		steps := []struct {
			name string
			sql  string
			args []any
		}{
			{"roles", `INSERT INTO roles (name) VALUES ('superadmin'), ('admin') ON CONFLICT (name) DO NOTHING`, nil},
			{"departments", `INSERT INTO departments (code, name) VALUES ('GA', 'General Affairs') ON CONFLICT (code) DO NOTHING`, nil},
			{
				"admin user",
				`INSERT INTO users (name, username, password_hash, role_id)
				 SELECT 'Dev Admin', 'admin', ?, id FROM roles WHERE name = 'superadmin'
				 ON CONFLICT (username) DO NOTHING`,
				[]any{passwordHash},
			},
			{
				"employee",
				`INSERT INTO employees (nik, full_name, password_hash, department_id, is_active)
				 SELECT '0001', 'Dev Employee', ?, id, true FROM departments WHERE code = 'GA'
				 ON CONFLICT (nik) DO NOTHING`,
				[]any{passwordHash},
			},
			{
				"office location",
				`INSERT INTO office_locations (name, latitude, longitude, radius_meters, is_active)
				 SELECT 'Kantor Pusat (dev placeholder)', -6.200000, 106.816666, 100, true
				 WHERE NOT EXISTS (SELECT 1 FROM office_locations WHERE name = 'Kantor Pusat (dev placeholder)')`,
				nil,
			},
			{
				"shift",
				`INSERT INTO shifts (code, name, is_day_off, start_time, end_time, late_grace_minutes)
				 VALUES ('SH01', 'Reguler 09-17', false, '09:00:00', '17:00:00', 15)
				 ON CONFLICT (code) DO NOTHING`,
				nil,
			},
			{
				"weekly shift default (Mon-Fri for the seeded employee)",
				`INSERT INTO weekly_shift_defaults (employee_id, day_of_week, shift_id)
				 SELECT e.id, d.dow, s.id
				 FROM employees e, shifts s, generate_series(1, 5) AS d(dow)
				 WHERE e.nik = '0001' AND s.code = 'SH01'
				 ON CONFLICT (employee_id, day_of_week) DO NOTHING`,
				nil,
			},
		}

		for _, step := range steps {
			if err := tx.Exec(step.sql, step.args...).Error; err != nil {
				return fmt.Errorf("%s: %w", step.name, err)
			}
		}
		return nil
	})
}
