package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"skyfox/bookings/constants"
	"skyfox/bookings/model"
	ae "skyfox/error"
)

const (
	testUploadUserID    = uint(1)
	testContentTypeJPEG = "image/jpeg"
	testContentTypePNG  = "image/png"
	testBase64Data      = "aGVsbG8="
	testDataURLJPEG     = "data:image/jpeg;base64,aGVsbG8="
)

type MockProfilePictureRepository struct {
	mock.Mock
}

func (m *MockProfilePictureRepository) Upsert(ctx context.Context, pic model.ProfilePicture) error {
	args := m.Called(ctx, pic)
	return args.Error(0)
}

func (m *MockProfilePictureRepository) GetByUserID(ctx context.Context, userID uint) (*model.ProfilePicture, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProfilePicture), args.Error(1)
}

func (m *MockProfilePictureRepository) DeleteByUserID(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestUploadProfileImage(t *testing.T) {
	ctx := context.Background()

	smallImage := []byte("hello")
	oversizedImage := make([]byte, constants.MaxProfileImageSizeBytes+1)

	tests := []struct {
		name        string
		userID      uint
		imageData   []byte
		contentType string
		setupMock   func(avatarRepo *MockAvatarRepository, pfpRepo *MockProfilePictureRepository)
		wantErrCode string
		wantType    string
	}{
		{
			name:        "should upload profile image and store base64 in db when image is within size limit",
			userID:      testUploadUserID,
			imageData:   smallImage,
			contentType: testContentTypeJPEG,
			setupMock: func(avatarRepo *MockAvatarRepository, pfpRepo *MockProfilePictureRepository) {
				pfpRepo.On("Upsert", ctx, mock.MatchedBy(func(p model.ProfilePicture) bool {
					return p.UserID == testUploadUserID && p.ContentType == testContentTypeJPEG && p.ImageData != ""
				})).Return(nil)
				avatarRepo.On("UpdateUserAvatar", ctx, testUploadUserID, mock.MatchedBy(func(url string) bool {
					return strings.HasPrefix(url, "data:image/jpeg;base64,")
				}), Uploaded).Return(nil)
			},
			wantType: Uploaded,
		},
		{
			name:        "should return ErrImageTooLarge when image size exceeds the 5MB limit",
			userID:      testUploadUserID,
			imageData:   oversizedImage,
			contentType: testContentTypeJPEG,
			setupMock:   func(avatarRepo *MockAvatarRepository, pfpRepo *MockProfilePictureRepository) {},
			wantErrCode: "ERR_IMAGE_TOO_LARGE",
		},
		{
			name:        "should return error when profile picture repository upsert fails",
			userID:      testUploadUserID,
			imageData:   smallImage,
			contentType: testContentTypePNG,
			setupMock: func(avatarRepo *MockAvatarRepository, pfpRepo *MockProfilePictureRepository) {
				pfpRepo.On("Upsert", ctx, mock.Anything).Return(errors.New("db error"))
			},
			wantErrCode: "",
		},
		{
			name:        "should return error when avatar repository update fails after successful pfp upsert",
			userID:      testUploadUserID,
			imageData:   smallImage,
			contentType: testContentTypeJPEG,
			setupMock: func(avatarRepo *MockAvatarRepository, pfpRepo *MockProfilePictureRepository) {
				pfpRepo.On("Upsert", ctx, mock.Anything).Return(nil)
				avatarRepo.On("UpdateUserAvatar", ctx, testUploadUserID, mock.Anything, Uploaded).Return(errors.New("avatar update error"))
			},
			wantErrCode: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			avatarRepo := new(MockAvatarRepository)
			pfpRepo := new(MockProfilePictureRepository)
			tt.setupMock(avatarRepo, pfpRepo)

			svc := NewProfileImageService(avatarRepo, pfpRepo)
			resp, err := svc.UploadProfileImage(ctx, tt.userID, tt.imageData, tt.contentType)

			if tt.wantErrCode != "" {
				var appErr *ae.AppError
				assert.True(t, errors.As(err, &appErr))
				assert.Equal(t, tt.wantErrCode, appErr.Code)
				assert.Nil(t, resp)
			} else if tt.wantType != "" {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.wantType, resp.AvatarType)
				assert.True(t, strings.HasPrefix(resp.AvatarURL, "data:"+tt.contentType+";base64,"))
			} else {
				assert.Error(t, err)
				assert.Nil(t, resp)
			}

			avatarRepo.AssertExpectations(t)
			pfpRepo.AssertExpectations(t)
		})
	}
}

func TestGetProfileImage(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		userID      uint
		setupMock   func(avatarRepo *MockAvatarRepository, pfpRepo *MockProfilePictureRepository)
		wantType    string
		wantURLPart string
	}{
		{
			name:   "should return data URL from profile_pictures table when an uploaded image exists for the user",
			userID: testUploadUserID,
			setupMock: func(avatarRepo *MockAvatarRepository, pfpRepo *MockProfilePictureRepository) {
				pfpRepo.On("GetByUserID", ctx, testUploadUserID).Return(
					&model.ProfilePicture{UserID: testUploadUserID, ImageData: testBase64Data, ContentType: testContentTypeJPEG},
					nil,
				)
			},
			wantType:    Uploaded,
			wantURLPart: "data:image/jpeg;base64,",
		},
		{
			name:   "should fall back to avatar repository when no uploaded image exists in profile_pictures",
			userID: testUploadUserID,
			setupMock: func(avatarRepo *MockAvatarRepository, pfpRepo *MockProfilePictureRepository) {
				pfpRepo.On("GetByUserID", ctx, testUploadUserID).Return(nil, nil)
				avatarRepo.On("GetUserAvatar", ctx, testUploadUserID).Return("https://predefined.com/av.jpg", Predefined, nil)
			},
			wantType:    Predefined,
			wantURLPart: "https://predefined.com/av.jpg",
		},
		{
			name:   "should return error when profile_pictures repository lookup fails",
			userID: testUploadUserID,
			setupMock: func(avatarRepo *MockAvatarRepository, pfpRepo *MockProfilePictureRepository) {
				pfpRepo.On("GetByUserID", ctx, testUploadUserID).Return(nil, errors.New("db error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			avatarRepo := new(MockAvatarRepository)
			pfpRepo := new(MockProfilePictureRepository)
			tt.setupMock(avatarRepo, pfpRepo)

			svc := NewProfileImageService(avatarRepo, pfpRepo)
			resp, err := svc.GetProfileImage(ctx, tt.userID)

			if tt.wantType != "" {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.wantType, resp.AvatarType)
				assert.Contains(t, resp.AvatarURL, tt.wantURLPart)
			} else {
				assert.Error(t, err)
				assert.Nil(t, resp)
			}

			avatarRepo.AssertExpectations(t)
			pfpRepo.AssertExpectations(t)
		})
	}
}
