package repo

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
)

type NationalHolidayRepo struct{ db *gorm.DB }

func NewNationalHolidayRepo(db *gorm.DB) NationalHolidayRepo { return NationalHolidayRepo{db: db} }

func (r NationalHolidayRepo) FindByDate(ctx context.Context, date time.Time) (*domain.NationalHoliday, error) {
	var h domain.NationalHoliday
	err := r.db.WithContext(ctx).Where("holiday_date = ?", date).First(&h).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &h, nil
}

// ListByYear returns every national holiday cached for a given year,
// ordered by date -- used by both the sync diffing and the admin calendar
// UI.
func (r NationalHolidayRepo) ListByYear(ctx context.Context, year int) ([]domain.NationalHoliday, error) {
	var out []domain.NationalHoliday
	err := r.db.WithContext(ctx).Where("year = ?", year).Order("holiday_date").Find(&out).Error
	return out, err
}

// ListByDateRange returns cached national holidays overlapping [start, end]
// -- used by the combined calendar endpoint and the recap loop.
func (r NationalHolidayRepo) ListByDateRange(ctx context.Context, start, end time.Time) ([]domain.NationalHoliday, error) {
	var out []domain.NationalHoliday
	err := r.db.WithContext(ctx).
		Where("holiday_date BETWEEN ? AND ?", start, end).
		Order("holiday_date").
		Find(&out).Error
	return out, err
}

// UpsertSynced inserts or updates a row coming from the sync job. Rows an
// admin has since hand-edited (source='manual') are left untouched -- the
// sync must never silently overwrite a manual correction/override.
func (r NationalHolidayRepo) UpsertSynced(ctx context.Context, h *domain.NationalHoliday) error {
	h.Source = domain.HolidaySourceSync
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "holiday_date"}},
			Where:     clause.Where{Exprs: []clause.Expression{clause.Eq{Column: "national_holidays.source", Value: "sync"}}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "year", "is_cuti_bersama", "updated_at"}),
		}).
		Create(h).Error
}

func (r NationalHolidayRepo) Update(ctx context.Context, h *domain.NationalHoliday) error {
	h.Source = domain.HolidaySourceManual
	return r.db.WithContext(ctx).Save(h).Error
}

func (r NationalHolidayRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&domain.NationalHoliday{}, id).Error
}

type CompanyHolidayRepo struct{ db *gorm.DB }

func NewCompanyHolidayRepo(db *gorm.DB) CompanyHolidayRepo { return CompanyHolidayRepo{db: db} }

func (r CompanyHolidayRepo) List(ctx context.Context) ([]domain.CompanyHoliday, error) {
	var out []domain.CompanyHoliday
	err := r.db.WithContext(ctx).Order("start_date").Find(&out).Error
	return out, err
}

// ListOverlapping returns manual company holidays whose [start,end] range
// overlaps [rangeStart, rangeEnd].
func (r CompanyHolidayRepo) ListOverlapping(ctx context.Context, rangeStart, rangeEnd time.Time) ([]domain.CompanyHoliday, error) {
	var out []domain.CompanyHoliday
	err := r.db.WithContext(ctx).
		Where("start_date <= ? AND end_date >= ?", rangeEnd, rangeStart).
		Order("start_date").
		Find(&out).Error
	return out, err
}

func (r CompanyHolidayRepo) FindByID(ctx context.Context, id int64) (*domain.CompanyHoliday, error) {
	var h domain.CompanyHoliday
	err := r.db.WithContext(ctx).First(&h, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r CompanyHolidayRepo) Create(ctx context.Context, h *domain.CompanyHoliday) error {
	return r.db.WithContext(ctx).Create(h).Error
}

func (r CompanyHolidayRepo) Update(ctx context.Context, h *domain.CompanyHoliday) error {
	return r.db.WithContext(ctx).Save(h).Error
}

func (r CompanyHolidayRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&domain.CompanyHoliday{}, id).Error
}
