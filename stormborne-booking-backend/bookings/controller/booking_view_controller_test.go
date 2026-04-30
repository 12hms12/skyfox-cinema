package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"skyfox/bookings/common"
	"skyfox/bookings/dto/request"
	"skyfox/bookings/dto/response"
	"skyfox/bookings/controller/mocks"
	ae "skyfox/error"
)

func setupBookingViewRouter(mockSvc *mocks.MockBookingViewService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	controller := NewBookingViewController(mockSvc)
	router.GET("/admin/view-bookings", controller.GetBookings)
	return router
}


func setupCustomerBookingViewRouter(mockSvc *mocks.MockBookingViewService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	controller := NewBookingViewController(mockSvc)

	router.GET("/admin/view-bookings/customer/:onlineCustomerId", controller.GetAllBookingsCustomer)

	return router
}

func TestGetBookings(t *testing.T) {
	originalFlag := common.Feature_AdminViewBookings
	defer func() { common.Feature_AdminViewBookings = originalFlag }()

	mockItems := []response.AdminBookingViewItem{
		{
			ID:               1,
			TransactionID:    "txn-001",
			OnlineCustomerID: 10,
			ShowID:           5,
			MovieName:        "Inception",
			ShowDate:         "2026-03-17",
			StartTime:        "18:00:00",
			TotalPrice:       500.00,
			Status:           "CONFIRMED",
		},
	}
	mockResponse := &response.AdminBookingViewResponse{
		Bookings: mockItems,
		Total:    1,
	}

	tests := []struct {
		name           string
		featureFlag    bool
		queryParams    string
		mockFilter     request.AdminBookingViewRequest
		mockReturn     *response.AdminBookingViewResponse
		mockErr        error
		setupMock      bool
		expectedStatus int
		expectedTotal  int
	}{
		{
			name:           "feature flag disabled returns 503",
			featureFlag:    false,
			queryParams:    "",
			setupMock:      false,
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:        "returns all bookings with no filters",
			featureFlag: true,
			queryParams: "",
			mockFilter:  request.AdminBookingViewRequest{},
			mockReturn:  mockResponse,
			mockErr:     nil,
			setupMock:   true,
			expectedStatus: http.StatusOK,
			expectedTotal:  1,
		},
		{
			name:        "returns filtered bookings by movie name",
			featureFlag: true,
			queryParams: "?movieName=Inception",
			mockFilter:  request.AdminBookingViewRequest{MovieName: "Inception"},
			mockReturn:  mockResponse,
			mockErr:     nil,
			setupMock:   true,
			expectedStatus: http.StatusOK,
			expectedTotal:  1,
		},
		{
			name:        "returns filtered bookings by date range",
			featureFlag: true,
			queryParams: "?startDate=2026-03-17&endDate=2026-03-19",
			mockFilter:  request.AdminBookingViewRequest{StartDate: "2026-03-17", EndDate: "2026-03-19"},
			mockReturn:  mockResponse,
			mockErr:     nil,
			setupMock:   true,
			expectedStatus: http.StatusOK,
			expectedTotal:  1,
		},
		{
			name:        "returns empty list when no bookings found",
			featureFlag: true,
			queryParams: "",
			mockFilter:  request.AdminBookingViewRequest{},
			mockReturn:  &response.AdminBookingViewResponse{Bookings: []response.AdminBookingViewItem{}, Total: 0},
			mockErr:     nil,
			setupMock:   true,
			expectedStatus: http.StatusOK,
			expectedTotal:  0,
		},
		{
			name:        "returns 500 on service error",
			featureFlag: true,
			queryParams: "",
			mockFilter:  request.AdminBookingViewRequest{},
			mockReturn:  nil,
			mockErr:     ae.InternalServerError("InternalServerError","something went wrong",nil),
			setupMock:   true,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			common.Feature_AdminViewBookings = tt.featureFlag
			mockSvc := mocks.NewMockBookingViewService(t)

			if tt.setupMock {
				mockSvc.EXPECT().
					GetAllBookings(mock.Anything, tt.mockFilter).
					Return(tt.mockReturn, tt.mockErr).
					Once()
			}

			router := setupBookingViewRouter(mockSvc)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/admin/view-bookings"+tt.queryParams, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var body response.AdminBookingViewResponse
				json.Unmarshal(w.Body.Bytes(), &body)
				assert.Equal(t, tt.expectedTotal, body.Total)
			}
		})
	}
}

