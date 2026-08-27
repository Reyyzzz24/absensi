package repo

import (
	"context"

	"gorm.io/gorm"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
)

type CompanySettingsRepo struct{ db *gorm.DB }

func NewCompanySettingsRepo(db *gorm.DB) CompanySettingsRepo { return CompanySettingsRepo{db: db} }

// Get always reads the single row (id = 1) seeded by migration 000016.
func (r CompanySettingsRepo) Get(ctx context.Context) (*domain.CompanySettings, error) {
	var c domain.CompanySettings
	if err := r.db.WithContext(ctx).First(&c, 1).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r CompanySettingsRepo) Update(ctx context.Context, c *domain.CompanySettings) error {
	c.ID = 1
	return r.db.WithContext(ctx).Save(c).Error
}
