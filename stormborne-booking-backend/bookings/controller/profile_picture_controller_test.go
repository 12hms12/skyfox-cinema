package controller

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"skyfox/bookings/common"
	"skyfox/bookings/dto/response"
	ae "skyfox/error"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testImageFieldName    = "image"
	testImageFileName     = "photo.jpg"
	testImageContentType  = "image/jpeg"
	testImageBytes        = "fake-image-bytes"
)

type MockProfileImageService struct {
	mock.Mock
}

func (m *MockProfileImageService) UploadProfileImage(ctx context.Context, userID uint, imageData []byte, contentType string) (*response.AvatarResponse, error) {
	args := m.Called(userID, imageData, contentType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.AvatarResponse), args.Error(1)
}

func (m *MockProfileImageService) GetProfileImage(ctx context.Context, userID uint) (*response.AvatarResponse, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.AvatarResponse), args.Error(1)
}

func buildMultipartRequest(t *testing.T, fieldName, filename, contentType, body string) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile(fieldName, filename)
	assert.NoError(t, err)
	_, err = io.WriteString(part, body)
	assert.NoError(t, err)
	assert.NoError(t, w.Close())

	req, err := http.NewRequest(http.MethodPost, "/profile/profile-image", &buf)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func TestUploadProfileImage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	featureFlagDefault := common.Feature_Upload_Profile_Image
	defer func() { common.Feature_Upload_Profile_Image = featureFlagDefault }()

	tests := []struct {
		name       string
		setup      func(c *gin.Context, svc *MockProfileImageService)
		wantStatus int
	}{
		{
			name: "should return 404 Not Found when Feature_Upload_Profile_Image feature flag is disabled",
			setup: func(c *gin.Context, svc *MockProfileImageService) {
				common.Feature_Upload_Profile_Image = false
				c.Request, _ = http.NewRequest(http.MethodPost, "/", nil)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "should return 401 Unauthorized when userID is missing from JWT context",
			setup: func(c *gin.Context, svc *MockProfileImageService) {
				common.Feature_Upload_Profile_Image = true
				c.Request, _ = http.NewRequest(http.MethodPost, "/", nil)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "should return 400 Bad Request when no image file is included in the request",
			setup: func(c *gin.Context, svc *MockProfileImageService) {
				common.Feature_Upload_Profile_Image = true
				c.Set("userID", uint(1))
				c.Request, _ = http.NewRequest(http.MethodPost, "/", nil)
				c.Request.Header.Set("Content-Type", "multipart/form-data")
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "should return 200 OK when profile image is uploaded successfully",
			setup: func(c *gin.Context, svc *MockProfileImageService) {
				common.Feature_Upload_Profile_Image = true
				c.Set("userID", uint(1))
				svc.On("UploadProfileImage", uint(1), mock.Anything, mock.Anything).Return(
					&response.AvatarResponse{UserID: 1, AvatarURL: "data:image/jpeg;base64,abc", AvatarType: "uploaded"},
					nil,
				)
				c.Request = buildMultipartRequest(t, testImageFieldName, testImageFileName, testImageContentType, testImageBytes)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "should return 400 Bad Request when service returns ErrImageTooLarge",
			setup: func(c *gin.Context, svc *MockProfileImageService) {
				common.Feature_Upload_Profile_Image = true
				c.Set("userID", uint(1))
				svc.On("UploadProfileImage", uint(1), mock.Anything, mock.Anything).Return(nil, ae.ErrImageTooLarge())
				c.Request = buildMultipartRequest(t, testImageFieldName, testImageFileName, testImageContentType, testImageBytes)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "should return 500 Internal Server Error when service returns a non-AppError",
			setup: func(c *gin.Context, svc *MockProfileImageService) {
				common.Feature_Upload_Profile_Image = true
				c.Set("userID", uint(1))
				svc.On("UploadProfileImage", uint(1), mock.Anything, mock.Anything).Return(nil, errors.New("unexpected error"))
				c.Request = buildMultipartRequest(t, testImageFieldName, testImageFileName, testImageContentType, testImageBytes)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(MockProfileImageService)
			ctrl := NewProfileImageController(svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			tt.setup(c, svc)
			ctrl.UploadProfileImage(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			svc.AssertExpectations(t)
		})
	}
}

func TestGetProfileImage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	featureFlagDefault := common.Feature_Upload_Profile_Image
	defer func() { common.Feature_Upload_Profile_Image = featureFlagDefault }()

	tests := []struct {
		name       string
		setup      func(c *gin.Context, svc *MockProfileImageService)
		wantStatus int
	}{
		{
			name: "should return 404 Not Found when Feature_Upload_Profile_Image feature flag is disabled",
			setup: func(c *gin.Context, svc *MockProfileImageService) {
				common.Feature_Upload_Profile_Image = false
				c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "should return 401 Unauthorized when userID is missing from JWT context",
			setup: func(c *gin.Context, svc *MockProfileImageService) {
				common.Feature_Upload_Profile_Image = true
				c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "should return 200 OK when profile image is fetched successfully",
			setup: func(c *gin.Context, svc *MockProfileImageService) {
				common.Feature_Upload_Profile_Image = true
				c.Set("userID", uint(2))
				svc.On("GetProfileImage", uint(2)).Return(
					&response.AvatarResponse{UserID: 2, AvatarURL: "data:image/jpeg;base64,xyz", AvatarType: "uploaded"},
					nil,
				)
				c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "should return 500 Internal Server Error when service returns a non-AppError",
			setup: func(c *gin.Context, svc *MockProfileImageService) {
				common.Feature_Upload_Profile_Image = true
				c.Set("userID", uint(2))
				svc.On("GetProfileImage", uint(2)).Return(nil, errors.New("db failure"))
				c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(MockProfileImageService)
			ctrl := NewProfileImageController(svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			tt.setup(c, svc)
			ctrl.GetProfileImage(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			svc.AssertExpectations(t)
		})
	}
}
