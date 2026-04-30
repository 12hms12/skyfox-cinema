package repository

import (
	"context"
	"fmt"
	"time"

	"skyfox/bookings/database/common"
	"skyfox/bookings/model"
	ae "skyfox/error"

	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CustomerBookingRepository interface {
	GetSeatStatusesWithSeatByIDs(ctx context.Context, ids []int) ([]model.ShowSeatStatus, error)
	CreateBookingWithSeatLock(ctx context.Context, onlineCustomerID int64, showID int, seatStatusIDs []int, totalPrice float64, expiresAt time.Time, bookingRef string) (*model.Booking, error)
	GetBookingByID(ctx context.Context, bookingID int) (*model.Booking, error)
	ConfirmBooking(ctx context.Context, bookingID int) error
	CancelBooking(ctx context.Context, bookingID int) error
	ReleaseExpiredBookings(ctx context.Context) (int64, error)
	CheckInBooking(ctx context.Context, bookingID int) error
}

type customerBookingRepository struct {
	*common.BaseDB
}

func NewCustomerBookingRepository(db *common.BaseDB) CustomerBookingRepository {
	return &customerBookingRepository{BaseDB: db}
}

func (repo *customerBookingRepository) GetSeatStatusesWithSeatByIDs(ctx context.Context, ids []int) ([]model.ShowSeatStatus, error) {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	var statuses []model.ShowSeatStatus
	result := dbCtx.Preload("Seat").Where("id IN ?", ids).Find(&statuses)
	if result.Error != nil {
		return nil, ae.InternalServerError("FetchSeatStatusFailed", "failed to fetch seat statuses", result.Error)
	}
	return statuses, nil
}

func (repo *customerBookingRepository) CreateBookingWithSeatLock(
	ctx context.Context,
	onlineCustomerID int64,
	showID int,
	seatStatusIDs []int,
	totalPrice float64,
	expiresAt time.Time,
	bookingRef string,
) (*model.Booking, error) {
	var createdBooking model.Booking

	err := repo.GormDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var statuses []model.ShowSeatStatus
		result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id IN ? AND show_id = ?", seatStatusIDs, showID).
			Find(&statuses)
		if result.Error != nil {
			return ae.InternalServerError("SeatLockFailed", "failed to lock seats", result.Error)
		}

		if len(statuses) != len(seatStatusIDs) {
			return ae.BadRequestError("InvalidSeats", "one or more seats not found for this show", nil)
		}

		for _, s := range statuses {
			if s.Status != "AVAILABLE" {
				return ae.BadRequestError("SeatsUnavailable", "one or more seats are no longer available", nil)
			}
		}

		if updateErr := tx.Model(&model.ShowSeatStatus{}).
			Where("id IN ?", seatStatusIDs).
			Updates(map[string]interface{}{
				"status":       "BLOCKED",
				"locked_until": expiresAt,
			}).Error; updateErr != nil {
			return ae.InternalServerError("SeatBlockFailed", "failed to block seats", updateErr)
		}

		booking := model.Booking{
			TransactionID:    bookingRef,
			OnlineCustomerID: uint(onlineCustomerID),
			ShowID:           showID,
			TotalPrice:       totalPrice,
			Status:           "PENDING",
			ExpiresAt:        expiresAt,
		}
		if createErr := tx.Create(&booking).Error; createErr != nil {
			return ae.InternalServerError("BookingCreationFailed", "failed to create booking", createErr)
		}

		for _, seatStatusID := range seatStatusIDs {
			bookedSeat := model.BookedSeat{
				BookingID:        booking.ID,
				ShowSeatStatusID: seatStatusID,
			}
			if createErr := tx.Create(&bookedSeat).Error; createErr != nil {
				return ae.InternalServerError("BookedSeatCreationFailed", "failed to link seat to booking", createErr)
			}
		}

		createdBooking = booking
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &createdBooking, nil
}

func (repo *customerBookingRepository) GetBookingByID(ctx context.Context, bookingID int) (*model.Booking, error) {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	var booking model.Booking
	result := dbCtx.
		Preload("Show.Slot").
		Preload("Show.Screen").
		Preload("BookedSeats.ShowSeatStatus.Seat").
		First(&booking, bookingID)
	if result.Error != nil {
		return nil, ae.NotFoundError("BookingNotFound", "booking not found", result.Error)
	}
	return &booking, nil
}

func (repo *customerBookingRepository) ConfirmBooking(ctx context.Context, bookingID int) error {
	return repo.GormDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var booking model.Booking
		if fetchErr := tx.Preload("BookedSeats").First(&booking, bookingID).Error; fetchErr != nil {
			return ae.NotFoundError("BookingNotFound", "booking not found", fetchErr)
		}

		if booking.Status != "PENDING" {
			return ae.BadRequestError("InvalidBookingStatus", fmt.Sprintf("cannot confirm a booking with status %s", booking.Status), nil)
		}

		seatStatusIDs := make([]int, len(booking.BookedSeats))
		for i, bs := range booking.BookedSeats {
			seatStatusIDs[i] = bs.ShowSeatStatusID
		}

		if updateErr := tx.Model(&model.ShowSeatStatus{}).
			Where("id IN ?", seatStatusIDs).
			Updates(map[string]interface{}{
				"status":       "BOOKED",
				"locked_until": nil,
			}).Error; updateErr != nil {
			return ae.InternalServerError("SeatConfirmFailed", "failed to confirm seats", updateErr)
		}

		if updateErr := tx.Model(&model.Booking{}).
			Where("id = ?", bookingID).
			Update("status", "CONFIRMED").Error; updateErr != nil {
			return ae.InternalServerError("BookingConfirmFailed", "failed to confirm booking", updateErr)
		}

		return nil
	})
}

