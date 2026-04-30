package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"skyfox/bookings/constants"
	"skyfox/bookings/dto/response"
	"skyfox/bookings/model"
	ae "skyfox/error"
)

type ProfileImageService interface {
	UploadProfileImage(ctx context.Context, userID uint, imageData []byte, contentType string) (*response.AvatarResponse, error)
	GetProfileImage(ctx context.Context, userID uint) (*response.AvatarResponse, error)
}

type ProfilePictureRepository interface {
	Upsert(ctx context.Context, pic model.ProfilePicture) error
	GetByUserID(ctx context.Context, userID uint) (*model.ProfilePicture, error)
	DeleteByUserID(ctx context.Context, userID uint) error
}

type profileImageService struct {
	avatarRepo AvatarRepository
	pfpRepo    ProfilePictureRepository
}

func NewProfileImageService(avatarRepo AvatarRepository, pfpRepo ProfilePictureRepository) ProfileImageService {
	return &profileImageService{avatarRepo: avatarRepo, pfpRepo: pfpRepo}
}

func (s *profileImageService) UploadProfileImage(ctx context.Context, userID uint, imageData []byte, contentType string) (*response.AvatarResponse, error) {
	if len(imageData) > constants.MaxProfileImageSizeBytes {
		return nil, ae.ErrImageTooLarge()
	}

	base64Data := base64.StdEncoding.EncodeToString(imageData)
	dataURL := fmt.Sprintf("data:%s;base64,%s", contentType, base64Data)

	pic := model.ProfilePicture{
		UserID:      userID,
		ImageData:   base64Data,
		ContentType: contentType,
	}
	if err := s.pfpRepo.Upsert(ctx, pic); err != nil {
		return nil, err
	}

	if err := s.avatarRepo.UpdateUserAvatar(ctx, userID, dataURL, Uploaded); err != nil {
		return nil, err
	}

	return &response.AvatarResponse{
		UserID:     userID,
		AvatarURL:  dataURL,
		AvatarType: Uploaded,
		UpdatedAt:  time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *profileImageService) GetProfileImage(ctx context.Context, userID uint) (*response.AvatarResponse, error) {
	pic, err := s.pfpRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if pic == nil {
		avatarURL, avatarType, err := s.avatarRepo.GetUserAvatar(ctx, userID)
		if err != nil {
			return nil, err
		}
		return &response.AvatarResponse{
			UserID:     userID,
			AvatarURL:  avatarURL,
			AvatarType: avatarType,
		}, nil
	}

	dataURL := fmt.Sprintf("data:%s;base64,%s", pic.ContentType, pic.ImageData)
	return &response.AvatarResponse{
		UserID:     userID,
		AvatarURL:  dataURL,
		AvatarType: Uploaded,
	}, nil
}
