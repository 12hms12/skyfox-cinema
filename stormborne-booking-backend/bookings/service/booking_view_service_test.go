package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"skyfox/bookings/dto/request"
	"skyfox/bookings/repository"
	"skyfox/bookings/service/mocks"
	ae "skyfox/error"
	movieServiceMock "skyfox/movieservice/movie_gateway/mocks"
	"skyfox/bookings/model"
)

func TestBookingViewService(t *testing.T) {
	mockRows := []repository.BookingViewRow{
		{
			ID:               1,
			TransactionID:    "txn-001",
			OnlineCustomerID: 10,
			ShowID:           5,
			MovieID:          "tt1234567",
			ShowDate:         "2026-03-17",
			StartTime:        "18:00:00",
			TotalPrice:       500.00,
			Status:           "CONFIRMED",
		},
	}

	mockMovie := &model.Movie{
		MovieId: "tt1234567",
		Name:    "Inception",
	}

	tests := []struct {
		name           string
		filter         request.AdminBookingViewRequest
		repoReturn     []repository.BookingViewRow
		repoErr        error
		movieReturn    *model.Movie
		movieErr       error
		expectedTotal  int
		expectedErr    bool
	}{
		{
			name:          "returns all bookings with no filters",
			filter:        request.AdminBookingViewRequest{},
			repoReturn:    mockRows,
			repoErr:       nil,
			movieReturn:   mockMovie,
			movieErr:      nil,
			expectedTotal: 1,
			expectedErr:   false,
		},
		{
			name:          "filters by movie name correctly",
			filter:        request.AdminBookingViewRequest{MovieName: "Inception"},
			repoReturn:    mockRows,
			repoErr:       nil,
			movieReturn:   mockMovie,
			movieErr:      nil,
			expectedTotal: 1,
			expectedErr:   false,
		},
		{
			name:          "excludes booking when movie name does not match",
			filter:        request.AdminBookingViewRequest{MovieName: "Avatar"},
			repoReturn:    mockRows,
			repoErr:       nil,
			movieReturn:   mockMovie,
			movieErr:      nil,
			expectedTotal: 0,
			expectedErr:   false,
		},
		{
			name:          "skips booking when movie gateway returns error",
			filter:        request.AdminBookingViewRequest{},
			repoReturn:    mockRows,
			repoErr:       nil,
			movieReturn:   nil,
			movieErr:      ae.InternalServerError("MovieError", "movie not found", nil),
			expectedTotal: 0,
			expectedErr:   false,
		},
		{
			name:          "returns error when repository fails",
			filter:        request.AdminBookingViewRequest{},
			repoReturn:    nil,
			repoErr:       ae.InternalServerError("DBError", "db failed", nil),
			movieReturn:   nil,
			movieErr:      nil,
			expectedTotal: 0,
			expectedErr:   true,
		},
		{
			name:          "returns empty list when no bookings in db",
			filter:        request.AdminBookingViewRequest{},
			repoReturn:    []repository.BookingViewRow{},
			repoErr:       nil,
			movieReturn:   nil,
			movieErr:      nil,
			expectedTotal: 0,
			expectedErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockBookingViewRepository(t)
			mockGateway := movieServiceMock.MockMovieGateWay{}

			mockRepo.EXPECT().
				GetAllBookings(mock.Anything, tt.filter).
				Return(tt.repoReturn, tt.repoErr).
				Once()

			if tt.repoErr == nil && len(tt.repoReturn) > 0 {
				mockGateway.On("MovieById", mock.Anything, mockRows[0].MovieID).
					Return(tt.movieReturn, tt.movieErr).Once()
			}

			svc := NewBookingViewService(mockRepo, &mockGateway)
			result, err := svc.GetAllBookings(context.Background(), tt.filter)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedTotal, result.Total)
				if tt.expectedTotal > 0 {
					assert.Equal(t, mockMovie.Name, result.Bookings[0].MovieName)
				}
			}

			mockGateway.AssertExpectations(t)
		})
	}
}

func TestGetAllBookingsCustomer(t *testing.T) {
	mockRows := []repository.BookingViewRow{
		{
			ID:               1,
			TransactionID:    "txn-001",
			OnlineCustomerID: 10,
			ShowID:           5,
			MovieID:          "tt1234567",
			ShowDate:         "2026-03-17",
			StartTime:        "18:00:00",
			TotalPrice:       500.00,
			Status:           "CONFIRMED",
		},
	}

	mockMovie := &model.Movie{
		MovieId: "tt1234567",
		Name:    "Inception",
	}

	tests := []struct {
		name           string
		filter         request.CustomerBookingViewRequest
		repoReturn     []repository.BookingViewRow
		repoErr        error
		movieReturn    *model.Movie
		movieErr       error
		expectedTotal  int
		expectedErr    bool
	}{
		{
			name:          "returns customer bookings successfully",
			filter:        request.CustomerBookingViewRequest{},
			repoReturn:    mockRows,
			repoErr:       nil,
			movieReturn:   mockMovie,
			movieErr:      nil,
			expectedTotal: 1,
			expectedErr:   false,
		},
		{
			name:          "filters by movie name correctly",
			filter:        request.CustomerBookingViewRequest{MovieName: "Inception"},
			repoReturn:    mockRows,
			repoErr:       nil,
			movieReturn:   mockMovie,
			movieErr:      nil,
			expectedTotal: 1,
			expectedErr:   false,
		},
		{
			name:          "excludes booking when movie name does not match",
			filter:        request.CustomerBookingViewRequest{MovieName: "Avatar"},
			repoReturn:    mockRows,
			repoErr:       nil,
			movieReturn:   mockMovie,
			movieErr:      nil,
			expectedTotal: 0,
			expectedErr:   false,
		},
		{
			name:          "skips booking when movie gateway fails",
			filter:        request.CustomerBookingViewRequest{},
			repoReturn:    mockRows,
			repoErr:       nil,
			movieReturn:   nil,
			movieErr:      ae.InternalServerError("MovieError", "movie not found", nil),
			expectedTotal: 0,
			expectedErr:   false,
		},
		{
			name:          "returns error when repository fails",
			filter:        request.CustomerBookingViewRequest{},
			repoReturn:    nil,
			repoErr:       ae.InternalServerError("DBError", "db error", nil),
			movieReturn:   nil,
			movieErr:      nil,
			expectedTotal: 0,
			expectedErr:   true,
		},
		{
			name:          "returns empty when no bookings found",
			filter:        request.CustomerBookingViewRequest{},
			repoReturn:    []repository.BookingViewRow{},
			repoErr:       nil,
			movieReturn:   nil,
			movieErr:      nil,
			expectedTotal: 0,
			expectedErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockBookingViewRepository(t)
			mockGateway := movieServiceMock.MockMovieGateWay{}

			mockRepo.EXPECT().
				GetAllBookingsCustomer(mock.Anything, tt.filter).
				Return(tt.repoReturn, tt.repoErr).
				Once()

			if tt.repoErr == nil && len(tt.repoReturn) > 0 {
				mockGateway.
					On("MovieById", mock.Anything, mockRows[0].MovieID).
					Return(tt.movieReturn, tt.movieErr).
					Once()
			}

			svc := NewBookingViewService(mockRepo, &mockGateway)
			result, err := svc.GetAllBookingsCustomer(context.Background(), tt.filter)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedTotal, result.Total)

				if tt.expectedTotal > 0 {
					assert.Equal(t, mockMovie.Name, result.Bookings[0].MovieName)
				}
			}

			mockGateway.AssertExpectations(t)
		})
	}
}