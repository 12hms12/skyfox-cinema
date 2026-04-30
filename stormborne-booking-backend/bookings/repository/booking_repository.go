package repository

import (
	"context"
	"database/sql"
	"skyfox/bookings/database/common"
	"skyfox/bookings/model"
)

type bookingRepository struct {
	*common.BaseDB
}

func NewBookingRepository(db *common.BaseDB) *bookingRepository {
	return &bookingRepository{
		BaseDB: db,
	}
}

func (repo *bookingRepository) BookingAmountByShows(ctx context.Context, shows []int) float64 {
	var amount sql.NullFloat64

	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	dbCtx.Table(model.Booking{}.TableName()).Select("sum(total_price)").Where("show_id IN ? AND status = ?", shows, "CONFIRMED").Scan(&amount)

	return amount.Float64
}

type showRevenueRow struct {
	ShowID int     `gorm:"column:show_id"`
	Total  float64 `gorm:"column:total"`
}

func (repo *bookingRepository) BookingRevenueMap(ctx context.Context, shows []int) map[int]float64 {
	revenueMap := make(map[int]float64)

	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	var onlineResults []showRevenueRow
	dbCtx.Table(model.Booking{}.TableName()).
		Select("show_id, sum(total_price) as total").
		Where("show_id IN ? AND status = ?", shows, "CONFIRMED").
		Group("show_id").
		Scan(&onlineResults)

	for _, res := range onlineResults {
		revenueMap[res.ShowID] += res.Total
	}

	return revenueMap
}
