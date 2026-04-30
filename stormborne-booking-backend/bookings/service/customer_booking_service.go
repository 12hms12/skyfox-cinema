package service

import (
	"context"
	"fmt"
	"time"

	"skyfox/bookings/dto/response"
	"skyfox/bookings/repository"
	ae "skyfox/error"

	"github.com/google/uuid"
)

type CustomerBookingService interface {
	CreateBooking(ctx context.Context, customerID int64, showID int, seatStatusIDs []int) (*response.CreateCustomerBookingResponse, error)
	ProcessPayment(ctx context.Context, bookingID int, paymentSuccess bool) (*response.PaymentResponse, error)
	GetBookingDetails(ctx context.Context, bookingID int) (*response.BookingDetailsResponse, error)
	CheckIn(ctx context.Context, bookingID int) error
}

type customerBookingService struct {
	customerBookingRepo repository.CustomerBookingRepository
	seatRepo            repository.SeatRepository
}

func NewCustomerBookingService(customerBookingRepo repository.CustomerBookingRepository, seatRepo repository.SeatRepository) CustomerBookingService {
	return &customerBookingService{
		customerBookingRepo: customerBookingRepo,
		seatRepo:            seatRepo,
	}
}

func (s *customerBookingService) CreateBooking(ctx context.Context, customerID int64, showID int, seatStatusIDs []int) (*response.CreateCustomerBookingResponse, error) {
	show, err := s.seatRepo.GetShowByID(ctx, uint(showID))
	if err != nil {
		return nil, err
	}

	pricingList, err := s.seatRepo.GetPricingByShowID(ctx, uint(showID))
	if err != nil {
		return nil, err
	}

	pricingMap := make(map[string]float64)
	for _, p := range pricingList {
		pricingMap[p.SeatType] = p.Price
	}

	seatStatuses, err := s.customerBookingRepo.GetSeatStatusesWithSeatByIDs(ctx, seatStatusIDs)
	if err != nil {
		return nil, err
	}

	if len(seatStatuses) != len(seatStatusIDs) {
		return nil, ae.BadRequestError("InvalidSeats", "one or more seats not found for this show", nil)
	}

	const weekendSurcharge = 50.0
	weekend := isWeekend(show.Date)

	var seatInfos []response.BookingSeatInfo
	var totalPrice float64
	for _, ss := range seatStatuses {
		basePrice := pricingMap[ss.Seat.SeatType]
		if weekend {
			basePrice += weekendSurcharge
		}
		row := rowLabel(ss.Seat.RowNumber)
		label := fmt.Sprintf("%s%d", row, ss.Seat.ColumnNumber)
		seatInfos = append(seatInfos, response.BookingSeatInfo{
			SeatID:   ss.ID,
			Label:    label,
			SeatType: ss.Seat.SeatType,
			Price:    basePrice,
		})
		totalPrice += basePrice
	}

	bookingRef := fmt.Sprintf("SKY-%s", uuid.New().String()[:8])
	expiresAt := time.Now().Add(15 * time.Minute)

	booking, err := s.customerBookingRepo.CreateBookingWithSeatLock(ctx, customerID, showID, seatStatusIDs, totalPrice, expiresAt, bookingRef)
	if err != nil {
		return nil, err
	}

	return &response.CreateCustomerBookingResponse{
		BookingID:        booking.ID,
		BookingReference: booking.TransactionID,
		Seats:            seatInfos,
		TotalPrice:       totalPrice,
		Status:           booking.Status,
		ExpiresAt:        booking.ExpiresAt,
	}, nil
}

func (s *customerBookingService) ProcessPayment(ctx context.Context, bookingID int, paymentSuccess bool) (*response.PaymentResponse, error) {
	booking, err := s.customerBookingRepo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	paymentID := fmt.Sprintf("PAY-%s", uuid.New().String()[:8])

	if paymentSuccess {
		if confirmErr := s.customerBookingRepo.ConfirmBooking(ctx, bookingID); confirmErr != nil {
			return nil, confirmErr
		}
		return &response.PaymentResponse{
			PaymentID: paymentID,
			BookingID: bookingID,
			Amount:    booking.TotalPrice,
			Currency:  "INR",
			Status:    "SUCCESS",
		}, nil
	}

	if cancelErr := s.customerBookingRepo.CancelBooking(ctx, bookingID); cancelErr != nil {
		return nil, cancelErr
	}
	return &response.PaymentResponse{
		PaymentID: paymentID,
		BookingID: bookingID,
		Amount:    booking.TotalPrice,
		Currency:  "INR",
		Status:    "FAILED",
	}, nil
}

func (s *customerBookingService) GetBookingDetails(ctx context.Context, bookingID int) (*response.BookingDetailsResponse, error) {
	booking, err := s.customerBookingRepo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	var seats []response.BookingSeatInfo
	for _, bs := range booking.BookedSeats {
		seat := bs.ShowSeatStatus.Seat
		row := rowLabel(seat.RowNumber)
		label := fmt.Sprintf("%s%d", row, seat.ColumnNumber)
		seats = append(seats, response.BookingSeatInfo{
			SeatID:   bs.ShowSeatStatusID,
			Label:    label,
			SeatType: seat.SeatType,
		})
	}

	return &response.BookingDetailsResponse{
		BookingID:        booking.ID,
		BookingReference: booking.TransactionID,
		MovieID:          booking.Show.MovieId,
		ShowTime:         booking.Show.Date.Format("2006-01-02 15:04"),
		ScreenName:       booking.Show.Screen.ScreenName,
		Seats:            seats,
		TotalPrice:       booking.TotalPrice,
		PaymentStatus:    paymentStatusFromBookingStatus(booking.Status),
		BookingStatus:    booking.Status,
	}, nil
}

func paymentStatusFromBookingStatus(status string) string {
	switch status {
	case "CONFIRMED":
		return "SUCCESS"
	case "CANCELLED":
		return "FAILED"
	default:
		return "PENDING"
	}
}

func (s *customerBookingService) CheckIn(ctx context.Context, bookingID int) error {
	return s.customerBookingRepo.CheckInBooking(ctx, bookingID)
}
