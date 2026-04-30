package repository

import (
	"context"
	"skyfox/bookings/database/common"
	"skyfox/bookings/dto/request"
	"skyfox/common/logger"
)

type BookingViewRow struct {
	ID               int     `gorm:"column:id"`
	TransactionID    string  `gorm:"column:transaction_id"`
	OnlineCustomerID uint    `gorm:"column:online_customer_id"`
	ShowID           int     `gorm:"column:show_id"`
	MovieID          string  `gorm:"column:movie_id"`
	ShowDate         string  `gorm:"column:show_date"`
	StartTime        string  `gorm:"column:start_time"`
	TotalPrice       float64 `gorm:"column:total_price"`
	Status           string  `gorm:"column:status"`
}

type bookingViewRepository struct {
	*common.BaseDB
}

func NewBookingViewRepository(db *common.BaseDB) *bookingViewRepository {
	return &bookingViewRepository{
		BaseDB: db,
	}
}

func (repo *bookingViewRepository) GetAllBookings(ctx context.Context, filter request.AdminBookingViewRequest) ([]BookingViewRow, error) {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	query := dbCtx.Table("booking b").
		Select(`
			b.id,
			b.transaction_id,
			b.online_customer_id,
			b.show_id,
			s.movie_id      AS movie_id,
			s.date          AS show_date,
			sl.start_time   AS start_time,
			b.total_price,
			b.status
		`).
		Joins("JOIN show s ON s.id = b.show_id").
		Joins("JOIN slot sl ON sl.id = s.slot_id")

	if filter.StartDate != "" {
		query = query.Where("s.date >= ?", filter.StartDate)
	}
	if filter.EndDate != "" {
		query = query.Where("s.date <= ?", filter.EndDate)
	}

	var rows []BookingViewRow
	if result := query.Scan(&rows); result.Error != nil {
		logger.Error("error fetching bookings: %v", result.Error)
		return nil, result.Error
	}

	return rows, nil
}



func (repo *bookingViewRepository) GetAllBookingsCustomer(
	ctx context.Context,
	filter request.CustomerBookingViewRequest,
) ([]BookingViewRow, error) {

	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	query := dbCtx.Table("booking b").
		Select(`
			b.id,
			b.transaction_id,
			b.online_customer_id,
			b.show_id,
			s.movie_id      AS movie_id,
			s.date          AS show_date,
			sl.start_time   AS start_time,
			b.total_price,
			b.status
		`).
		Joins("JOIN show s ON s.id = b.show_id").
		Joins("JOIN slot sl ON sl.id = s.slot_id")

	// Filter by customer
	if filter.OnlineCustomerID != 0 {
		query = query.Where("b.online_customer_id = ?", filter.OnlineCustomerID)
	}

	// Date filters
	if filter.StartDate != "" {
		query = query.Where("s.date >= ?", filter.StartDate)
	}

	if filter.EndDate != "" {
		query = query.Where("s.date <= ?", filter.EndDate)
	}

	var rows []BookingViewRow
	if result := query.Scan(&rows); result.Error != nil {
		logger.Error("error fetching bookings: %v", result.Error)
		return nil, result.Error
	}

	return rows, nil
}
