package repo

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
)

type NotificationPreferenceRepo struct{ db *gorm.DB }

func NewNotificationPreferenceRepo(db *gorm.DB) NotificationPreferenceRepo {
	return NotificationPreferenceRepo{db: db}
}

// List returns only the rows explicitly saved (opt-outs / explicit
// opt-ins). A type with no row is enabled by default -- the caller merges
// this against the known type list, it is not the full picture on its own.
func (r NotificationPreferenceRepo) List(ctx context.Context, audience string, id int64) ([]domain.NotificationPreference, error) {
	var out []domain.NotificationPreference
	err := r.db.WithContext(ctx).
		Where("recipient_audience = ? AND recipient_id = ?", audience, id).
		Find(&out).Error
	return out, err
}

// IsEnabled defaults to true (opt-out model) when no row exists for this type.
func (r NotificationPreferenceRepo) IsEnabled(ctx context.Context, audience string, id int64, notifType string) (bool, error) {
	var pref domain.NotificationPreference
	err := r.db.WithContext(ctx).
		Where("recipient_audience = ? AND recipient_id = ? AND type = ?", audience, id, notifType).
		First(&pref).Error
	if err == gorm.ErrRecordNotFound {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return pref.Enabled, nil
}

func (r NotificationPreferenceRepo) Set(ctx context.Context, audience string, id int64, notifType string, enabled bool) error {
	pref := domain.NotificationPreference{
		RecipientAudience: audience,
		RecipientID:       id,
		Type:              notifType,
		Enabled:           enabled,
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "recipient_audience"}, {Name: "recipient_id"}, {Name: "type"}},
			DoUpdates: clause.AssignmentColumns([]string{"enabled"}),
		}).
		Create(&pref).Error
}
