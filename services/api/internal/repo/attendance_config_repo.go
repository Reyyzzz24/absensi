package repo

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
)

type OfficeLocationRepo struct{ db *gorm.DB }

func NewOfficeLocationRepo(db *gorm.DB) OfficeLocationRepo { return OfficeLocationRepo{db: db} }

// ActiveLocations returns all active office locations. A check-in only needs
// to be within radius of at least one of them -- D-1 (legacy hardcoded a
// single office coordinate and never enforced radius at all).
func (r OfficeLocationRepo) ActiveLocations(ctx context.Context) ([]domain.OfficeLocation, error) {
	var locs []domain.OfficeLocation
	err := r.db.WithContext(ctx).Where("is_active", true).Find(&locs).Error
	return locs, err
}

func (r OfficeLocationRepo) List(ctx context.Context) ([]domain.OfficeLocation, error) {
	var locs []domain.OfficeLocation
	err := r.db.WithContext(ctx).Order("id").Find(&locs).Error
	return locs, err
}

func (r OfficeLocationRepo) Create(ctx context.Context, loc *domain.OfficeLocation) error {
	return r.db.WithContext(ctx).Create(loc).Error
}

func (r OfficeLocationRepo) FindByID(ctx context.Context, id int64) (*domain.OfficeLocation, error) {
	var loc domain.OfficeLocation
	if err := r.db.WithContext(ctx).First(&loc, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &loc, nil
}

func (r OfficeLocationRepo) Update(ctx context.Context, loc *domain.OfficeLocation) error {
	return r.db.WithContext(ctx).Model(&domain.OfficeLocation{}).Where("id = ?", loc.ID).Updates(map[string]any{
		"name":          loc.Name,
		"latitude":      loc.Latitude,
		"longitude":     loc.Longitude,
		"radius_meters": loc.RadiusMeters,
		"is_active":     loc.IsActive,
	}).Error
}

type FieldAssignmentRepo struct{ db *gorm.DB }

func NewFieldAssignmentRepo(db *gorm.DB) FieldAssignmentRepo { return FieldAssignmentRepo{db: db} }

// Exists reports whether the employee has an admin-approved "dinas luar"
// assignment for workDate -- D-21. When true, the check-in geofence check
// is bypassed for that day.
func (r FieldAssignmentRepo) Exists(ctx context.Context, employeeID int64, workDate time.Time) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.FieldAssignment{}).
		Where("employee_id = ? AND work_date = ?", employeeID, workDate).
		Count(&count).Error
	return count > 0, err
}

func (r FieldAssignmentRepo) Create(ctx context.Context, fa *domain.FieldAssignment) error {
	return r.db.WithContext(ctx).Create(fa).Error
}

func (r FieldAssignmentRepo) ListByEmployee(ctx context.Context, employeeID int64) ([]domain.FieldAssignment, error) {
	var out []domain.FieldAssignment
	q := r.db.WithContext(ctx).Order("work_date DESC")
	if employeeID != 0 {
		q = q.Where("employee_id = ?", employeeID)
	}
	err := q.Find(&out).Error
	return out, err
}

type ShiftScheduleRepo struct{ db *gorm.DB }

func NewShiftScheduleRepo(db *gorm.DB) ShiftScheduleRepo { return ShiftScheduleRepo{db: db} }

// ResolveShift determines an employee's shift for a given date: an explicit
// per-date work_schedules entry always wins over the recurring
// weekly_shift_defaults entry for that weekday -- D-18. Returns nil, nil if
// no shift is assigned either way.
func (r ShiftScheduleRepo) ResolveShift(ctx context.Context, employeeID int64, date time.Time) (*domain.Shift, error) {
	var ws domain.WorkSchedule
	err := r.db.WithContext(ctx).Preload("Shift").
		Where("employee_id = ? AND work_date = ?", employeeID, date).
		First(&ws).Error
	if err == nil {
		return &ws.Shift, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	var wd domain.WeeklyShiftDefault
	err = r.db.WithContext(ctx).Preload("Shift").
		Where("employee_id = ? AND day_of_week = ?", employeeID, int(date.Weekday())).
		First(&wd).Error
	if err == nil {
		return &wd.Shift, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return nil, nil
}
