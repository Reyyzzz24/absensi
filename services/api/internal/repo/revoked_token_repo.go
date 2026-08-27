package repo

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
)

type RevokedTokenRepo struct {
	db *gorm.DB
}

func NewRevokedTokenRepo(db *gorm.DB) RevokedTokenRepo {
	return RevokedTokenRepo{db: db}
}

// Revoke is idempotent -- logging out twice with the same token must not error.
func (r RevokedTokenRepo) Revoke(ctx context.Context, jti string, expiresAt time.Time) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&domain.RevokedRefreshToken{
		JTI:       jti,
		ExpiresAt: expiresAt,
	}).Error
}

func (r RevokedTokenRepo) IsRevoked(ctx context.Context, jti string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.RevokedRefreshToken{}).Where("jti = ?", jti).Count(&count).Error
	return count > 0, err
}
