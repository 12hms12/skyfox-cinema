package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"skyfox/bookings/dto/request"
	"skyfox/bookings/model"
	ae "skyfox/error"
)

type MockAvatarRepository struct {
	mock.Mock
}

func (m *MockAvatarRepository) ListAll(ctx context.Context) ([]model.PredefinedAvatar, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.PredefinedAvatar), args.Error(1)
}

func (m *MockAvatarRepository) ListByGender(ctx context.Context, gender string) ([]model.PredefinedAvatar, error) {
	args := m.Called(ctx, gender)
	return args.Get(0).([]model.PredefinedAvatar), args.Error(1)
}

func (m *MockAvatarRepository) GetByID(ctx context.Context, id int64) (*model.PredefinedAvatar, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.PredefinedAvatar), args.Error(1)
}

func (m *MockAvatarRepository) ExistsByURL(ctx context.Context, url string) (bool, error) {
	args := m.Called(ctx, url)
	return args.Bool(0), args.Error(1)
}

func (m *MockAvatarRepository) BulkInsertIfNotExists(ctx context.Context, avatars []model.PredefinedAvatar) error {
	args := m.Called(ctx, avatars)
	return args.Error(0)
}

func (m *MockAvatarRepository) GetRandomByGender(ctx context.Context, gender string) (*model.PredefinedAvatar, error) {
	args := m.Called(ctx, gender)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.PredefinedAvatar), args.Error(1)
}

func (m *MockAvatarRepository) GetUserAvatar(ctx context.Context, userID uint) (string, string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockAvatarRepository) UpdateUserAvatar(ctx context.Context, userID uint, avatarURL string, avatarType string) error {
	args := m.Called(ctx, userID, avatarURL, avatarType)
	return args.Error(0)
}

type MockProfilePictureDeleter struct {
	mock.Mock
}

