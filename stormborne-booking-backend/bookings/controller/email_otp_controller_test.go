package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"skyfox/bookings/dto/response"
	ae "skyfox/error"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockEmailOTPService struct {
	mock.Mock
}

func (m *MockEmailOTPService) RequestOTP(ctx context.Context, email, ip string) (*response.OTPSendResponse, error) {
	args := m.Called(email, ip)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.OTPSendResponse), args.Error(1)
}

func (m *MockEmailOTPService) VerifyOTP(ctx context.Context, email, code string) (*response.OTPVerifyResponse, error) {
	args := m.Called(email, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.OTPVerifyResponse), args.Error(1)
}

func (m *MockEmailOTPService) ResendOTP(ctx context.Context, email, ip string) (*response.OTPSendResponse, error) {
	args := m.Called(email, ip)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.OTPSendResponse), args.Error(1)
}

func TestRequestEmailOTP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       interface{}
		setup      func(svc *MockEmailOTPService)
		wantStatus int
	}{
		{
			name: "should return 200 OK when a valid email is provided and the email OTP is generated successfully",
			body: map[string]string{"recipient": "user@example.com"},
			setup: func(svc *MockEmailOTPService) {
				svc.On("RequestOTP", "user@example.com", mock.AnythingOfType("string")).Return(
					&response.OTPSendResponse{Success: true, Message: "OTP sent to email address", ExpiresIn: 300}, nil,
				)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "should return 400 Bad Request when the request body is missing the required recipient field",
			body:       map[string]string{},
			setup:      func(svc *MockEmailOTPService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "should return 400 Bad Request when the request body contains malformed JSON",
			body:       "!!!invalid-json!!!",
			setup:      func(svc *MockEmailOTPService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "should return 500 Internal Server Error when the email OTP service returns a non-AppError failure",
			body: map[string]string{"recipient": "user@example.com"},
			setup: func(svc *MockEmailOTPService) {
				svc.On("RequestOTP", "user@example.com", mock.AnythingOfType("string")).Return(nil, errors.New("unexpected"))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "should return the appropriate HTTP status code when the email OTP service returns an AppError",
			body: map[string]string{"recipient": "user@example.com"},
			setup: func(svc *MockEmailOTPService) {
				svc.On("RequestOTP", "user@example.com", mock.AnythingOfType("string")).Return(
					nil, ae.BadRequestError("OTPFailed", "email invalid", nil),
				)
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emailSvc := new(MockEmailOTPService)
			tt.setup(emailSvc)

			ctrl := NewEmailOTPController(emailSvc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			var bodyBytes []byte
			switch v := tt.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}
			c.Request, _ = http.NewRequest("POST", "/otp/email/request", bytes.NewBuffer(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			ctrl.RequestEmailOTP(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestVerifyEmailOTP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       interface{}
		setup      func(svc *MockEmailOTPService)
		wantStatus int
	}{
		{
			name: "should return 200 OK with verified=true when the correct OTP code is submitted for an active email OTP",
			body: map[string]string{"recipient": "user@example.com", "code": "1234"},
			setup: func(svc *MockEmailOTPService) {
				svc.On("VerifyOTP", "user@example.com", "1234").Return(
					&response.OTPVerifyResponse{Success: true, Message: "email address verified successfully", Verified: true}, nil,
				)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "should return 200 OK with verified=false when an incorrect OTP code is submitted",
			body: map[string]string{"recipient": "user@example.com", "code": "0000"},
			setup: func(svc *MockEmailOTPService) {
				svc.On("VerifyOTP", "user@example.com", "0000").Return(
					&response.OTPVerifyResponse{Success: false, Message: "incorrect OTP", Verified: false}, nil,
				)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "should return 400 Bad Request when the request body is missing both recipient and code fields",
			body:       map[string]string{},
			setup:      func(svc *MockEmailOTPService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "should return 400 Bad Request when the OTP code field has fewer than 4 digits",
			body:       map[string]string{"recipient": "user@example.com", "code": "12"},
			setup:      func(svc *MockEmailOTPService) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emailSvc := new(MockEmailOTPService)
			tt.setup(emailSvc)

			ctrl := NewEmailOTPController(emailSvc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			bodyBytes, _ := json.Marshal(tt.body)
			c.Request, _ = http.NewRequest("POST", "/otp/email/verify", bytes.NewBuffer(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			ctrl.VerifyEmailOTP(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestResendEmailOTP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       interface{}
		setup      func(svc *MockEmailOTPService)
		wantStatus int
	}{
		{
			name: "should return 200 OK when the email OTP is successfully regenerated and sent",
			body: map[string]string{"recipient": "user@example.com"},
			setup: func(svc *MockEmailOTPService) {
				svc.On("ResendOTP", "user@example.com", mock.AnythingOfType("string")).Return(
					&response.OTPSendResponse{Success: true, Message: "OTP sent to email address", ExpiresIn: 300}, nil,
				)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "should return 400 Bad Request when the 3-resend rate limit has been exceeded within the 5-minute window",
			body: map[string]string{"recipient": "user@example.com"},
			setup: func(svc *MockEmailOTPService) {
				svc.On("ResendOTP", "user@example.com", mock.AnythingOfType("string")).Return(
					nil, ae.BadRequestError("ResendLimitExceeded", "resend limit exceeded. try again after 5 minutes", nil),
				)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "should return 400 Bad Request when the request body is missing the required recipient field",
			body:       map[string]string{},
			setup:      func(svc *MockEmailOTPService) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emailSvc := new(MockEmailOTPService)
			tt.setup(emailSvc)

			ctrl := NewEmailOTPController(emailSvc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			bodyBytes, _ := json.Marshal(tt.body)
			c.Request, _ = http.NewRequest("POST", "/otp/email/resend", bytes.NewBuffer(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			ctrl.ResendEmailOTP(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
