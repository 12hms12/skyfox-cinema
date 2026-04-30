package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"skyfox/bookings/dto/response"
	"skyfox/bookings/model"
	"skyfox/bookings/repository/mocks"

	"github.com/stretchr/testify/assert"
)



var (
	weekdayShow = &model.Show{Id: 1, Date: time.Date(2025, time.March, 10, 18, 0, 0, 0, time.UTC)}
	weekendShow = &model.Show{Id: 2, Date: time.Date(2025, time.March, 15, 18, 0, 0, 0, time.UTC)}

	mockSeatStatuses = []model.ShowSeatStatus{
		{ID: 1, ShowID: 1, SeatID: 1, Status: "AVAILABLE", Seat: model.Seat{ID: 1, RowNumber: 1, ColumnNumber: 1, SeatType: "REGULAR"}},
		{ID: 2, ShowID: 1, SeatID: 2, Status: "BOOKED", Seat: model.Seat{ID: 2, RowNumber: 1, ColumnNumber: 2, SeatType: "REGULAR"}},
	}

	mockPricing = []model.ShowPricing{
		{ID: 1, ShowID: 1, SeatType: "REGULAR", Price: 200.0},
	}
)

func TestGetSeatStatus(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		showID   uint
		setup    func(repo *mocks.MockSeatRepository)
		wantErr  bool
		validate func(t *testing.T, resp *response.SeatStatusResponse)
	}{
		{
			name:   "should return error when GetShowByID fails",
			showID: 1,
			setup: func(repo *mocks.MockSeatRepository) {
				repo.On("GetShowByID", ctx, uint(1)).Return(nil, errors.New("show not found"))
			},
			wantErr: true,
		},
		{
			name:   "should return error when GetSeatStatusByShowID fails",
			showID: 1,
			setup: func(repo *mocks.MockSeatRepository) {
				repo.On("GetShowByID", ctx, uint(1)).Return(weekdayShow, nil)
				repo.On("GetSeatStatusByShowID", ctx, uint(1)).Return([]model.ShowSeatStatus{}, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name:   "should return error when no seats found for show",
			showID: 1,
			setup: func(repo *mocks.MockSeatRepository) {
				repo.On("GetShowByID", ctx, uint(1)).Return(weekdayShow, nil)
				repo.On("GetSeatStatusByShowID", ctx, uint(1)).Return([]model.ShowSeatStatus{}, nil)
			},
			wantErr: true,
		},
		{
			name:   "should return error when GetPricingByShowID fails",
			showID: 1,
			setup: func(repo *mocks.MockSeatRepository) {
				repo.On("GetShowByID", ctx, uint(1)).Return(weekdayShow, nil)
				repo.On("GetSeatStatusByShowID", ctx, uint(1)).Return(mockSeatStatuses, nil)
				repo.On("GetPricingByShowID", ctx, uint(1)).Return([]model.ShowPricing{}, errors.New("pricing error"))
			},
			wantErr: true,
		},
		{
			name:   "should return correct seat status with available and sold seats on a weekday show",
			showID: 1,
			setup: func(repo *mocks.MockSeatRepository) {
				repo.On("GetShowByID", ctx, uint(1)).Return(weekdayShow, nil)
				repo.On("GetSeatStatusByShowID", ctx, uint(1)).Return(mockSeatStatuses, nil)
				repo.On("GetPricingByShowID", ctx, uint(1)).Return(mockPricing, nil)
			},
			wantErr: false,
			validate: func(t *testing.T, resp *response.SeatStatusResponse) {
				assert.Equal(t, uint(1), resp.ShowID)
				assert.Len(t, resp.Seats, 2)
				assert.Equal(t, "A1", resp.Seats[0].Label)
				assert.Equal(t, "available", resp.Seats[0].Status)
				assert.Equal(t, "A2", resp.Seats[1].Label)
				assert.Equal(t, "sold", resp.Seats[1].Status)
				assert.Equal(t, 200.0, resp.Seats[0].BasePrice)
				assert.False(t, resp.Seats[0].IsWeekend)
				assert.Equal(t, "REGULAR", resp.Seats[0].SeatType)
				assert.Equal(t, "A", resp.Seats[0].Row)
				assert.Equal(t, 1, resp.Seats[0].Column)
			},
		},
		{
			name:   "should apply weekend surcharge of 50 when show date falls on Saturday or Sunday",
			showID: 2,
			setup: func(repo *mocks.MockSeatRepository) {
				repo.On("GetShowByID", ctx, uint(2)).Return(weekendShow, nil)
				repo.On("GetSeatStatusByShowID", ctx, uint(2)).Return(mockSeatStatuses, nil)
				repo.On("GetPricingByShowID", ctx, uint(2)).Return(mockPricing, nil)
			},
			wantErr: false,
			validate: func(t *testing.T, resp *response.SeatStatusResponse) {
				assert.Equal(t, 250.0, resp.Seats[0].BasePrice)
				assert.True(t, resp.Seats[0].IsWeekend)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockSeatRepository{}
			tt.setup(repo)

			svc := NewSeatService(repo)
			resp, err := svc.GetSeatStatus(ctx, tt.showID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validate != nil {
					tt.validate(t, resp)
				}
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestIsWeekend(t *testing.T) {
	tests := []struct {
		name string
		date time.Time
		want bool
	}{
		{"Saturday is weekend", time.Date(2025, time.March, 15, 0, 0, 0, 0, time.UTC), true},
		{"Sunday is weekend", time.Date(2025, time.March, 16, 0, 0, 0, 0, time.UTC), true},
		{"Monday is not weekend", time.Date(2025, time.March, 10, 0, 0, 0, 0, time.UTC), false},
		{"Friday is not weekend", time.Date(2025, time.March, 14, 0, 0, 0, 0, time.UTC), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isWeekend(tt.date))
		})
	}
}

func TestRowLabel(t *testing.T) {
	tests := []struct {
		rowNumber int
		want      string
	}{
		{1, "A"},
		{2, "B"},
		{10, "J"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, rowLabel(tt.rowNumber))
		})
	}
}