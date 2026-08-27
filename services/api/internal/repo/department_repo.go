package repo

import (
	"context"

	"gorm.io/gorm"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
)

type DepartmentRepo struct{ db *gorm.DB }

func NewDepartmentRepo(db *gorm.DB) DepartmentRepo { return DepartmentRepo{db: db} }

func (r DepartmentRepo) List(ctx context.Context) ([]domain.Department, error) {
	var out []domain.Department
	err := r.db.WithContext(ctx).Order("name").Find(&out).Error
	return out, err
}

func (r DepartmentRepo) Create(ctx context.Context, d *domain.Department) error {
	return r.db.WithContext(ctx).Create(d).Error
}
