package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"skyfox/bookings/model"
	"skyfox/bookings/repository/mocks"
	ae "skyfox/error"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	weekdayShowForBooking = &model.Show{
		Id:      10,
		MovieId: "tt1375666",
		Date:    time.Date(2025, time.March, 10, 18, 0, 0, 0, time.UTC),
		Screen:  model.Screen{ID: 1, ScreenName: "Screen 1"},
	}

	weekendShowForBooking = &model.Show{
		Id:      11,
		MovieId: "tt1375666",
		Date:    time.Date(2025, time.March, 15, 18, 0, 0, 0, time.UTC),
		Screen:  model.Screen{ID: 1, ScreenName: "Screen 1"},
	}

	mockPricingForBooking = []model.ShowPricing{
		{ID: 1, ShowID: 10, SeatType: "REGULAR", Price: 200.0},
		{ID: 2, ShowID: 10, SeatType: "PREMIUM", Price: 350.0},
	}

	availableSeatStatuses = []model.ShowSeatStatus{
		{ID: 5, ShowID: 10, SeatID: 5, Status: "AVAILABLE", Seat: model.Seat{ID: 5, RowNumber: 1, ColumnNumber: 5, SeatType: "REGULAR"}},
		{ID: 6, ShowID: 10, SeatID: 6, Status: "AVAILABLE", Seat: model.Seat{ID: 6, RowNumber: 2, ColumnNumber: 3, SeatType: "PREMIUM"}},
	}

	pendingBooking = &model.Booking{
		ID:            42,
		TransactionID: "SKY-abc12345",
		ShowID:        10,
		TotalPrice:    550.0,
		Status:        "PENDING",
		ExpiresAt:     time.Now().Add(15 * time.Minute),
		Show: model.Show{
			Id:      10,
			MovieId: "tt1375666",
			Date:    time.Date(2025, time.March, 10, 18, 0, 0, 0, time.UTC),
			Screen:  model.Screen{ScreenName: "Screen 1"},
		},
		BookedSeats: []model.BookedSeat{
			{
				ID:               1,
				BookingID:        42,
				ShowSeatStatusID: 5,
				ShowSeatStatus: model.ShowSeatStatus{
					Seat: model.Seat{ID: 5, RowNumber: 1, ColumnNumber: 5, SeatType: "REGULAR"},
				},
			},
		},
	}
)

