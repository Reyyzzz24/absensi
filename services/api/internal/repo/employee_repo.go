package repo

import (
	"context"

	"gorm.io/gorm"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
)

type EmployeeRepo struct {
	db *gorm.DB
}

func NewEmployeeRepo(db *gorm.DB) EmployeeRepo {
	return EmployeeRepo{db: db}
}

func (r EmployeeRepo) FindByNIK(ctx context.Context, nik string) (*domain.Employee, error) {
	var e domain.Employee
	if err := r.db.WithContext(ctx).Where("nik = ? AND is_active", nik).First(&e).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (r EmployeeRepo) FindByID(ctx context.Context, id int64) (*domain.Employee, error) {
	var e domain.Employee
	if err := r.db.WithContext(ctx).Preload("Department").First(&e, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (r EmployeeRepo) List(ctx context.Context) ([]domain.Employee, error) {
	var out []domain.Employee
	err := r.db.WithContext(ctx).Preload("Department").Order("full_name").Find(&out).Error
	return out, err
}

func (r EmployeeRepo) Create(ctx context.Context, e *domain.Employee) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r EmployeeRepo) Update(ctx context.Context, e *domain.Employee) error {
	return r.db.WithContext(ctx).Save(e).Error
}
