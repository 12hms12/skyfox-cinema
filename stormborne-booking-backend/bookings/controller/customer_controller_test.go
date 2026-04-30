package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"skyfox/bookings/controller/mocks"
	"skyfox/bookings/dto/request"
	"skyfox/bookings/dto/response"
	servicemocks "skyfox/bookings/service/mocks"
	ae "skyfox/error"
)

func setupOnlineCustomerRouter(
	mockCustomerSvc *mocks.MockCustomerInterface,
	mockCaptchaSvc *servicemocks.MockCaptchaServiceInterface,
) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	ctrl := NewOnlineCustomerController(mockCustomerSvc, mockCaptchaSvc)
	router.POST("/customer/signup", ctrl.SignupController)
	router.POST("/customer/login", ctrl.LoginController)
	router.POST("/forgot-password", ctrl.ForgotPassword)
	router.POST("/reset-password", ctrl.ResetPassword)
	router.GET("/verify-reset-token", ctrl.VerifyResetToken)
	router.GET("/reset-password", ctrl.ResetPasswordPage)
	return router
}

var validSignupPayload = map[string]interface{}{
	"firstName":         "John",
	"lastName":          "Doe",
	"email":             "john@example.com",
	"username":          "john_doe",
	"countryCode":       "+91",
	"phoneNumber":       "9876543210",
	"password":          "Password123!",
	"age":               25,
	"gender":            "male",
	"captcha_id":        "test-captcha-id",
	"slider_position_x": 100.0,
}


func TestSignupController(t *testing.T) {
	tests := []struct {
		name               string
		body               interface{}
		setupMocks         func(*mocks.MockCustomerInterface, *servicemocks.MockCaptchaServiceInterface)
		expectedStatus     int
		expectCaptchaCalled bool
		expectSignupCalled  bool
	}{
		{
			name: "returns 201 on success",
			body: validSignupPayload,
			setupMocks: func(customerSvc *mocks.MockCustomerInterface, captchaSvc *servicemocks.MockCaptchaServiceInterface) {
				captchaSvc.On("Verify", "test-captcha-id", 100.0).Return(true)
				customerSvc.On("Signup", mock.Anything, mock.AnythingOfType("request.CustomerSignupRequest")).
					Return(&response.LoginResponse{}, nil)
			},
			expectedStatus:      http.StatusCreated,
			expectCaptchaCalled: true,
			expectSignupCalled:  true,
		},
		{
			name:               "returns 400 on invalid request body",
			body:               "invalid json",
			setupMocks:         func(_ *mocks.MockCustomerInterface, _ *servicemocks.MockCaptchaServiceInterface) {},
			expectedStatus:     http.StatusBadRequest,
			expectCaptchaCalled: false,
			expectSignupCalled:  false,
		},
		{
			name: "returns 400 on captcha failure",
			body: validSignupPayload,
			setupMocks: func(_ *mocks.MockCustomerInterface, captchaSvc *servicemocks.MockCaptchaServiceInterface) {
				captchaSvc.On("Verify", "test-captcha-id", 100.0).Return(false)
			},
			expectedStatus:      http.StatusBadRequest,
			expectCaptchaCalled: true,
			expectSignupCalled:  false,
		},
		{
			name: "returns 500 on generic service error",
			body: validSignupPayload,
			setupMocks: func(customerSvc *mocks.MockCustomerInterface, captchaSvc *servicemocks.MockCaptchaServiceInterface) {
				captchaSvc.On("Verify", "test-captcha-id", 100.0).Return(true)
				customerSvc.On("Signup", mock.Anything, mock.AnythingOfType("request.CustomerSignupRequest")).
					Return(nil, assert.AnError)
			},
			expectedStatus:      http.StatusInternalServerError,
			expectCaptchaCalled: true,
			expectSignupCalled:  true,
		},
		{
			name: "returns 400 on app error (duplicate user)",
			body: validSignupPayload,
			setupMocks: func(customerSvc *mocks.MockCustomerInterface, captchaSvc *servicemocks.MockCaptchaServiceInterface) {
				captchaSvc.On("Verify", "test-captcha-id", 100.0).Return(true)
				appErr := ae.BadRequestError("DuplicateUser", "user already exists", nil)
				customerSvc.On("Signup", mock.Anything, mock.AnythingOfType("request.CustomerSignupRequest")).
					Return(nil, appErr)
			},
			expectedStatus:      http.StatusBadRequest,
			expectCaptchaCalled: true,
			expectSignupCalled:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockCustomerSvc := new(mocks.MockCustomerInterface)
			mockCaptchaSvc := new(servicemocks.MockCaptchaServiceInterface)
			tc.setupMocks(mockCustomerSvc, mockCaptchaSvc)

			router := setupOnlineCustomerRouter(mockCustomerSvc, mockCaptchaSvc)

			var bodyBytes []byte
			if s, ok := tc.body.(string); ok {
				bodyBytes = []byte(s)
			} else {
				bodyBytes, _ = json.Marshal(tc.body)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/customer/signup", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectCaptchaCalled {
				mockCaptchaSvc.AssertExpectations(t)
			} else {
				mockCaptchaSvc.AssertNotCalled(t, "Verify", mock.Anything, mock.Anything)
			}

			if tc.expectSignupCalled {
				mockCustomerSvc.AssertExpectations(t)
			} else {
				mockCustomerSvc.AssertNotCalled(t, "Signup", mock.Anything, mock.Anything)
			}
		})
	}
}