func (repo *customerBookingRepository) CancelBooking(ctx context.Context, bookingID int) error {
	return repo.GormDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var booking model.Booking
		if fetchErr := tx.Preload("BookedSeats").First(&booking, bookingID).Error; fetchErr != nil {
			return ae.NotFoundError("BookingNotFound", "booking not found", fetchErr)
		}

		if booking.Status == "CANCELLED" {
			return nil
		}

		if booking.Status != "PENDING" {
			return ae.BadRequestError("InvalidBookingStatus", fmt.Sprintf("cannot cancel a booking with status %s", booking.Status), nil)
		}

		seatStatusIDs := make([]int, len(booking.BookedSeats))
		for i, bs := range booking.BookedSeats {
			seatStatusIDs[i] = bs.ShowSeatStatusID
		}

		if updateErr := tx.Model(&model.ShowSeatStatus{}).
			Where("id IN ?", seatStatusIDs).
			Updates(map[string]interface{}{
				"status":       "AVAILABLE",
				"locked_until": nil,
			}).Error; updateErr != nil {
			return ae.InternalServerError("SeatReleaseFailed", "failed to release seats", updateErr)
		}

		if updateErr := tx.Model(&model.Booking{}).
			Where("id = ?", bookingID).
			Update("status", "CANCELLED").Error; updateErr != nil {
			return ae.InternalServerError("BookingCancelFailed", "failed to cancel booking", updateErr)
		}

		return nil
	})
}

func (repo *customerBookingRepository) ReleaseExpiredBookings(ctx context.Context) (int64, error) {
	var released int64

	err := repo.GormDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		var expiredBookings []model.Booking
		if err := tx.Preload("BookedSeats").
			Where("status = ? AND expires_at < ?", "PENDING", time.Now()).
			Find(&expiredBookings).Error; err != nil {
			return ae.InternalServerError("FetchExpiredFailed", "failed to fetch expired bookings", err)
		}

		if len(expiredBookings) == 0 {
			return nil
		}

		var allSeatStatusIDs []int
		var allBookingIDs []int
		for _, b := range expiredBookings {
			allBookingIDs = append(allBookingIDs, b.ID)
			for _, bs := range b.BookedSeats {
				allSeatStatusIDs = append(allSeatStatusIDs, bs.ShowSeatStatusID)
			}
		}

		if err := tx.Model(&model.ShowSeatStatus{}).
			Where("id IN ?", allSeatStatusIDs).
			Updates(map[string]interface{}{
				"status":       "AVAILABLE",
				"locked_until": nil,
			}).Error; err != nil {
			return ae.InternalServerError("SeatReleaseFailed", "failed to release expired seats", err)
		}

		if err := tx.Model(&model.Booking{}).
			Where("id IN ?", allBookingIDs).
			Update("status", "CANCELLED").Error; err != nil {
			return ae.InternalServerError("BookingCancelFailed", "failed to cancel expired bookings", err)
		}

		released = int64(len(expiredBookings))
		return nil
	})

	return released, err
}

func (repo *customerBookingRepository) CheckInBooking(ctx context.Context, bookingID int) error {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	var booking model.Booking

	// Find booking
	if err := dbCtx.Where("id = ?", bookingID).First(&booking).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ae.NotFoundError("BookingNotFound", "booking not found", err)
		}
		return ae.InternalServerError("FetchFailed", "failed to fetch booking", err)
	}

	// Already checked in
	if booking.IsCheckedIn {
		return ae.BadRequestError("AlreadyCheckedIn", "booking already checked in", nil)
	}

	// Update checkedIn = true
	if err := dbCtx.Model(&booking).
		Update("is_checked_in", true).Error; err != nil {
		return ae.InternalServerError("UpdateFailed", "failed to update check-in status", err)
	}

	return nil
}
