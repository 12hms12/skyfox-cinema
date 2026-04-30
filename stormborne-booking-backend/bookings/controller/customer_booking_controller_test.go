package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"skyfox/bookings/common"
	"skyfox/bookings/dto/response"
	ae "skyfox/error"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCustomerBookingService struct {
	mock.Mock
}

func (m *mockCustomerBookingService) CreateBooking(ctx context.Context, customerID int64, showID int, seatStatusIDs []int) (*response.CreateCustomerBookingResponse, error) {
	ret := m.Called(customerID, showID, seatStatusIDs)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*response.CreateCustomerBookingResponse), ret.Error(1)
}

func (m *mockCustomerBookingService) ProcessPayment(ctx context.Context, bookingID int, paymentSuccess bool) (*response.PaymentResponse, error) {
	ret := m.Called(bookingID, paymentSuccess)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*response.PaymentResponse), ret.Error(1)
}

func (m *mockCustomerBookingService) GetBookingDetails(ctx context.Context, bookingID int) (*response.BookingDetailsResponse, error) {
	ret := m.Called(bookingID)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*response.BookingDetailsResponse), ret.Error(1)
}

func setupCustomerBookingRouter(svc *mockCustomerBookingService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	ctrl := NewCustomerBookingController(svc)

	r.POST("/customer/bookings", func(c *gin.Context) {
		c.Set("userID", int64(1))
		ctrl.CreateBooking(c)
	})
	r.POST("/customer/bookings/:bookingId/payment", ctrl.ProcessPayment)
	r.GET("/customer/bookings/:bookingId", ctrl.GetBookingDetails)
	return r
}

func (m *mockCustomerBookingService) CheckIn(ctx context.Context, bookingID int) error {
	ret := m.Called(bookingID)
	return ret.Error(0)
}

