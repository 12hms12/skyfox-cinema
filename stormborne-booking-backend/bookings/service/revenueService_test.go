package service

import (
	"context"
	"errors"
	"skyfox/bookings/dto/response"
	"skyfox/bookings/model"
	"skyfox/bookings/repository"
	"skyfox/bookings/service/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRevenueService(t *testing.T) {

	t.Run("RevenueByDate when shows exists", func(t *testing.T) {
		bookrepo := mocks.NewMockBookingRepository(t)
		showrepo := mocks.NewMockShowRepository(t)
		shows := []model.Show{ 
			{ Id: 0, Cost: 100, Date: time.Date(2022,time.October,14,0,0,0,0,time.UTC), },
			{ Id: 1, Cost: 100, Date: time.Date(2022,time.October,14,0,0,0,0,time.UTC), },
		}
		expected := &response.RevenueResponse{
			GrossRevenue: 300,
			Shows: []response.RevenueShow{
				{
					Id:      0,
					Revenue: 100.0,
					TicketSold: 1,
					Date: "2022-10-14",
				},
				{
					Id:      1,
					Revenue: 200.0,
					TicketSold: 2,
					Date: "2022-10-14",
				},
			},
		}
		showrepo.On("GetAllShowsBy", mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("*repository.RevenueQuery")).Return(shows, nil).Once()
		bookrepo.On("BookingRevenueMap", mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("[]int")).Return(map[int]float64{0: 100.0, 1: 200.0}).Once()

		service := NewRevenueService(bookrepo, showrepo)

		got, err := service.RevenueBy(context.Background(), &repository.RevenueQuery{})

		assert.Nil(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("RevenueByDate when no shows exist", func(t *testing.T) {
		bookrepo := mocks.MockBookingRepository{}
		showrepo := mocks.MockShowRepository{}
		shows := make([]model.Show, 0)
		expected := &response.RevenueResponse{
			GrossRevenue: 0,
			Shows: make([]response.RevenueShow, 0),
		}
		showrepo.On("GetAllShowsBy", mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("*repository.RevenueQuery")).Return(shows, nil).Once()
		bookrepo.On("BookingRevenueMap", mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("[]int")).Return(map[int]float64{0: 0.0}).Once()

		service := NewRevenueService(&bookrepo, &showrepo)

		got, err := service.RevenueBy(context.Background(), &repository.RevenueQuery{})

		assert.Nil(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("RevenueByMovieId when shows exists", func(t *testing.T) {
		bookrepo := mocks.NewMockBookingRepository(t)
		showrepo := mocks.NewMockShowRepository(t)
		shows := []model.Show{ 
			{ Id: 0, Cost: 100, Date: time.Date(2022,time.October,14,0,0,0,0,time.UTC), },
			{ Id: 1, Cost: 100, Date: time.Date(2022,time.October,14,0,0,0,0,time.UTC), },
		}
		expected := &response.RevenueResponse{
			GrossRevenue: 300,
			Shows: []response.RevenueShow{
				{
					Id:      0,
					Revenue: 100.0,
					TicketSold: 1,
					Date: "2022-10-14",
				},
				{
					Id:      1,
					Revenue: 200.0,
					TicketSold: 2,
					Date: "2022-10-14",
				},
			},
		}
		showrepo.On("GetAllShowsBy", mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("*repository.RevenueQuery")).Return(shows, nil).Once()
		bookrepo.On("BookingRevenueMap", mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("[]int")).Return(map[int]float64{0: 100.0, 1: 200.0}).Once()

		service := NewRevenueService(bookrepo, showrepo)

		got, err := service.RevenueBy(context.Background(), &repository.RevenueQuery{})

		assert.Nil(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("RevenueByMovieId when no shows exist", func(t *testing.T) {
		bookrepo := mocks.NewMockBookingRepository(t)
		showrepo := mocks.NewMockShowRepository(t)
		expected := &response.RevenueResponse{
			GrossRevenue: 0,
			Shows: nil,
		}
		showrepo.On("GetAllShowsBy", mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("*repository.RevenueQuery")).Return(nil, errors.New("a mock error")).Once()
		
		service := NewRevenueService(bookrepo, showrepo)
		got, err := service.RevenueBy(context.Background(), &repository.RevenueQuery{})

		assert.Error(t, err)
		assert.Equal(t, expected, got)
	})
}
