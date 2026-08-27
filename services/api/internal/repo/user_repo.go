package repo

import (
	"context"

	"gorm.io/gorm"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) UserRepo {
	return UserRepo{db: db}
}

func (r UserRepo) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var u domain.User
	if err := r.db.WithContext(ctx).Preload("Role").Where("username = ?", username).First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r UserRepo) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	var u domain.User
	if err := r.db.WithContext(ctx).Preload("Role").First(&u, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r UserRepo) Update(ctx context.Context, u *domain.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}