func TestLoginController(t *testing.T) {
	validLoginBody := request.CustomerLoginRequest{
		Username: "john_doe",
		Password: "Password123!",
	}

	tests := []struct {
		name             string
		body             interface{}
		setupMocks       func(*mocks.MockCustomerInterface)
		expectedStatus   int
		expectLoginCalled bool
	}{
		{
			name: "returns 200 on success",
			body: validLoginBody,
			setupMocks: func(customerSvc *mocks.MockCustomerInterface) {
				customerSvc.On("Login", mock.Anything, mock.AnythingOfType("request.CustomerLoginRequest")).
					Return(&response.LoginResponse{Token: "jwt-token"}, nil)
			},
			expectedStatus:    http.StatusOK,
			expectLoginCalled: true,
		},
		{
			name:              "returns 400 on invalid request body",
			body:              "invalid json",
			setupMocks:        func(_ *mocks.MockCustomerInterface) {},
			expectedStatus:    http.StatusBadRequest,
			expectLoginCalled: false,
		},
		{
			name: "returns 500 on service error",
			body: validLoginBody,
			setupMocks: func(customerSvc *mocks.MockCustomerInterface) {
				customerSvc.On("Login", mock.Anything, mock.AnythingOfType("request.CustomerLoginRequest")).
					Return(nil, assert.AnError)
			},
			expectedStatus:    http.StatusInternalServerError,
			expectLoginCalled: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockCustomerSvc := new(mocks.MockCustomerInterface)
			mockCaptchaSvc := new(servicemocks.MockCaptchaServiceInterface)
			tc.setupMocks(mockCustomerSvc)

			router := setupOnlineCustomerRouter(mockCustomerSvc, mockCaptchaSvc)

			var bodyBytes []byte
			if s, ok := tc.body.(string); ok {
				bodyBytes = []byte(s)
			} else {
				bodyBytes, _ = json.Marshal(tc.body)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/customer/login", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectLoginCalled {
				mockCustomerSvc.AssertExpectations(t)
			} else {
				mockCustomerSvc.AssertNotCalled(t, "Login", mock.Anything, mock.Anything)
			}
		})
	}
}


func TestForgotPassword(t *testing.T) {
	tests := []struct {
		name                      string
		body                      interface{}
		setupMocks                func(*mocks.MockCustomerInterface)
		expectedStatus            int
		expectForgotPasswordCalled bool
	}{
		{
			name: "returns 200 on success",
			body: map[string]string{"email": "john@example.com"},
			setupMocks: func(customerSvc *mocks.MockCustomerInterface) {
				customerSvc.On("ForgotPassword", mock.Anything, "john@example.com").
					Return("https://reset-link.com/token", nil)
			},
			expectedStatus:             http.StatusOK,
			expectForgotPasswordCalled: true,
		},
		{
			name:                       "returns 400 on invalid request body",
			body:                       "invalid json",
			setupMocks:                 func(_ *mocks.MockCustomerInterface) {},
			expectedStatus:             http.StatusBadRequest,
			expectForgotPasswordCalled: false,
		},
		{
			name: "returns 400 on service error",
			body: map[string]string{"email": "john@example.com"},
			setupMocks: func(customerSvc *mocks.MockCustomerInterface) {
				customerSvc.On("ForgotPassword", mock.Anything, "john@example.com").
					Return("", assert.AnError)
			},
			expectedStatus:             http.StatusBadRequest,
			expectForgotPasswordCalled: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockCustomerSvc := new(mocks.MockCustomerInterface)
			mockCaptchaSvc := new(servicemocks.MockCaptchaServiceInterface)
			tc.setupMocks(mockCustomerSvc)

			router := setupOnlineCustomerRouter(mockCustomerSvc, mockCaptchaSvc)

			var bodyBytes []byte
			if s, ok := tc.body.(string); ok {
				bodyBytes = []byte(s)
			} else {
				bodyBytes, _ = json.Marshal(tc.body)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/forgot-password", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectForgotPasswordCalled {
				mockCustomerSvc.AssertExpectations(t)
			} else {
				mockCustomerSvc.AssertNotCalled(t, "ForgotPassword", mock.Anything, mock.Anything)
			}
		})
	}
}


