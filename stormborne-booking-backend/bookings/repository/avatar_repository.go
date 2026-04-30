package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"skyfox/bookings/model"
)

type gormProvider interface {
	GormDB() *gorm.DB
}

type avatarRepository struct {
	db *gorm.DB
}

func NewAvatarRepository(db gormProvider) *avatarRepository {
	return &avatarRepository{db: db.GormDB()}
}

func NewAvatarRepositoryFromGorm(gdb *gorm.DB) *avatarRepository {
	return &avatarRepository{db: gdb}
}

func (r *avatarRepository) ListAll(ctx context.Context) ([]model.PredefinedAvatar, error) {
	var items []model.PredefinedAvatar
	if err := r.db.WithContext(ctx).
		Order("id ASC").
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *avatarRepository) ListByGender(ctx context.Context, gender string) ([]model.PredefinedAvatar, error) {
	var items []model.PredefinedAvatar
	if err := r.db.WithContext(ctx).
		Where("gender = ?", gender).
		Order("id ASC").
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *avatarRepository) GetByID(ctx context.Context, id int64) (*model.PredefinedAvatar, error) {
	var av model.PredefinedAvatar
	if err := r.db.WithContext(ctx).First(&av, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &av, nil
}

func (r *avatarRepository) ExistsByURL(ctx context.Context, url string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.PredefinedAvatar{}).
		Where("url = ?", url).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *avatarRepository) BulkInsertIfNotExists(ctx context.Context, avatars []model.PredefinedAvatar) error {
	if len(avatars) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		const chunk = 100
		for i := 0; i < len(avatars); i += chunk {
			end := i + chunk
			if end > len(avatars) {
				end = len(avatars)
			}
			part := avatars[i:end]
			if err := tx.Clauses(
				clause.OnConflict{
					Columns:   []clause.Column{{Name: "url"}},
					DoNothing: true,
				},
			).Create(&part).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *avatarRepository) GetRandomByGender(ctx context.Context, gender string) (*model.PredefinedAvatar, error) {
	var av model.PredefinedAvatar
	q := r.db.WithContext(ctx).Model(&model.PredefinedAvatar{})
	if gender != "" {
		q = q.Where("gender = ?", gender)
	}
	if err := q.Order("RANDOM()").First(&av).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &av, nil
}

func (r *avatarRepository) GetUserAvatar(ctx context.Context, userID uint) (string, string, error) {
	var user model.Customer
	if err := r.db.WithContext(ctx).Select("avatar_url", "avatar_type").First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", nil
		}
		return "", "", err
	}
	return user.AvatarURL, string(user.AvatarType), nil
}

func (r *avatarRepository) UpdateUserAvatar(ctx context.Context, userID uint, avatarURL string, avatarType string) error {
	return r.db.WithContext(ctx).
		Model(&model.Customer{ID: userID}).
		Updates(map[string]any{
			"avatar_url":  avatarURL,
			"avatar_type": avatarType,
		}).Error
}