func TestCreateBookingController(t *testing.T) {
	expiresAt := time.Now().Add(15 * time.Minute)

	successResponse := &response.CreateCustomerBookingResponse{
		BookingID:        42,
		BookingReference: "SKY-abc12345",
		Seats:            []response.BookingSeatInfo{{SeatID: 5, Label: "A5", SeatType: "REGULAR", Price: 200.0}},
		TotalPrice:       200.0,
		Status:           "PENDING",
		ExpiresAt:        expiresAt,
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		setup          func(svc *mockCustomerBookingService)
		wantStatusCode int
		wantBodyKey    string
		wantBodyValue  string
	}{
		{
			name:           "returns 400 Bad Request when request body is missing the required showId field",
			requestBody:    map[string]interface{}{"seatIds": []int{5, 6}},
			setup:          func(svc *mockCustomerBookingService) {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "returns 400 Bad Request when seatIds list is empty (minimum 1 seat must be selected)",
			requestBody:    map[string]interface{}{"showId": 10, "seatIds": []int{}},
			setup:          func(svc *mockCustomerBookingService) {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:        "returns 400 when the requested seats are no longer available because another customer booked them concurrently",
			requestBody: map[string]interface{}{"showId": 10, "seatIds": []int{5, 6}},
			setup: func(svc *mockCustomerBookingService) {
				svc.On("CreateBooking", int64(1), 10, []int{5, 6}).
					Return(nil, ae.BadRequestError("SeatsUnavailable", "one or more seats are no longer available", nil))
			},
			wantStatusCode: http.StatusBadRequest,
			wantBodyKey:    "error",
			wantBodyValue:  "one or more seats are no longer available",
		},
		{
			name:        "returns 201 Created with PENDING status, bookingId, bookingReference and expiresAt when booking succeeds",
			requestBody: map[string]interface{}{"showId": 10, "seatIds": []int{5, 6}},
			setup: func(svc *mockCustomerBookingService) {
				svc.On("CreateBooking", int64(1), 10, []int{5, 6}).Return(successResponse, nil)
			},
			wantStatusCode: http.StatusCreated,
			wantBodyKey:    "status",
			wantBodyValue:  "PENDING",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockCustomerBookingService{}
			tt.setup(svc)
			router := setupCustomerBookingRouter(svc)

			bodyBytes, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/customer/bookings", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)

			if tt.wantBodyKey != "" {
				var body map[string]interface{}
				_ = json.Unmarshal(w.Body.Bytes(), &body)
				assert.Equal(t, tt.wantBodyValue, body[tt.wantBodyKey])
			}

			svc.AssertExpectations(t)
		})
	}
}

func TestProcessPaymentController(t *testing.T) {
	successPaymentResponse := &response.PaymentResponse{
		PaymentID: "PAY-xyz98765",
		BookingID: 42,
		Amount:    200.0,
		Currency:  "INR",
		Status:    "SUCCESS",
	}

	failedPaymentResponse := &response.PaymentResponse{
		PaymentID: "PAY-xyz98765",
		BookingID: 42,
		Amount:    200.0,
		Currency:  "INR",
		Status:    "FAILED",
	}

	tests := []struct {
		name           string
		bookingIDParam string
		requestBody    interface{}
		setup          func(svc *mockCustomerBookingService)
		wantStatusCode int
		wantBodyKey    string
		wantBodyValue  string
	}{
		{
			name:           "returns 400 Bad Request when booking ID path parameter is not a valid integer",
			bookingIDParam: "not-a-number",
			requestBody:    map[string]bool{"success": true},
			setup:          func(svc *mockCustomerBookingService) {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "returns 404 Not Found when no booking exists for the given booking ID",
			bookingIDParam: "99",
			requestBody:    map[string]bool{"success": true},
			setup: func(svc *mockCustomerBookingService) {
				svc.On("ProcessPayment", 99, true).Return(nil, ae.NotFoundError("BookingNotFound", "booking not found", nil))
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "returns 200 OK with payment status SUCCESS when customer confirms payment by selecting yes",
			bookingIDParam: "42",
			requestBody:    map[string]bool{"success": true},
			setup: func(svc *mockCustomerBookingService) {
				svc.On("ProcessPayment", 42, true).Return(successPaymentResponse, nil)
			},
			wantStatusCode: http.StatusOK,
			wantBodyKey:    "status",
			wantBodyValue:  "SUCCESS",
		},
		{
			name:           "returns 200 OK with payment status FAILED and triggers seat release when customer declines payment by selecting no",
			bookingIDParam: "42",
			requestBody:    map[string]bool{"success": false},
			setup: func(svc *mockCustomerBookingService) {
				svc.On("ProcessPayment", 42, false).Return(failedPaymentResponse, nil)
			},
			wantStatusCode: http.StatusOK,
			wantBodyKey:    "status",
			wantBodyValue:  "FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockCustomerBookingService{}
			tt.setup(svc)
			router := setupCustomerBookingRouter(svc)

			bodyBytes, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/customer/bookings/"+tt.bookingIDParam+"/payment", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)

			if tt.wantBodyKey != "" {
				var body map[string]interface{}
				_ = json.Unmarshal(w.Body.Bytes(), &body)
				assert.Equal(t, tt.wantBodyValue, body[tt.wantBodyKey])
			}

			svc.AssertExpectations(t)
		})
	}
}

func TestGetBookingDetailsController(t *testing.T) {
	detailsResponse := &response.BookingDetailsResponse{
		BookingID:        42,
		BookingReference: "SKY-abc12345",
		MovieID:          "tt1375666",
		ShowTime:         "2025-03-10 18:00",
		ScreenName:       "Screen 1",
		Seats:            []response.BookingSeatInfo{{SeatID: 5, Label: "A5", SeatType: "REGULAR"}},
		TotalPrice:       200.0,
		PaymentStatus:    "SUCCESS",
		BookingStatus:    "CONFIRMED",
	}

	tests := []struct {
		name           string
		bookingIDParam string
		setup          func(svc *mockCustomerBookingService)
		wantStatusCode int
		validate       func(t *testing.T, body map[string]interface{})
	}{
		{
			name:           "returns 400 Bad Request when the booking ID path parameter is not a valid integer",
			bookingIDParam: "invalid",
			setup:          func(svc *mockCustomerBookingService) {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "returns 404 Not Found when no booking record exists for the provided booking ID",
			bookingIDParam: "999",
			setup: func(svc *mockCustomerBookingService) {
				svc.On("GetBookingDetails", 999).Return(nil, ae.NotFoundError("BookingNotFound", "booking not found", nil))
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "returns 200 OK with complete booking details including movieId, screenName, seats, payment and booking statuses",
			bookingIDParam: "42",
			setup: func(svc *mockCustomerBookingService) {
				svc.On("GetBookingDetails", 42).Return(detailsResponse, nil)
			},
			wantStatusCode: http.StatusOK,
			validate: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "tt1375666", body["movieId"])
				assert.Equal(t, "Screen 1", body["screenName"])
				assert.Equal(t, "SUCCESS", body["paymentStatus"])
				assert.Equal(t, "CONFIRMED", body["bookingStatus"])
				assert.Equal(t, "SKY-abc12345", body["bookingReference"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockCustomerBookingService{}
			tt.setup(svc)
			router := setupCustomerBookingRouter(svc)

			req := httptest.NewRequest(http.MethodGet, "/customer/bookings/"+tt.bookingIDParam, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)

			if tt.validate != nil {
				var body map[string]interface{}
				_ = json.Unmarshal(w.Body.Bytes(), &body)
				tt.validate(t, body)
			}

			svc.AssertExpectations(t)
		})
	}
}

func TestFeatureFlag_CustomerBooking(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		wantStatusCode int
	}{
		{
			name:           "returns 404 for POST /customer/bookings when Feature_CustomerBooking flag is disabled",
			method:         http.MethodPost,
			path:           "/customer/bookings",
			body:           map[string]interface{}{"showId": 10, "seatIds": []int{5}},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "returns 404 for POST payment endpoint when Feature_CustomerBooking flag is disabled",
			method:         http.MethodPost,
			path:           "/customer/bookings/42/payment",
			body:           map[string]bool{"success": true},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "returns 404 for GET booking details endpoint when Feature_CustomerBooking flag is disabled",
			method:         http.MethodGet,
			path:           "/customer/bookings/42",
			wantStatusCode: http.StatusNotFound,
		},
	}

	original := common.Feature_CustomerBooking
	common.Feature_CustomerBooking = false
	defer func() { common.Feature_CustomerBooking = original }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockCustomerBookingService{}
			router := setupCustomerBookingRouter(svc)

			var bodyBytes []byte
			if tt.body != nil {
				bodyBytes, _ = json.Marshal(tt.body)
			}
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)
			svc.AssertExpectations(t)
		})
	}
}
