package repo

import (
	"context"

	"gorm.io/gorm"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
)

type TaskRepo struct{ db *gorm.DB }

func NewTaskRepo(db *gorm.DB) TaskRepo { return TaskRepo{db: db} }

func (r TaskRepo) Create(ctx context.Context, t *domain.Task) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r TaskRepo) ListByEmployee(ctx context.Context, employeeID int64) ([]domain.Task, error) {
	var out []domain.Task
	err := r.db.WithContext(ctx).Where("employee_id = ?", employeeID).Order("starts_at DESC").Find(&out).Error
	return out, err
}

// FindOwnedByID fetches a task only if it belongs to employeeID -- this
// ownership scoping is what the legacy edit_task/update_task endpoints were
// missing (IDOR, A5/D-5). Returns nil, nil if not found OR not owned by this
// employee, so callers can respond 404 either way without revealing which.
func (r TaskRepo) FindOwnedByID(ctx context.Context, id, employeeID int64) (*domain.Task, error) {
	var t domain.Task
	err := r.db.WithContext(ctx).Where("id = ? AND employee_id = ?", id, employeeID).First(&t).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r TaskRepo) Save(ctx context.Context, t *domain.Task) error {
	return r.db.WithContext(ctx).Save(t).Error
}