func (m *MockProfilePictureDeleter) DeleteByUserID(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestUpdateAvatar(t *testing.T) {
	ctx := context.Background()
	userID := uint(1)
	predefined := model.PredefinedAvatar{ID: 10, Gender: "male", URL: "https://skyfox.com/pre.jpg"}

	tests := []struct {
		name        string
		req         request.UpdateAvatarRequest
		setupMock   func(m *MockAvatarRepository)
		wantErrCode string
		wantURL     string
		wantType    string
	}{
		{
			name: "should successfully update user avatar when valid predefined avatar ID is provided",
			req:  request.UpdateAvatarRequest{PredefinedAvatarID: 10},
			setupMock: func(m *MockAvatarRepository) {
				m.On("GetByID", ctx, int64(10)).Return(&predefined, nil)
				m.On("UpdateUserAvatar", ctx, userID, predefined.URL, "predefined").Return(nil)
			},
			wantURL:  predefined.URL,
			wantType: "predefined",
		},
		{
			name: "should successfully update user avatar when valid custom URL with type uploaded is provided",
			req:  request.UpdateAvatarRequest{AvatarURL: "https://custom.com/my.jpg", AvatarType: "uploaded"},
			setupMock: func(m *MockAvatarRepository) {
				m.On("UpdateUserAvatar", ctx, userID, "https://custom.com/my.jpg", "uploaded").Return(nil)
			},
			wantURL:  "https://custom.com/my.jpg",
			wantType: "uploaded",
		},
		{
			name: "should successfully update user avatar when valid custom URL with type predefined is provided",
			req:  request.UpdateAvatarRequest{AvatarURL: "https://cdn.com/pre.jpg", AvatarType: "predefined"},
			setupMock: func(m *MockAvatarRepository) {
				m.On("UpdateUserAvatar", ctx, userID, "https://cdn.com/pre.jpg", "predefined").Return(nil)
			},
			wantURL:  "https://cdn.com/pre.jpg",
			wantType: "predefined",
		},
		{
			name:        "should return ErrPickEitherIDOrURL when both predefined avatar ID and custom URL are provided",
			req:         request.UpdateAvatarRequest{PredefinedAvatarID: 10, AvatarURL: "https://x.com"},
			setupMock:   func(m *MockAvatarRepository) {},
			wantErrCode: "ERR_PICK_EITHER_ID_OR_URL",
		},
		{
			name:        "should return ErrPickEitherIDOrURL when neither predefined avatar ID nor custom URL is provided",
			req:         request.UpdateAvatarRequest{},
			setupMock:   func(m *MockAvatarRepository) {},
			wantErrCode: "ERR_PICK_EITHER_ID_OR_URL",
		},
		{
			name: "should return ErrAvatarNotFound when provided predefined avatar ID does not exist in database",
			req:  request.UpdateAvatarRequest{PredefinedAvatarID: 999},
			setupMock: func(m *MockAvatarRepository) {
				m.On("GetByID", ctx, int64(999)).Return(nil, nil)
			},
			wantErrCode: "ERR_AVATAR_NOT_FOUND",
		},
		{
			name:        "should return ErrInvalidAvatarURL when custom URL format is invalid or malformed",
			req:         request.UpdateAvatarRequest{AvatarURL: "not-a-url", AvatarType: "uploaded"},
			setupMock:   func(m *MockAvatarRepository) {},
			wantErrCode: "ERR_INVALID_AVATAR_URL",
		},
		{
			name:        "should return ErrInvalidAvatarType when avatar type is neither predefined nor uploaded",
			req:         request.UpdateAvatarRequest{AvatarURL: "https://ok.com/i.png", AvatarType: "invalid"},
			setupMock:   func(m *MockAvatarRepository) {},
			wantErrCode: "ERR_INVALID_AVATAR_TYPE",
		},
		{
			name:        "should return ErrInvalidAvatarType when avatar type is empty for custom URL",
			req:         request.UpdateAvatarRequest{AvatarURL: "https://ok.com/i.png", AvatarType: ""},
			setupMock:   func(m *MockAvatarRepository) {},
			wantErrCode: "ERR_INVALID_AVATAR_TYPE",
		},
		{
			name:        "should return ErrPickEitherIDOrURL when custom URL contains only whitespace",
			req:         request.UpdateAvatarRequest{AvatarURL: "   ", AvatarType: "uploaded"},
			setupMock:   func(m *MockAvatarRepository) {},
			wantErrCode: "ERR_PICK_EITHER_ID_OR_URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAvatarRepository)
			tt.setupMock(mockRepo)
			svc := NewAvatarService(mockRepo)

			resp, err := svc.UpdateAvatar(ctx, userID, tt.req)

			if tt.wantErrCode != "" {
				var appErr *ae.AppError
				if assert.True(t, errors.As(err, &appErr)) {
					assert.Equal(t, tt.wantErrCode, appErr.Code)
				}
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.wantURL, resp.AvatarURL)
				assert.Equal(t, tt.wantType, resp.AvatarType)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetUserAvatar(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		userID    uint
		setupMock func(m *MockAvatarRepository)
		wantURL   string
		wantType  string
	}{
		{
			name:   "should return user avatar URL and type when user has an avatar configured",
			userID: 1,
			setupMock: func(m *MockAvatarRepository) {
				m.On("GetUserAvatar", ctx, uint(1)).Return("https://skyfox.com/a.jpg", "predefined", nil)
			},
			wantURL:  "https://skyfox.com/a.jpg",
			wantType: "predefined",
		},
		{
			name:   "should return empty avatar when user has none set",
			userID: 2,
			setupMock: func(m *MockAvatarRepository) {
				m.On("GetUserAvatar", ctx, uint(2)).Return("", "", nil)
			},
			wantURL:  "",
			wantType: "",
		},
		{
			name:   "should return empty avatar URL and type when user ID does not exist in database",
			userID: 999,
			setupMock: func(m *MockAvatarRepository) {
				m.On("GetUserAvatar", ctx, uint(999)).Return("", "", nil)
			},
			wantURL:  "",
			wantType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAvatarRepository)
			tt.setupMock(mockRepo)
			svc := NewAvatarService(mockRepo)

			resp, err := svc.GetUserAvatar(ctx, tt.userID)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantURL, resp.AvatarURL)
			assert.Equal(t, tt.wantType, resp.AvatarType)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAssignRandomAvatar(t *testing.T) {
	ctx := context.Background()
	userID := uint(1)

	tests := []struct {
		name        string
		gender      string
		setupMock   func(m *MockAvatarRepository)
		wantErrCode string
	}{
		{
			name:   "should assign a random female avatar when user gender is female",
			gender: "female",
			setupMock: func(m *MockAvatarRepository) {
				m.On("GetRandomByGender", ctx, "female").Return(&model.PredefinedAvatar{ID: 1, URL: "f1.jpg"}, nil)
				m.On("UpdateUserAvatar", ctx, userID, "f1.jpg", "predefined").Return(nil)
			},
		},
		{
			name:   "should fall back to any available avatar when no avatar matches the requested gender",
			gender: "neutral",
			setupMock: func(m *MockAvatarRepository) {
				// Based on repo logic, if gender not found, service usually tries empty string or receives nil
				m.On("GetRandomByGender", ctx, "neutral").Return(nil, nil)
				m.On("GetRandomByGender", ctx, "").Return(&model.PredefinedAvatar{ID: 3, URL: "any.jpg"}, nil)
				m.On("UpdateUserAvatar", ctx, userID, "any.jpg", "predefined").Return(nil)
			},
		},
		{
			name:   "should return ErrAvatarNotFound when no predefined avatars exist in database",
			gender: "male",
			setupMock: func(m *MockAvatarRepository) {
				m.On("GetRandomByGender", ctx, "male").Return(nil, nil)
				m.On("GetRandomByGender", ctx, "").Return(nil, nil)
			},
			wantErrCode: "ERR_AVATAR_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAvatarRepository)
			tt.setupMock(mockRepo)
			svc := NewAvatarService(mockRepo)

			_, err := svc.AssignRandomAvatar(ctx, userID, tt.gender)
			if tt.wantErrCode != "" {
				var appErr *ae.AppError
				assert.True(t, errors.As(err, &appErr))
				assert.Equal(t, tt.wantErrCode, appErr.Code)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestListPredefinedAvatars(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		gender    string
		setupMock func(m *MockAvatarRepository)
		wantTotal int
	}{
		{
			name:   "should return only male avatars when gender filter is male",
			gender: "male",
			setupMock: func(m *MockAvatarRepository) {
				m.On("ListByGender", ctx, "male").Return([]model.PredefinedAvatar{{URL: "m1"}, {URL: "m2"}}, nil)
			},
			wantTotal: 2,
		},
		{
			name:   "should return all avatars when gender filter is empty",
			gender: "",
			setupMock: func(m *MockAvatarRepository) {
				m.On("ListAll", ctx).Return([]model.PredefinedAvatar{{URL: "m1"}, {URL: "f1"}}, nil)
			},
			wantTotal: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAvatarRepository)
			tt.setupMock(mockRepo)
			svc := NewAvatarService(mockRepo)

			resp, err := svc.ListPredefinedAvatars(ctx, tt.gender)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantTotal, resp.Total)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateAvatar_DeletesProfilePictureWhenSwitchingToPredefined(t *testing.T) {
	ctx := context.Background()
	userID := uint(1)
	predefined := model.PredefinedAvatar{ID: 10, Gender: "male", URL: "https://skyfox.com/pre.jpg"}

	tests := []struct {
		name              string
		req               request.UpdateAvatarRequest
		setupAvatarRepo   func(m *MockAvatarRepository)
		setupPFPDeleter   func(m *MockProfilePictureDeleter)
		wantPFPDeleteCall bool
	}{
		{
			name: "should delete profile picture when switching avatar type from uploaded to predefined",
			req:  request.UpdateAvatarRequest{PredefinedAvatarID: 10},
			setupAvatarRepo: func(m *MockAvatarRepository) {
				m.On("GetByID", ctx, int64(10)).Return(&predefined, nil)
				m.On("UpdateUserAvatar", ctx, userID, predefined.URL, "predefined").Return(nil)
			},
			setupPFPDeleter: func(m *MockProfilePictureDeleter) {
				m.On("DeleteByUserID", ctx, userID).Return(nil)
			},
			wantPFPDeleteCall: true,
		},
		{
			name: "should not delete profile picture when setting a custom uploaded avatar URL",
			req:  request.UpdateAvatarRequest{AvatarURL: "https://custom.com/my.jpg", AvatarType: "uploaded"},
			setupAvatarRepo: func(m *MockAvatarRepository) {
				m.On("UpdateUserAvatar", ctx, userID, "https://custom.com/my.jpg", "uploaded").Return(nil)
			},
			setupPFPDeleter:   func(m *MockProfilePictureDeleter) {},
			wantPFPDeleteCall: false,
		},
		{
			name: "should still update avatar successfully even when profile picture deletion fails",
			req:  request.UpdateAvatarRequest{PredefinedAvatarID: 10},
			setupAvatarRepo: func(m *MockAvatarRepository) {
				m.On("GetByID", ctx, int64(10)).Return(&predefined, nil)
				m.On("UpdateUserAvatar", ctx, userID, predefined.URL, "predefined").Return(nil)
			},
			setupPFPDeleter: func(m *MockProfilePictureDeleter) {
				m.On("DeleteByUserID", ctx, userID).Return(errors.New("delete error"))
			},
			wantPFPDeleteCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAvatarRepository)
			mockPFP := new(MockProfilePictureDeleter)
			tt.setupAvatarRepo(mockRepo)
			tt.setupPFPDeleter(mockPFP)

			svc := NewAvatarService(mockRepo, mockPFP)
			resp, err := svc.UpdateAvatar(ctx, userID, tt.req)

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			mockRepo.AssertExpectations(t)
			mockPFP.AssertExpectations(t)
		})
	}
}
