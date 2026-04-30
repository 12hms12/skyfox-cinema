package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/payment", HandlePayment)
	return r
}

func makeRequest(t *testing.T, body map[string]interface{}) *httptest.ResponseRecorder {
	r := setupRouter()
	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, "/payment", bytes.NewBuffer(jsonBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestShouldReturnPaymentSuccessful_WhenAllFieldsAreValid(t *testing.T) {
	w := makeRequest(t, map[string]interface{}{
		"cardNumber": "1234567890123456",
		"cvv":        123,
		"name":       "Jane Doe",
		"expiryDate": "12/99",
	})

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Payment Successful", resp["message"])
}

func TestShouldReturnBadRequest_WhenCardNumberIsLessThan16Digits(t *testing.T) {
	w := makeRequest(t, map[string]interface{}{
		"cardNumber": "123456789012",
		"cvv":        123,
		"name":       "Jane Doe",
		"expiryDate": "12/99",
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Payment Failed: Card Number should be 16 digits long!", resp["message"])
}

func TestShouldReturnBadRequest_WhenCardNumberIsMoreThan16Digits(t *testing.T) {
	w := makeRequest(t, map[string]interface{}{
		"cardNumber": "12345678901234567",
		"cvv":        123,
		"name":       "Jane Doe",
		"expiryDate": "12/99",
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Payment Failed: Card Number should be 16 digits long!", resp["message"])
}

func TestShouldReturnBadRequest_WhenCardNumberIsMissing(t *testing.T) {
	w := makeRequest(t, map[string]interface{}{
		"cvv":        123,
		"name":       "Jane Doe",
		"expiryDate": "12/99",
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestShouldReturnBadRequest_WhenCVVIsLessThan3Digits(t *testing.T) {
	w := makeRequest(t, map[string]interface{}{
		"cardNumber": "1234567890123456",
		"cvv":        12,
		"name":       "Jane Doe",
		"expiryDate": "12/99",
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Payment Failed: cvv should be 3 digit long!", resp["message"])
}

func TestShouldReturnBadRequest_WhenCVVIsMoreThan3Digits(t *testing.T) {
	w := makeRequest(t, map[string]interface{}{
		"cardNumber": "1234567890123456",
		"cvv":        1234,
		"name":       "Jane Doe",
		"expiryDate": "12/99",
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Payment Failed: cvv should be 3 digit long!", resp["message"])
}

func TestShouldReturnBadRequest_WhenCVVIsMissing(t *testing.T) {
	w := makeRequest(t, map[string]interface{}{
		"cardNumber": "1234567890123456",
		"name":       "Jane Doe",
		"expiryDate": "12/99",
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestShouldReturnBadRequest_WhenNameIsLessThan3Characters(t *testing.T) {
	w := makeRequest(t, map[string]interface{}{
		"cardNumber": "1234567890123456",
		"cvv":        123,
		"name":       "Jo",
		"expiryDate": "12/99",
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Payment Failed: Name on card should have min 3 characters!", resp["message"])
}

func TestShouldReturnSuccess_WhenNameIsExactly3Characters(t *testing.T) {
	w := makeRequest(t, map[string]interface{}{
		"cardNumber": "1234567890123456",
		"cvv":        123,
		"name":       "Joe",
		"expiryDate": "12/99",
	})

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestShouldReturnBadRequest_WhenNameIsMissing(t *testing.T) {
	w := makeRequest(t, map[string]interface{}{
		"cardNumber": "1234567890123456",
		"cvv":        123,
		"expiryDate": "12/99",
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestShouldReturnBadRequest_WhenCardIsExpired(t *testing.T) {
	w := makeRequest(t, map[string]interface{}{
		"cardNumber": "1234567890123456",
		"cvv":        123,
		"name":       "Jane Doe",
		"expiryDate": "01/20",
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Payment Failed: Card Expired!", resp["message"])
}

func TestShouldReturnSuccess_WhenCardExpiresInCurrentMonth(t *testing.T) {
	w := makeRequest(t, map[string]interface{}{
		"cardNumber": "1234567890123456",
		"cvv":        123,
		"name":       "Jane Doe",
		"expiryDate": "03/26",
	})

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestShouldReturnBadRequest_WhenExpiryDateIsMissing(t *testing.T) {
	w := makeRequest(t, map[string]interface{}{
		"cardNumber": "1234567890123456",
		"cvv":        123,
		"name":       "Jane Doe",
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestShouldReturnBadRequest_WhenRequestBodyIsEmpty(t *testing.T) {
	w := makeRequest(t, map[string]interface{}{})

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestShouldReturnTrue_WhenExpiryDateIsInTheFuture(t *testing.T) {
	assert.True(t, isWithinExpiryDate("12/99"))
}

func TestShouldReturnFalse_WhenExpiryDateIsInThePast(t *testing.T) {
	assert.False(t, isWithinExpiryDate("01/20"))
}

func TestShouldReturnFalse_WhenExpiryDateFormatIsInvalid(t *testing.T) {
	assert.False(t, isWithinExpiryDate("invalid"))
}

func TestShouldReturnFalse_WhenExpiryMonthIsGreaterThan12(t *testing.T) {
	assert.False(t, isWithinExpiryDate("13/99"))
}

func TestShouldReturnFalse_WhenExpiryMonthIsZero(t *testing.T) {
	assert.False(t, isWithinExpiryDate("00/99"))
}
