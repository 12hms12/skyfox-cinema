package service

import (
	"context"
	"skyfox/bookings/dto/request"
	"skyfox/bookings/dto/response"
	"skyfox/bookings/repository"
	"skyfox/common/logger"
	movieservice "skyfox/movieservice/movie_gateway"
	"strings"

)

type BookingViewRepository interface {
	GetAllBookings(ctx context.Context, filter request.AdminBookingViewRequest) ([]repository.BookingViewRow, error)
	GetAllBookingsCustomer(
		ctx context.Context,
		filter request.CustomerBookingViewRequest,
	) ([]repository.BookingViewRow, error)
}

type bookingViewService struct {
	bookingViewRepo BookingViewRepository
	movieGateway    movieservice.MovieGateWay
}

func NewBookingViewService(bookingViewRepo BookingViewRepository, movieGateway movieservice.MovieGateWay) *bookingViewService {
	return &bookingViewService{
		bookingViewRepo: bookingViewRepo,
		movieGateway:    movieGateway,
	}
}

func (s *bookingViewService) GetAllBookings(ctx context.Context, filter request.AdminBookingViewRequest) (*response.AdminBookingViewResponse, error) {
	rows, err := s.bookingViewRepo.GetAllBookings(ctx, filter)
	if err != nil {
		logger.Error("error fetching bookings: %v", err)
		return nil, err
	}

	items := make([]response.AdminBookingViewItem, 0)
	for _, row := range rows {
		movie, err := s.movieGateway.MovieById(ctx, row.MovieID)
		if err != nil || movie == nil {
			logger.Error("error fetching movie for id %s: %v", row.MovieID, err)
			continue
		}

		if filter.MovieName != "" && !strings.Contains(strings.ToLower(movie.Name), strings.ToLower(filter.MovieName)) {
			continue
		}

		items = append(items, response.AdminBookingViewItem{
			ID:               row.ID,
			TransactionID:    row.TransactionID,
			OnlineCustomerID: row.OnlineCustomerID,
			ShowID:           row.ShowID,
			MovieName:        movie.Name,
			ShowDate:         row.ShowDate,
			StartTime:        row.StartTime,
			TotalPrice:       row.TotalPrice,
			Status:           row.Status,
		})
	}

	return response.NewAdminBookingViewResponse(items), nil
}

func (s *bookingViewService) GetAllBookingsCustomer(
	ctx context.Context,
	filter request.CustomerBookingViewRequest,
) (*response.AdminBookingViewResponse, error) {

	rows, err := s.bookingViewRepo.GetAllBookingsCustomer(ctx, filter)
	if err != nil {
		logger.Error("error fetching bookings: %v", err)
		return nil, err
	}

	items := make([]response.AdminBookingViewItem, 0, len(rows))

	for _, row := range rows {
		movie, err := s.movieGateway.MovieById(ctx, row.MovieID)
		if err != nil || movie == nil {
			logger.Error("error fetching movie for id %s: %v", row.MovieID, err)
			continue
		}

		// Movie name filter
		if filter.MovieName != "" &&
			!strings.Contains(strings.ToLower(movie.Name), strings.ToLower(filter.MovieName)) {
			continue
		}

		items = append(items, response.AdminBookingViewItem{
			ID:               row.ID,
			TransactionID:    row.TransactionID,
			OnlineCustomerID: row.OnlineCustomerID,
			ShowID:           row.ShowID,
			MovieName:        movie.Name,
			ShowDate:         row.ShowDate,
			StartTime:        row.StartTime,
			TotalPrice:       row.TotalPrice,
			Status:           row.Status,
		})
	}

	return response.NewAdminBookingViewResponse(items), nil
}