package attendance

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
)

// Today returns the employee's current attendance state: a still-open
// overnight cycle from yesterday if one exists (D-14), otherwise the latest
// cycle for today (D-23), otherwise nil if nothing has been recorded yet.
func (s Service) Today(ctx context.Context, employeeID int64) (*domain.Attendance, error) {
	now := time.Now().In(jakarta)
	today := civilDate(now)
	yesterday := today.AddDate(0, 0, -1)

	var openOvernight domain.Attendance
	err := s.db.WithContext(ctx).
		Joins("JOIN shifts ON shifts.id = attendances.shift_id").
		Where("attendances.employee_id = ? AND attendances.work_date = ? AND attendances.status = ? AND shifts.is_overnight = true",
			employeeID, yesterday, domain.AttendanceStatusOpen).
		Order("attendances.cycle_number DESC").
		First(&openOvernight).Error
	if err == nil {
		return &openOvernight, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var latest domain.Attendance
	err = s.db.WithContext(ctx).
		Where("employee_id = ? AND work_date = ?", employeeID, today).
		Order("cycle_number DESC").
		First(&latest).Error
	if err == nil {
		return &latest, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return nil, err
}
