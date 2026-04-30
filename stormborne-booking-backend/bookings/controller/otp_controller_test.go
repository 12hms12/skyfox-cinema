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

type MockSMSOTPService struct {
	mock.Mock
}

func (m *MockSMSOTPService) RequestOTP(ctx context.Context, phoneNumber, ip string) (*response.OTPSendResponse, error) {
	args := m.Called(phoneNumber, ip)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.OTPSendResponse), args.Error(1)
}

func (m *MockSMSOTPService) VerifyOTP(ctx context.Context, phoneNumber, code string) (*response.OTPVerifyResponse, error) {
	args := m.Called(phoneNumber, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.OTPVerifyResponse), args.Error(1)
}

func (m *MockSMSOTPService) ResendOTP(ctx context.Context, phoneNumber, ip string) (*response.OTPSendResponse, error) {
	args := m.Called(phoneNumber, ip)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.OTPSendResponse), args.Error(1)
}

func (m *MockSMSOTPService) ListAllOTPs(ctx context.Context) (*response.OTPListResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.OTPListResponse), args.Error(1)
}

func TestRequestSMSOTP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       interface{}
		setup      func(sms *MockSMSOTPService)
		wantStatus int
	}{
		{
			name: "should return 200 OK when a valid phone number is provided and the SMS OTP is generated successfully",
			body: map[string]string{"recipient": "+919876543210"},
			setup: func(sms *MockSMSOTPService) {
				sms.On("RequestOTP", "+919876543210", mock.AnythingOfType("string")).Return(
					&response.OTPSendResponse{Success: true, Message: "OTP sent to phone number", ExpiresIn: 300}, nil,
				)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "should return 400 Bad Request when the request body is missing the required recipient field",
			body:       map[string]string{},
			setup:      func(sms *MockSMSOTPService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "should return 400 Bad Request when the request body contains malformed JSON that cannot be parsed",
			body:       "!!!invalid-json!!!",
			setup:      func(sms *MockSMSOTPService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "should return 500 Internal Server Error when the SMS OTP service returns a non-AppError unexpected failure",
			body: map[string]string{"recipient": "+919876543210"},
			setup: func(sms *MockSMSOTPService) {
				sms.On("RequestOTP", "+919876543210", mock.AnythingOfType("string")).Return(nil, errors.New("unexpected"))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "should return the appropriate HTTP status code when the SMS OTP service returns an AppError with specific HTTP code",
			body: map[string]string{"recipient": "+919876543210"},
			setup: func(sms *MockSMSOTPService) {
				sms.On("RequestOTP", "+919876543210", mock.AnythingOfType("string")).Return(
					nil, ae.BadRequestError("OTPFailed", "phone number invalid", nil),
				)
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			smsSvc := new(MockSMSOTPService)
			tt.setup(smsSvc)

			ctrl := NewOTPController(smsSvc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			var bodyBytes []byte
			switch v := tt.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}
			c.Request, _ = http.NewRequest("POST", "/otp/sms/request", bytes.NewBuffer(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			ctrl.RequestSMSOTP(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestVerifySMSOTP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       interface{}
		setup      func(sms *MockSMSOTPService)
		wantStatus int
	}{
		{
			name: "should return 200 OK with verified=true when the correct OTP code is submitted for an active phone OTP",
			body: map[string]string{"recipient": "+919876543210", "code": "1234"},
			setup: func(sms *MockSMSOTPService) {
				sms.On("VerifyOTP", "+919876543210", "1234").Return(
					&response.OTPVerifyResponse{Success: true, Message: "phone number verified successfully", Verified: true}, nil,
				)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "should return 200 OK with verified=false when an incorrect OTP code is submitted for the phone number",
			body: map[string]string{"recipient": "+919876543210", "code": "0000"},
			setup: func(sms *MockSMSOTPService) {
				sms.On("VerifyOTP", "+919876543210", "0000").Return(
					&response.OTPVerifyResponse{Success: false, Message: "incorrect OTP", Verified: false}, nil,
				)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "should return 400 Bad Request when the request body is missing both recipient and code fields",
			body:       map[string]string{},
			setup:      func(sms *MockSMSOTPService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "should return 400 Bad Request when the OTP code field is present but has fewer than 4 digits",
			body:       map[string]string{"recipient": "+919876543210", "code": "12"},
			setup:      func(sms *MockSMSOTPService) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			smsSvc := new(MockSMSOTPService)
			tt.setup(smsSvc)

			ctrl := NewOTPController(smsSvc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			bodyBytes, _ := json.Marshal(tt.body)
			c.Request, _ = http.NewRequest("POST", "/otp/sms/verify", bytes.NewBuffer(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			ctrl.VerifySMSOTP(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestResendSMSOTP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       interface{}
		setup      func(sms *MockSMSOTPService)
		wantStatus int
	}{
		{
			name: "should return 200 OK when the SMS OTP is successfully regenerated and sent for the given phone number",
			body: map[string]string{"recipient": "+919876543210"},
			setup: func(sms *MockSMSOTPService) {
				sms.On("ResendOTP", "+919876543210", mock.AnythingOfType("string")).Return(
					&response.OTPSendResponse{Success: true, Message: "OTP sent to phone number", ExpiresIn: 300}, nil,
				)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "should return 400 Bad Request when the 3-resend rate limit has been exceeded within the 5-minute window",
			body: map[string]string{"recipient": "+919876543210"},
			setup: func(sms *MockSMSOTPService) {
				sms.On("ResendOTP", "+919876543210", mock.AnythingOfType("string")).Return(
					nil, ae.BadRequestError("ResendLimitExceeded", "resend limit exceeded. try again after 5 minutes", nil),
				)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "should return 400 Bad Request when the request body is missing the required recipient field for resend",
			body:       map[string]string{},
			setup:      func(sms *MockSMSOTPService) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			smsSvc := new(MockSMSOTPService)
			tt.setup(smsSvc)

			ctrl := NewOTPController(smsSvc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			bodyBytes, _ := json.Marshal(tt.body)
			c.Request, _ = http.NewRequest("POST", "/otp/sms/resend", bytes.NewBuffer(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			ctrl.ResendSMSOTP(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestListAllOTPs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setup      func(sms *MockSMSOTPService)
		wantStatus int
	}{
		{
			name: "should return 200 OK with all OTP records from the database for the illusion dashboard display",
			setup: func(sms *MockSMSOTPService) {
				sms.On("ListAllOTPs").Return(
					&response.OTPListResponse{OTPs: []response.OTPItem{
						{ID: 1, Code: "1234", Recipient: "+919876543210", Type: "SMS"},
					}}, nil,
				)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "should return 200 OK with an empty OTP list when no OTP records have been generated yet",
			setup: func(sms *MockSMSOTPService) {
				sms.On("ListAllOTPs").Return(&response.OTPListResponse{OTPs: []response.OTPItem{}}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "should return 500 Internal Server Error when the OTP listing service encounters a database connection failure",
			setup: func(sms *MockSMSOTPService) {
				sms.On("ListAllOTPs").Return(nil, errors.New("db connection lost"))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "should return the appropriate HTTP status code when the OTP listing service returns an AppError",
			setup: func(sms *MockSMSOTPService) {
				sms.On("ListAllOTPs").Return(nil, ae.InternalServerError("FetchFailed", "failed to fetch OTPs", nil))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			smsSvc := new(MockSMSOTPService)
			tt.setup(smsSvc)

			ctrl := NewOTPController(smsSvc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("GET", "/otp/all", nil)

			ctrl.ListAllOTPs(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
