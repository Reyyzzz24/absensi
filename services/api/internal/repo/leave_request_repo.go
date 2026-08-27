package repo

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
)

type LeaveRequestRepo struct{ db *gorm.DB }

func NewLeaveRequestRepo(db *gorm.DB) LeaveRequestRepo { return LeaveRequestRepo{db: db} }

func (r LeaveRequestRepo) Create(ctx context.Context, lr *domain.LeaveRequest) error {
	return r.db.WithContext(ctx).Create(lr).Error
}

func (r LeaveRequestRepo) ListByEmployee(ctx context.Context, employeeID int64) ([]domain.LeaveRequest, error) {
	var out []domain.LeaveRequest
	err := r.db.WithContext(ctx).Where("employee_id = ?", employeeID).Order("start_date DESC").Find(&out).Error
	return out, err
}

// ListByStatus is used by the admin review screen -- Preload("Employee") so
// the UI can show the employee's name/NIK instead of a bare ID.
func (r LeaveRequestRepo) ListByStatus(ctx context.Context, status domain.LeaveStatus) ([]domain.LeaveRequest, error) {
	var out []domain.LeaveRequest
	q := r.db.WithContext(ctx).Preload("Employee").Order("start_date DESC")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Find(&out).Error
	return out, err
}

// ListApprovedOverlapping backs the monthly recap -- approved leave days are
// excluded from "absent" in the recap grid, per D-15 (legacy never
// cross-referenced pengajuan_izin against attendance at all).
func (r LeaveRequestRepo) ListApprovedOverlapping(ctx context.Context, employeeID int64, start, end time.Time) ([]domain.LeaveRequest, error) {
	var out []domain.LeaveRequest
	err := r.db.WithContext(ctx).
		Where("employee_id = ? AND status = ? AND start_date <= ? AND end_date >= ?",
			employeeID, domain.LeaveStatusApproved, end, start).
		Find(&out).Error
	return out, err
}

func (r LeaveRequestRepo) FindByID(ctx context.Context, id int64) (*domain.LeaveRequest, error) {
	var lr domain.LeaveRequest
	if err := r.db.WithContext(ctx).First(&lr, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &lr, nil
}

func (r LeaveRequestRepo) Save(ctx context.Context, lr *domain.LeaveRequest) error {
	return r.db.WithContext(ctx).Save(lr).Error
}
