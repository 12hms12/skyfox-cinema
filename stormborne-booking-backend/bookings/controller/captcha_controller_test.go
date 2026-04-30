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

	"skyfox/bookings/dto/request"
	"skyfox/bookings/service/mocks"
)

func setupCaptchaRouter(mockSvc *mocks.MockCaptchaServiceInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	controller := NewCaptchaController(mockSvc)
	router.GET("/api/captcha/generate", controller.GenerateCaptcha)
	router.POST("/api/captcha/verify", controller.VerifyCaptcha)
	return router
}

func TestGenerateCaptcha_ReturnsOKWithCaptchaData(t *testing.T) {
	mockSvc := new(mocks.MockCaptchaServiceInterface)
	mockSvc.On("Generate").Return(
		"test-captcha-id", "master-base64", "tile-base64",
		100, 50, 60, 60, nil,
	)

	router := setupCaptchaRouter(mockSvc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/captcha/generate", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	assert.Equal(t, "test-captcha-id", body["captcha_id"])
	assert.Equal(t, "master-base64", body["master_image"])
	assert.Equal(t, "tile-base64", body["tile_image"])

	mockSvc.AssertExpectations(t)
}

func TestGenerateCaptcha_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	mockSvc := new(mocks.MockCaptchaServiceInterface)
	mockSvc.On("Generate").Return(
		"", "", "", 0, 0, 0, 0, assert.AnError,
	)

	router := setupCaptchaRouter(mockSvc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/captcha/generate", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	assert.Equal(t, "failed to generate captcha", body["error"])

	mockSvc.AssertExpectations(t)
}

func TestVerifyCaptcha_ReturnsTrueOnCorrectSliderPosition(t *testing.T) {
	mockSvc := new(mocks.MockCaptchaServiceInterface)
	mockSvc.On("Verify", "test-captcha-id", 100.0).Return(true)

	router := setupCaptchaRouter(mockSvc)

	body, _ := json.Marshal(request.CaptchaVerifyRequest{
		CaptchaID:       "test-captcha-id",
		SliderPositionX: 100.0,
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/captcha/verify", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]bool
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.True(t, resp["success"])

	mockSvc.AssertExpectations(t)
}

func TestVerifyCaptcha_ReturnsFalseOnWrongSliderPosition(t *testing.T) {
	mockSvc := new(mocks.MockCaptchaServiceInterface)
	mockSvc.On("Verify", "test-captcha-id", 999.0).Return(false)

	router := setupCaptchaRouter(mockSvc)

	body, _ := json.Marshal(request.CaptchaVerifyRequest{
		CaptchaID:       "test-captcha-id",
		SliderPositionX: 999.0,
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/captcha/verify", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]bool
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.False(t, resp["success"])

	mockSvc.AssertExpectations(t)
}

func TestVerifyCaptcha_ReturnsBadRequestOnInvalidBody(t *testing.T) {
	mockSvc := new(mocks.MockCaptchaServiceInterface)

	router := setupCaptchaRouter(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/captcha/verify", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	assert.Equal(t, "invalid request body", body["error"])

	mockSvc.AssertNotCalled(t, "Verify", mock.Anything, mock.Anything)
}

func TestVerifyCaptcha_ReturnsBadRequestOnMissingCaptchaID(t *testing.T) {
	mockSvc := new(mocks.MockCaptchaServiceInterface)

	router := setupCaptchaRouter(mockSvc)

	body, _ := json.Marshal(map[string]interface{}{
		"slider_position_x": 100.0,
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/captcha/verify", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	mockSvc.AssertNotCalled(t, "Verify", mock.Anything, mock.Anything)
}