func TestResetPassword(t *testing.T) {
	tests := []struct {
		name                     string
		body                     interface{}
		setupMocks               func(*mocks.MockCustomerInterface)
		expectedStatus           int
		expectResetPasswordCalled bool
	}{
		{
			name: "returns 200 on success",
			body: map[string]string{"token": "valid-token", "password": "NewPassword123!"},
			setupMocks: func(customerSvc *mocks.MockCustomerInterface) {
				customerSvc.On("ResetPassword", mock.Anything, "valid-token", "NewPassword123!").
					Return(nil)
			},
			expectedStatus:            http.StatusOK,
			expectResetPasswordCalled: true,
		},
		{
			name:                      "returns 400 on invalid request body",
			body:                      "invalid json",
			setupMocks:                func(_ *mocks.MockCustomerInterface) {},
			expectedStatus:            http.StatusBadRequest,
			expectResetPasswordCalled: false,
		},
		{
			name: "returns 400 on service error",
			body: map[string]string{"token": "invalid-token", "password": "NewPassword123!"},
			setupMocks: func(customerSvc *mocks.MockCustomerInterface) {
				customerSvc.On("ResetPassword", mock.Anything, "invalid-token", "NewPassword123!").
					Return(assert.AnError)
			},
			expectedStatus:            http.StatusBadRequest,
			expectResetPasswordCalled: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockCustomerSvc := new(mocks.MockCustomerInterface)
			mockCaptchaSvc := new(servicemocks.MockCaptchaServiceInterface)
			tc.setupMocks(mockCustomerSvc)

			router := setupOnlineCustomerRouter(mockCustomerSvc, mockCaptchaSvc)

			var bodyBytes []byte
			if s, ok := tc.body.(string); ok {
				bodyBytes = []byte(s)
			} else {
				bodyBytes, _ = json.Marshal(tc.body)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/reset-password", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectResetPasswordCalled {
				mockCustomerSvc.AssertExpectations(t)
			} else {
				mockCustomerSvc.AssertNotCalled(t, "ResetPassword", mock.Anything, mock.Anything, mock.Anything)
			}
		})
	}
}


func TestVerifyResetToken(t *testing.T) {
	tests := []struct {
		name                       string
		url                        string
		setupMocks                 func(*mocks.MockCustomerInterface)
		expectedStatus             int
		expectVerifyTokenCalled    bool
	}{
		{
			name: "returns 200 on success",
			url:  "/verify-reset-token?token=valid-token",
			setupMocks: func(customerSvc *mocks.MockCustomerInterface) {
				customerSvc.On("VerifyResetToken", mock.Anything, "valid-token").Return(nil)
			},
			expectedStatus:          http.StatusOK,
			expectVerifyTokenCalled: true,
		},
		{
			name:                    "returns 400 on missing token",
			url:                     "/verify-reset-token",
			setupMocks:              func(_ *mocks.MockCustomerInterface) {},
			expectedStatus:          http.StatusBadRequest,
			expectVerifyTokenCalled: false,
		},
		{
			name: "returns 400 on service error",
			url:  "/verify-reset-token?token=invalid-token",
			setupMocks: func(customerSvc *mocks.MockCustomerInterface) {
				customerSvc.On("VerifyResetToken", mock.Anything, "invalid-token").Return(assert.AnError)
			},
			expectedStatus:          http.StatusBadRequest,
			expectVerifyTokenCalled: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockCustomerSvc := new(mocks.MockCustomerInterface)
			mockCaptchaSvc := new(servicemocks.MockCaptchaServiceInterface)
			tc.setupMocks(mockCustomerSvc)

			router := setupOnlineCustomerRouter(mockCustomerSvc, mockCaptchaSvc)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, tc.url, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectVerifyTokenCalled {
				mockCustomerSvc.AssertExpectations(t)
			} else {
				mockCustomerSvc.AssertNotCalled(t, "VerifyResetToken", mock.Anything, mock.Anything)
			}
		})
	}
}


func TestResetPasswordPage(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectedToken  string
	}{
		{
			name:           "returns 200 with token in response body",
			url:            "/reset-password?token=some-token",
			expectedStatus: http.StatusOK,
			expectedToken:  "some-token",
		},
		{
			name:           "returns 200 with empty token when token not provided",
			url:            "/reset-password",
			expectedStatus: http.StatusOK,
			expectedToken:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockCustomerSvc := new(mocks.MockCustomerInterface)
			mockCaptchaSvc := new(servicemocks.MockCaptchaServiceInterface)

			router := setupOnlineCustomerRouter(mockCustomerSvc, mockCaptchaSvc)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, tc.url, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedToken != "" {
				var body map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &body)
				assert.Equal(t, tc.expectedToken, body["token"])
			}
		})
	}
}