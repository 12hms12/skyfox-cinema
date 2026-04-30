package service

import (
	"context"
	"skyfox/bookings/dto/response"
	"skyfox/bookings/repository"
)

type BookingRepository interface {
	BookingAmountByShows(context.Context, []int) float64
	BookingRevenueMap(ctx context.Context, shows []int) map[int]float64
}

type revenueService struct {
	showRepo ShowRepository
	bookRepo BookingRepository
}

func NewRevenueService(bookingRepository BookingRepository, showRepository ShowRepository) *revenueService {
	return &revenueService{
		showRepo: showRepository,
		bookRepo: bookingRepository,
	}
}

func (rs *revenueService) RevenueBy(ctx context.Context,revenueQuery *repository.RevenueQuery) (*response.RevenueResponse, error) {
	shows, err := rs.showRepo.GetAllShowsBy(ctx,revenueQuery)
	if err != nil {
		return &response.RevenueResponse{}, err
	}

	var showIds []int
	for _, show := range shows {
		showIds = append(showIds, show.Id)
	}

	revenue := rs.bookRepo.BookingRevenueMap(ctx, showIds)
	grossRevenue := 0.0
	var revenueShows []response.RevenueShow = make([]response.RevenueShow, 0)

	for _,v := range revenue{
		grossRevenue += v
	}

	for i := range shows{
		if revenue[shows[i].Id] != 0 {
			revenueShows = append(revenueShows, response.RevenueShow{
				Id: shows[i].Id,
				Date : shows[i].Date.Format("2006-01-02"),
				ShowTime: shows[i].Slot.StartTime,
				Movie: shows[i].Movie,
				Revenue: revenue[shows[i].Id],
				TicketSold: (uint)(revenue[shows[i].Id]/shows[i].Cost),
			})
		}
	}

	return &response.RevenueResponse{
		GrossRevenue: grossRevenue,
		Shows: revenueShows,
	}, nil
}