func TestCreateBooking(t *testing.T) {
	ctx := context.Background()
	seatStatusIDs := []int{5, 6}

	tests := []struct {
		name         string
		customerID   int64
		showID       int
		seatIDs      []int
		setup        func(cbRepo *mocks.MockCustomerBookingRepository, seatRepo *mocks.MockSeatRepository)
		wantErr      bool
		errCode      string
		validateResp func(t *testing.T, resp interface{})
	}{
		{
			name:       "returns error when show is not found in the repository",
			customerID: 1,
			showID:     10,
			seatIDs:    seatStatusIDs,
			setup: func(cbRepo *mocks.MockCustomerBookingRepository, seatRepo *mocks.MockSeatRepository) {
				seatRepo.On("GetShowByID", ctx, uint(10)).Return(nil, errors.New("show not found"))
			},
			wantErr: true,
		},
		{
			name:       "returns error when pricing data cannot be fetched for the show",
			customerID: 1,
			showID:     10,
			seatIDs:    seatStatusIDs,
			setup: func(cbRepo *mocks.MockCustomerBookingRepository, seatRepo *mocks.MockSeatRepository) {
				seatRepo.On("GetShowByID", ctx, uint(10)).Return(weekdayShowForBooking, nil)
				seatRepo.On("GetPricingByShowID", ctx, uint(10)).Return(nil, errors.New("pricing error"))
			},
			wantErr: true,
		},
		{
			name:       "returns error when seat statuses cannot be fetched for the requested seat IDs",
			customerID: 1,
			showID:     10,
			seatIDs:    seatStatusIDs,
			setup: func(cbRepo *mocks.MockCustomerBookingRepository, seatRepo *mocks.MockSeatRepository) {
				seatRepo.On("GetShowByID", ctx, uint(10)).Return(weekdayShowForBooking, nil)
				seatRepo.On("GetPricingByShowID", ctx, uint(10)).Return(mockPricingForBooking, nil)
				cbRepo.On("GetSeatStatusesWithSeatByIDs", ctx, seatStatusIDs).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name:       "returns bad request error when fewer seat statuses are returned than requested (seats not found for this show)",
			customerID: 1,
			showID:     10,
			seatIDs:    seatStatusIDs,
			setup: func(cbRepo *mocks.MockCustomerBookingRepository, seatRepo *mocks.MockSeatRepository) {
				seatRepo.On("GetShowByID", ctx, uint(10)).Return(weekdayShowForBooking, nil)
				seatRepo.On("GetPricingByShowID", ctx, uint(10)).Return(mockPricingForBooking, nil)
				cbRepo.On("GetSeatStatusesWithSeatByIDs", ctx, seatStatusIDs).Return(availableSeatStatuses[:1], nil)
			},
			wantErr: true,
		},
		{
			name:       "returns error when the atomic seat lock and booking creation fails in the repository",
			customerID: 1,
			showID:     10,
			seatIDs:    seatStatusIDs,
			setup: func(cbRepo *mocks.MockCustomerBookingRepository, seatRepo *mocks.MockSeatRepository) {
				seatRepo.On("GetShowByID", ctx, uint(10)).Return(weekdayShowForBooking, nil)
				seatRepo.On("GetPricingByShowID", ctx, uint(10)).Return(mockPricingForBooking, nil)
				cbRepo.On("GetSeatStatusesWithSeatByIDs", ctx, seatStatusIDs).Return(availableSeatStatuses, nil)
				cbRepo.On("CreateBookingWithSeatLock", ctx, int64(1), 10, seatStatusIDs, 550.0, mock.AnythingOfType("time.Time"), mock.AnythingOfType("string")).
					Return(nil, ae.BadRequestError("SeatsUnavailable", "one or more seats are no longer available", nil))
			},
			wantErr: true,
		},
		{
			name:       "successfully creates a booking with correct total price for a weekday show (no surcharge)",
			customerID: 1,
			showID:     10,
			seatIDs:    seatStatusIDs,
			setup: func(cbRepo *mocks.MockCustomerBookingRepository, seatRepo *mocks.MockSeatRepository) {
				seatRepo.On("GetShowByID", ctx, uint(10)).Return(weekdayShowForBooking, nil)
				seatRepo.On("GetPricingByShowID", ctx, uint(10)).Return(mockPricingForBooking, nil)
				cbRepo.On("GetSeatStatusesWithSeatByIDs", ctx, seatStatusIDs).Return(availableSeatStatuses, nil)
				cbRepo.On("CreateBookingWithSeatLock", ctx, int64(1), 10, seatStatusIDs, 550.0, mock.AnythingOfType("time.Time"), mock.AnythingOfType("string")).
					Return(&model.Booking{ID: 42, TransactionID: "SKY-abc12345", Status: "PENDING", TotalPrice: 550.0, ExpiresAt: time.Now().Add(15 * time.Minute)}, nil)
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp interface{}) {
				r := resp.(*struct {
					BookingID  int
					TotalPrice float64
					Status     string
					SeatCount  int
				})
				assert.Equal(t, 42, r.BookingID)
				assert.Equal(t, 550.0, r.TotalPrice)
				assert.Equal(t, "PENDING", r.Status)
				assert.Equal(t, 2, r.SeatCount)
			},
		},
		{
			name:       "applies weekend surcharge of 50 INR per seat when show falls on Saturday or Sunday",
			customerID: 1,
			showID:     11,
			seatIDs:    []int{5},
			setup: func(cbRepo *mocks.MockCustomerBookingRepository, seatRepo *mocks.MockSeatRepository) {
				seatRepo.On("GetShowByID", ctx, uint(11)).Return(weekendShowForBooking, nil)
				seatRepo.On("GetPricingByShowID", ctx, uint(11)).Return([]model.ShowPricing{
					{ID: 3, ShowID: 11, SeatType: "REGULAR", Price: 200.0},
				}, nil)
				cbRepo.On("GetSeatStatusesWithSeatByIDs", ctx, []int{5}).Return([]model.ShowSeatStatus{availableSeatStatuses[0]}, nil)
				cbRepo.On("CreateBookingWithSeatLock", ctx, int64(1), 11, []int{5}, 250.0, mock.AnythingOfType("time.Time"), mock.AnythingOfType("string")).
					Return(&model.Booking{ID: 43, TransactionID: "SKY-xyz98765", Status: "PENDING", TotalPrice: 250.0, ExpiresAt: time.Now().Add(15 * time.Minute)}, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cbRepo := &mocks.MockCustomerBookingRepository{}
			seatRepo := &mocks.MockSeatRepository{}
			tt.setup(cbRepo, seatRepo)

			svc := NewCustomerBookingService(cbRepo, seatRepo)
			resp, err := svc.CreateBooking(ctx, tt.customerID, tt.showID, tt.seatIDs)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}

			cbRepo.AssertExpectations(t)
			seatRepo.AssertExpectations(t)
		})
	}
}

func TestProcessPayment(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		bookingID      int
		paymentSuccess bool
		setup          func(cbRepo *mocks.MockCustomerBookingRepository)
		wantErr        bool
		wantStatus     string
	}{
		{
			name:           "returns error when booking is not found before processing payment",
			bookingID:      42,
			paymentSuccess: true,
			setup: func(cbRepo *mocks.MockCustomerBookingRepository) {
				cbRepo.On("GetBookingByID", ctx, 42).Return(nil, ae.NotFoundError("BookingNotFound", "booking not found", nil))
			},
			wantErr: true,
		},
		{
			name:           "returns error when confirming payment fails in the repository",
			bookingID:      42,
			paymentSuccess: true,
			setup: func(cbRepo *mocks.MockCustomerBookingRepository) {
				cbRepo.On("GetBookingByID", ctx, 42).Return(pendingBooking, nil)
				cbRepo.On("ConfirmBooking", ctx, 42).Return(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name:           "returns error when cancelling payment fails in the repository",
			bookingID:      42,
			paymentSuccess: false,
			setup: func(cbRepo *mocks.MockCustomerBookingRepository) {
				cbRepo.On("GetBookingByID", ctx, 42).Return(pendingBooking, nil)
				cbRepo.On("CancelBooking", ctx, 42).Return(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name:           "returns SUCCESS status and payment details when customer proceeds with payment (yes)",
			bookingID:      42,
			paymentSuccess: true,
			setup: func(cbRepo *mocks.MockCustomerBookingRepository) {
				cbRepo.On("GetBookingByID", ctx, 42).Return(pendingBooking, nil)
				cbRepo.On("ConfirmBooking", ctx, 42).Return(nil)
			},
			wantErr:    false,
			wantStatus: "SUCCESS",
		},
		{
			name:           "returns FAILED status and releases seats when customer declines payment (no)",
			bookingID:      42,
			paymentSuccess: false,
			setup: func(cbRepo *mocks.MockCustomerBookingRepository) {
				cbRepo.On("GetBookingByID", ctx, 42).Return(pendingBooking, nil)
				cbRepo.On("CancelBooking", ctx, 42).Return(nil)
			},
			wantErr:    false,
			wantStatus: "FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cbRepo := &mocks.MockCustomerBookingRepository{}
			seatRepo := &mocks.MockSeatRepository{}
			tt.setup(cbRepo)

			svc := NewCustomerBookingService(cbRepo, seatRepo)
			resp, err := svc.ProcessPayment(ctx, tt.bookingID, tt.paymentSuccess)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.wantStatus, resp.Status)
				assert.Equal(t, tt.bookingID, resp.BookingID)
				assert.Equal(t, "INR", resp.Currency)
				assert.NotEmpty(t, resp.PaymentID)
			}

			cbRepo.AssertExpectations(t)
		})
	}
}

func TestGetBookingDetails(t *testing.T) {
	ctx := context.Background()

	confirmedBooking := &model.Booking{
		ID:            42,
		TransactionID: "SKY-abc12345",
		ShowID:        10,
		TotalPrice:    550.0,
		Status:        "CONFIRMED",
		Show: model.Show{
			Id:      10,
			MovieId: "tt1375666",
			Date:    time.Date(2025, time.March, 10, 18, 0, 0, 0, time.UTC),
			Screen:  model.Screen{ScreenName: "Screen 1"},
		},
		BookedSeats: []model.BookedSeat{
			{
				ID:               1,
				BookingID:        42,
				ShowSeatStatusID: 5,
				ShowSeatStatus: model.ShowSeatStatus{
					Seat: model.Seat{ID: 5, RowNumber: 1, ColumnNumber: 5, SeatType: "REGULAR"},
				},
			},
		},
	}

	cancelledBooking := &model.Booking{
		ID:     43,
		Status: "CANCELLED",
		Show:   model.Show{MovieId: "tt1375666", Date: time.Now(), Screen: model.Screen{ScreenName: "Screen 2"}},
	}

	tests := []struct {
		name         string
		bookingID    int
		setup        func(cbRepo *mocks.MockCustomerBookingRepository)
		wantErr      bool
		validateResp func(t *testing.T, resp interface{})
	}{
		{
			name:      "returns error when booking does not exist for the given booking ID",
			bookingID: 99,
			setup: func(cbRepo *mocks.MockCustomerBookingRepository) {
				cbRepo.On("GetBookingByID", ctx, 99).Return(nil, ae.NotFoundError("BookingNotFound", "booking not found", nil))
			},
			wantErr: true,
		},
		{
			name:      "returns booking details with payment status SUCCESS for a confirmed booking",
			bookingID: 42,
			setup: func(cbRepo *mocks.MockCustomerBookingRepository) {
				cbRepo.On("GetBookingByID", ctx, 42).Return(confirmedBooking, nil)
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp interface{}) {
				r := resp.(map[string]string)
				assert.Equal(t, "SUCCESS", r["paymentStatus"])
				assert.Equal(t, "CONFIRMED", r["bookingStatus"])
				assert.Equal(t, "tt1375666", r["movieId"])
				assert.Equal(t, "Screen 1", r["screenName"])
			},
		},
		{
			name:      "returns booking details with payment status FAILED for a cancelled booking",
			bookingID: 43,
			setup: func(cbRepo *mocks.MockCustomerBookingRepository) {
				cbRepo.On("GetBookingByID", ctx, 43).Return(cancelledBooking, nil)
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp interface{}) {
				r := resp.(map[string]string)
				assert.Equal(t, "FAILED", r["paymentStatus"])
				assert.Equal(t, "CANCELLED", r["bookingStatus"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cbRepo := &mocks.MockCustomerBookingRepository{}
			seatRepo := &mocks.MockSeatRepository{}
			tt.setup(cbRepo)

			svc := NewCustomerBookingService(cbRepo, seatRepo)
			resp, err := svc.GetBookingDetails(ctx, tt.bookingID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validateResp != nil {
					tt.validateResp(t, map[string]string{
						"paymentStatus": resp.PaymentStatus,
						"bookingStatus": resp.BookingStatus,
						"movieId":       resp.MovieID,
						"screenName":    resp.ScreenName,
					})
				}
			}

			cbRepo.AssertExpectations(t)
		})
	}
}

func TestPaymentStatusFromBookingStatus(t *testing.T) {
	tests := []struct {
		bookingStatus string
		wantPayment   string
	}{
		{"CONFIRMED", "SUCCESS"},
		{"CANCELLED", "FAILED"},
		{"PENDING", "PENDING"},
		{"UNKNOWN", "PENDING"},
	}

	for _, tt := range tests {
		t.Run("booking status "+tt.bookingStatus+" maps to payment status "+tt.wantPayment, func(t *testing.T) {
			assert.Equal(t, tt.wantPayment, paymentStatusFromBookingStatus(tt.bookingStatus))
		})
	}
}
