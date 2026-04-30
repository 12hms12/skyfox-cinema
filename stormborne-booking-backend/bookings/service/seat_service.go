	package service

import (
	"context"
	"fmt"
	"time"

	"skyfox/bookings/dto/response"
	"skyfox/bookings/repository"
	ae "skyfox/error"
)

type SeatService interface {
	GetSeatStatus(ctx context.Context, showID uint) (*response.SeatStatusResponse, error)
}

type seatService struct {
	seatRepo repository.SeatRepository
}

func NewSeatService(seatRepo repository.SeatRepository) SeatService {
	return &seatService{seatRepo: seatRepo}
}

func (s *seatService) GetSeatStatus(ctx context.Context, showID uint) (*response.SeatStatusResponse, error) {
	show, err := s.seatRepo.GetShowByID(ctx, showID)
	if err != nil {
		return nil, err
	}

	seatStatuses, err := s.seatRepo.GetSeatStatusByShowID(ctx, showID)
	if err != nil {
		return nil, err
	}

	if len(seatStatuses) == 0 {
		return nil, ae.UnProcessableError("NoSeatsFound", "No seats found for the given show", nil)
	}

	pricingList, err := s.seatRepo.GetPricingByShowID(ctx, showID)
	if err != nil {
		return nil, err
	}

	pricingMap := make(map[string]float64)
	for _, p := range pricingList {
		pricingMap[p.SeatType] = p.Price
	}

	weekend := isWeekend(show.Date)
	const weekendSurcharge = 50.0

	var seats []response.SeatResponse
	for _, ss := range seatStatuses {
		seat := ss.Seat
		row := rowLabel(seat.RowNumber)
		label := fmt.Sprintf("%s%d", row, seat.ColumnNumber)

		status := "available"
		if ss.Status == "BOOKED" {
			status = "sold"
		} else if ss.Status == "BLOCKED" && ss.LockedUntil != nil && ss.LockedUntil.After(time.Now()) {
			status = "sold"
		}

		basePrice := pricingMap[seat.SeatType]
		if weekend {
			basePrice += weekendSurcharge
		}

		seats = append(seats, response.SeatResponse{
			SeatID:    uint(ss.ID),
			Label:     label,
			Row:       row,
			Column:    seat.ColumnNumber,
			SeatType:  seat.SeatType,
			Status:    status,
			BasePrice: basePrice,
			IsWeekend: weekend,
		})
	}

	return &response.SeatStatusResponse{
		ShowID:   showID,
		ShowDate: show.Date.Format("2006-01-02"),
		Seats:    seats,
	}, nil
}

func isWeekend(t time.Time) bool {
	day := t.Weekday()
	return day == time.Saturday || day == time.Sunday
}

func rowLabel(rowNumber int) string {
	return string(rune('A' + rowNumber - 1))
}