func TestGetAllBookingsCustomerController(t *testing.T) {
	originalFlag := common.Feature_AdminViewBookings
	defer func() { common.Feature_AdminViewBookings = originalFlag }()

	mockItems := []response.AdminBookingViewItem{
		{
			ID:               1,
			TransactionID:    "txn-001",
			OnlineCustomerID: 10,
			ShowID:           5,
			MovieName:        "Inception",
			ShowDate:         "2026-03-17",
			StartTime:        "18:00:00",
			TotalPrice:       500,
			Status:           "CONFIRMED",
		},
	}

	mockResponse := &response.AdminBookingViewResponse{
		Bookings: mockItems,
		Total:    1,
	}

	tests := []struct {
		name           string
		featureFlag    bool
		url            string
		mockFilter     request.CustomerBookingViewRequest
		mockReturn     *response.AdminBookingViewResponse
		mockErr        error
		setupMock      bool
		expectedStatus int
		expectedTotal  int
	}{
		{
			name:           "feature flag disabled",
			featureFlag:    false,
			url:            "/admin/view-bookings/customer/10",
			setupMock:      false,
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:        "successfully returns customer bookings",
			featureFlag: true,
			url:         "/admin/view-bookings/customer/10",
			mockFilter: request.CustomerBookingViewRequest{
				OnlineCustomerID: 10,
			},
			mockReturn:     mockResponse,
			mockErr:        nil,
			setupMock:      true,
			expectedStatus: http.StatusOK,
			expectedTotal:  1,
		},
		{
			name:        "movie name filter applied",
			featureFlag: true,
			url:         "/admin/view-bookings/customer/10?movieName=Inception",
			mockFilter: request.CustomerBookingViewRequest{
				OnlineCustomerID: 10,
				MovieName:        "Inception",
			},
			mockReturn:     mockResponse,
			mockErr:        nil,
			setupMock:      true,
			expectedStatus: http.StatusOK,
			expectedTotal:  1,
		},
		{
			name:           "invalid customer id",
			featureFlag:    true,
			url:            "/admin/view-bookings/customer/abc",
			setupMock:      false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "service error returns 500",
			featureFlag: true,
			url:         "/admin/view-bookings/customer/10",
			mockFilter: request.CustomerBookingViewRequest{
				OnlineCustomerID: 10,
			},
			mockReturn: nil,
			mockErr:    ae.InternalServerError("InternalError", "db failed", nil),
			setupMock:  true,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:        "empty bookings returned",
			featureFlag: true,
			url:         "/admin/view-bookings/customer/10",
			mockFilter: request.CustomerBookingViewRequest{
				OnlineCustomerID: 10,
			},
			mockReturn: &response.AdminBookingViewResponse{
				Bookings: []response.AdminBookingViewItem{},
				Total:    0,
			},
			mockErr:        nil,
			setupMock:      true,
			expectedStatus: http.StatusOK,
			expectedTotal:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			common.Feature_AdminViewBookings = tt.featureFlag
			mockSvc := mocks.NewMockBookingViewService(t)

			if tt.setupMock {
				mockSvc.EXPECT().
					GetAllBookingsCustomer(mock.Anything, tt.mockFilter).
					Return(tt.mockReturn, tt.mockErr).
					Once()
			}

			router := setupCustomerBookingViewRouter(mockSvc)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, tt.url, nil)

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var body response.AdminBookingViewResponse
				json.Unmarshal(w.Body.Bytes(), &body)
				assert.Equal(t, tt.expectedTotal, body.Total)
			}
		})
	}
}