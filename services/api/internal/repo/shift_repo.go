package repo

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
)

type ShiftRepo struct{ db *gorm.DB }

func NewShiftRepo(db *gorm.DB) ShiftRepo { return ShiftRepo{db: db} }

func (r ShiftRepo) List(ctx context.Context) ([]domain.Shift, error) {
	var out []domain.Shift
	err := r.db.WithContext(ctx).Order("id").Find(&out).Error
	return out, err
}

func (r ShiftRepo) FindByID(ctx context.Context, id int64) (*domain.Shift, error) {
	var s domain.Shift
	if err := r.db.WithContext(ctx).First(&s, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r ShiftRepo) Create(ctx context.Context, s *domain.Shift) error {
	return r.db.WithContext(ctx).Create(s).Error
}

// Update only touches the mutable fields -- Code and IsOvernight (generated)
// are never rewritten by an update.
func (r ShiftRepo) Update(ctx context.Context, s *domain.Shift) error {
	return r.db.WithContext(ctx).Model(&domain.Shift{}).Where("id = ?", s.ID).Updates(map[string]any{
		"name":               s.Name,
		"is_day_off":         s.IsDayOff,
		"start_time":         s.StartTime,
		"end_time":           s.EndTime,
		"late_grace_minutes": s.LateGraceMinutes,
	}).Error
}

type WeeklyShiftDefaultRepo struct{ db *gorm.DB }

func NewWeeklyShiftDefaultRepo(db *gorm.DB) WeeklyShiftDefaultRepo {
	return WeeklyShiftDefaultRepo{db: db}
}

// Upsert sets (or replaces) an employee's recurring shift for a weekday --
// legacy konfigurasi_jamkerja.
func (r WeeklyShiftDefaultRepo) Upsert(ctx context.Context, employeeID int64, dayOfWeek int16, shiftID int64) error {
	return r.db.WithContext(ctx).
		Where("employee_id = ? AND day_of_week = ?", employeeID, dayOfWeek).
		Assign("shift_id", shiftID).
		FirstOrCreate(&domain.WeeklyShiftDefault{EmployeeID: employeeID, DayOfWeek: dayOfWeek, ShiftID: shiftID}).Error
}

type WorkScheduleRepo struct{ db *gorm.DB }

func NewWorkScheduleRepo(db *gorm.DB) WorkScheduleRepo { return WorkScheduleRepo{db: db} }

// Upsert sets (or replaces) an employee's shift for one specific date --
// legacy jadwal_kerja. Always wins over WeeklyShiftDefault for that date, D-18.
func (r WorkScheduleRepo) Upsert(ctx context.Context, employeeID int64, workDate time.Time, shiftID int64) error {
	return r.db.WithContext(ctx).
		Where("employee_id = ? AND work_date = ?", employeeID, workDate).
		Assign("shift_id", shiftID).
		FirstOrCreate(&domain.WorkSchedule{EmployeeID: employeeID, WorkDate: workDate, ShiftID: shiftID}).Error
}
