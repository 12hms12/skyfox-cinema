package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"skyfox/bookings/model"
)

type ProfilePictureRepository interface {
	Upsert(ctx context.Context, pic model.ProfilePicture) error
	GetByUserID(ctx context.Context, userID uint) (*model.ProfilePicture, error)
	DeleteByUserID(ctx context.Context, userID uint) error
}

type profilePictureRepository struct {
	db *gorm.DB
}

func NewProfilePictureRepository(db gormProvider) ProfilePictureRepository {
	return &profilePictureRepository{db: db.GormDB()}
}

func NewProfilePictureRepositoryFromGorm(gdb *gorm.DB) ProfilePictureRepository {
	return &profilePictureRepository{db: gdb}
}

func (r *profilePictureRepository) Upsert(ctx context.Context, pic model.ProfilePicture) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"image_data", "content_type", "updated_at"}),
		}).
		Create(&pic).Error
}

func (r *profilePictureRepository) GetByUserID(ctx context.Context, userID uint) (*model.ProfilePicture, error) {
	var pic model.ProfilePicture
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&pic).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &pic, nil
}

func (r *profilePictureRepository) DeleteByUserID(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&model.ProfilePicture{}).Error
}
