package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"skyfox/bookings/dto/request"
	"skyfox/bookings/dto/response"
	ae "skyfox/error"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAvatarService struct {
	mock.Mock
}

func (m *MockAvatarService) UpdateAvatar(ctx context.Context, userID uint, req request.UpdateAvatarRequest) (*response.AvatarResponse, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.AvatarResponse), args.Error(1)
}

func (m *MockAvatarService) GetUserAvatar(ctx context.Context, userID uint) (*response.AvatarResponse, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.AvatarResponse), args.Error(1)
}

func (m *MockAvatarService) AssignRandomAvatar(ctx context.Context, userID uint, gender string) (*response.AvatarResponse, error) {
	args := m.Called(userID, gender)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.AvatarResponse), args.Error(1)
}

func (m *MockAvatarService) ListPredefinedAvatars(ctx context.Context, gender string) (*response.PredefinedAvatarListResponse, error) {
	args := m.Called(gender)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.PredefinedAvatarListResponse), args.Error(1)
}

func TestUpdateMyAvatar(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setup      func(c *gin.Context, svc *MockAvatarService)
		wantStatus int
	}{
		{
			name: "should return 401 Unauthorized when userID is missing from JWT context",
			setup: func(c *gin.Context, svc *MockAvatarService) {
				c.Request, _ = http.NewRequest("POST", "/", nil)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "should return 401 Unauthorized when userID in context is a string instead of numeric type",
			setup: func(c *gin.Context, svc *MockAvatarService) {
				c.Set("userID", "not-a-number")
				c.Request, _ = http.NewRequest("POST", "/", nil)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "should return 401 Unauthorized when userID in context is a negative int64",
			setup: func(c *gin.Context, svc *MockAvatarService) {
				c.Set("userID", int64(-1))
				c.Request, _ = http.NewRequest("POST", "/", nil)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "should return 401 Unauthorized when userID in context is a negative int",
			setup: func(c *gin.Context, svc *MockAvatarService) {
				c.Set("userID", int(-5))
				c.Request, _ = http.NewRequest("POST", "/", nil)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "should return 400 Bad Request when request body contains invalid JSON",
			setup: func(c *gin.Context, svc *MockAvatarService) {
				c.Set("userID", uint(1))
				c.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("!!invalid!!"))
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "should return 200 OK when avatar update succeeds and userID in context is uint",
			setup: func(c *gin.Context, svc *MockAvatarService) {
				req := request.UpdateAvatarRequest{PredefinedAvatarID: 1}
				svc.On("UpdateAvatar", uint(1), req).Return(
					&response.AvatarResponse{UserID: 1, AvatarURL: "https://img.com/a.jpg", AvatarType: "predefined"},
					nil,
				)
				c.Set("userID", uint(1))
				body, _ := json.Marshal(req)
				c.Request, _ = http.NewRequest("POST", "/", bytes.NewBuffer(body))
				c.Request.Header.Set("Content-Type", "application/json")
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "should return 200 OK when avatar update succeeds and userID in context is int64",
			setup: func(c *gin.Context, svc *MockAvatarService) {
				req := request.UpdateAvatarRequest{PredefinedAvatarID: 2}
				svc.On("UpdateAvatar", uint(5), req).Return(
					&response.AvatarResponse{UserID: 5, AvatarURL: "https://img.com/b.jpg", AvatarType: "predefined"},
					nil,
				)
				c.Set("userID", int64(5))
				body, _ := json.Marshal(req)
				c.Request, _ = http.NewRequest("POST", "/", bytes.NewBuffer(body))
				c.Request.Header.Set("Content-Type", "application/json")
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "should return 200 OK when avatar update succeeds and userID in context is int",
			setup: func(c *gin.Context, svc *MockAvatarService) {
				req := request.UpdateAvatarRequest{PredefinedAvatarID: 3}
				svc.On("UpdateAvatar", uint(7), req).Return(
					&response.AvatarResponse{UserID: 7, AvatarURL: "https://img.com/c.jpg", AvatarType: "predefined"},
					nil,
				)
				c.Set("userID", int(7))
				body, _ := json.Marshal(req)
				c.Request, _ = http.NewRequest("POST", "/", bytes.NewBuffer(body))
				c.Request.Header.Set("Content-Type", "application/json")
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "should return 404 Not Found when avatar service returns ErrAvatarNotFound",
			setup: func(c *gin.Context, svc *MockAvatarService) {
				req := request.UpdateAvatarRequest{PredefinedAvatarID: 999}
				svc.On("UpdateAvatar", uint(1), req).Return(nil, ae.ErrAvatarNotFound())
				c.Set("userID", uint(1))
				body, _ := json.Marshal(req)
				c.Request, _ = http.NewRequest("POST", "/", bytes.NewBuffer(body))
				c.Request.Header.Set("Content-Type", "application/json")
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "should return 500 Internal Server Error when avatar service returns non-AppError",
			setup: func(c *gin.Context, svc *MockAvatarService) {
				req := request.UpdateAvatarRequest{PredefinedAvatarID: 1}
				svc.On("UpdateAvatar", uint(1), req).Return(nil, errors.New("unexpected db error"))
				c.Set("userID", uint(1))
				body, _ := json.Marshal(req)
				c.Request, _ = http.NewRequest("POST", "/", bytes.NewBuffer(body))
				c.Request.Header.Set("Content-Type", "application/json")
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(MockAvatarService)
			ctrl := NewAvatarController(svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.setup(c, svc)
			ctrl.UpdateMyAvatar(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetMyAvatar(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setup      func(c *gin.Context, svc *MockAvatarService)
		wantStatus int
	}{
		{
			name: "should return 401 Unauthorized when userID is missing from JWT context",
			setup: func(c *gin.Context, svc *MockAvatarService) {
				c.Request, _ = http.NewRequest("GET", "/", nil)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "should return 401 Unauthorized when userID in context is a string instead of numeric type",
			setup: func(c *gin.Context, svc *MockAvatarService) {
				c.Set("userID", "bad-type")
				c.Request, _ = http.NewRequest("GET", "/", nil)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "should return 200 OK when user avatar is fetched successfully from service",
			setup: func(c *gin.Context, svc *MockAvatarService) {
				svc.On("GetUserAvatar", uint(1)).Return(
					&response.AvatarResponse{UserID: 1, AvatarURL: "https://skyfox.com/a.jpg", AvatarType: "predefined"},
					nil,
				)
				c.Set("userID", uint(1))
				c.Request, _ = http.NewRequest("GET", "/", nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "should return 500 Internal Server Error when avatar service returns an error",
			setup: func(c *gin.Context, svc *MockAvatarService) {
				svc.On("GetUserAvatar", uint(1)).Return(nil, errors.New("db error"))
				c.Set("userID", uint(1))
				c.Request, _ = http.NewRequest("GET", "/", nil)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(MockAvatarService)
			ctrl := NewAvatarController(svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.setup(c, svc)
			ctrl.GetMyAvatar(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestListPredefinedAvatars(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setup      func(svc *MockAvatarService, gender string)
		wantStatus int
	}{
		{
			name:  "should return 200 OK with gender-filtered avatars when gender query parameter is provided",
			query: "?gender=male",
			setup: func(svc *MockAvatarService, gender string) {
				svc.On("ListPredefinedAvatars", gender).Return(
					&response.PredefinedAvatarListResponse{
						Items: []response.PredefinedAvatarResponse{{ID: 1, Gender: "male", URL: "m1"}},
						Total: 1,
					},
					nil,
				)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "should return 200 OK with all avatars when no gender query parameter is provided",
			query: "",
			setup: func(svc *MockAvatarService, gender string) {
				svc.On("ListPredefinedAvatars", gender).Return(
					&response.PredefinedAvatarListResponse{
						Items: []response.PredefinedAvatarResponse{
							{ID: 1, Gender: "male", URL: "m1"},
							{ID: 2, Gender: "female", URL: "f1"},
						},
						Total: 2,
					},
					nil,
				)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "should return 500 Internal Server Error when predefined avatars service returns an error",
			query: "",
			setup: func(svc *MockAvatarService, gender string) {
				svc.On("ListPredefinedAvatars", gender).Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(MockAvatarService)
			ctrl := NewAvatarController(svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("GET", "/avatars"+tt.query, nil)

			var gender string
			if tt.query != "" {
				gender = c.Request.URL.Query().Get("gender")
			}
			tt.setup(svc, gender)

			ctrl.ListPredefinedAvatars(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
