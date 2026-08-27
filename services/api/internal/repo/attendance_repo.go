package repo

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
)

// AttendanceRepo backs the admin monitoring screens -- read-only, cross-
// employee visibility by design (legacy tampilkanpeta/tampilkanpeta2,
// LOGIC_SPEC.md §7), unlike every other employee-facing repo which scopes
// to the calling employee.
type AttendanceRepo struct{ db *gorm.DB }

func NewAttendanceRepo(db *gorm.DB) AttendanceRepo { return AttendanceRepo{db: db} }

func (r AttendanceRepo) ListByDate(ctx context.Context, date time.Time) ([]domain.Attendance, error) {
	var out []domain.Attendance
	err := r.db.WithContext(ctx).
		Preload("Employee").
		Where("work_date = ?", date).
		Order("check_in_at").
		Find(&out).Error
	return out, err
}

// ListByEmployeeAndDateRange backs the monthly recap (D-15-aware "hadir"
// detection) -- deliberately a plain range query computed day-by-day in Go
// (see internal/usecase/recap), not the legacy's hardcoded 31-column SQL
// pivot (cetakrekap, LOGIC_SPEC.md §12) which silently mishandled short
// months and was unmaintainable.
func (r AttendanceRepo) ListByEmployeeAndDateRange(ctx context.Context, employeeID int64, start, end time.Time) ([]domain.Attendance, error) {
	var out []domain.Attendance
	err := r.db.WithContext(ctx).
		Where("employee_id = ? AND work_date BETWEEN ? AND ?", employeeID, start, end).
		Find(&out).Error
	return out, err
}

func (r AttendanceRepo) FindByID(ctx context.Context, id int64) (*domain.Attendance, error) {
	var a domain.Attendance
	if err := r.db.WithContext(ctx).First(&a, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}
