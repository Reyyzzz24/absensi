package repo

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
)

type NotificationRepo struct{ db *gorm.DB }

func NewNotificationRepo(db *gorm.DB) NotificationRepo { return NotificationRepo{db: db} }

func (r NotificationRepo) Create(ctx context.Context, n *domain.Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

// ListForRecipient always filters on BOTH audience and id -- an employee id
// and an admin user id come from separate sequences and can collide
// numerically, so audience alone or id alone is not enough to stay
// IDOR-safe (D-5 pattern applied to notifications).
func (r NotificationRepo) ListForRecipient(ctx context.Context, audience string, id int64, limit, offset int) ([]domain.Notification, error) {
	var out []domain.Notification
	err := r.db.WithContext(ctx).
		Where("recipient_audience = ? AND recipient_id = ?", audience, id).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&out).Error
	return out, err
}

func (r NotificationRepo) UnreadCount(ctx context.Context, audience string, id int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Notification{}).
		Where("recipient_audience = ? AND recipient_id = ? AND read_at IS NULL", audience, id).
		Count(&count).Error
	return count, err
}

// MarkRead scopes the UPDATE by recipient too, not just id -- so one
// recipient can never mark another recipient's notification read by
// guessing an id (IDOR guard, not just a convenience filter).
func (r NotificationRepo) MarkRead(ctx context.Context, audience string, id, notificationID int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.Notification{}).
		Where("id = ? AND recipient_audience = ? AND recipient_id = ?", notificationID, audience, id).
		Update("read_at", now).Error
}

func (r NotificationRepo) MarkAllRead(ctx context.Context, audience string, id int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.Notification{}).
		Where("recipient_audience = ? AND recipient_id = ? AND read_at IS NULL", audience, id).
		Update("read_at", now).Error
}
