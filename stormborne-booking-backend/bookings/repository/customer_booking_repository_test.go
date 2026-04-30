package repository_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"skyfox/bookings/model"
	"skyfox/bookings/repository"
	"skyfox/integration_test/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type customerBookingFixtures struct {
	customerID   int64
	showID       int
	seatStatusID int
}

func setupCustomerBookingDB(t *testing.T) (repository.CustomerBookingRepository, *customerBookingFixtures) {
	t.Helper()

	database := db.GetDB()
	database.GormDB().AutoMigrate(
		&model.Customer{},
		&model.Slot{},
		&model.Screen{},
		&model.Seat{},
		&model.Show{},
		&model.ShowSeatStatus{},
		&model.Booking{},
		&model.BookedSeat{},
	)

	database.GormDB().Exec("TRUNCATE booked_seat CASCADE")
	database.GormDB().Exec("TRUNCATE booking CASCADE")
	database.GormDB().Exec("TRUNCATE show_seat_status CASCADE")
	database.GormDB().Exec("TRUNCATE show CASCADE")
	database.GormDB().Exec("TRUNCATE seat CASCADE")
	database.GormDB().Exec("TRUNCATE screen CASCADE")
	database.GormDB().Exec("TRUNCATE slot CASCADE")

	customer := &model.Customer{
		FirstName:   "Jane",
		LastName:    "Doe",
		Email:       "jane.doe@test.com",
		PhoneNumber: "9876543210",
		CountryCode: "+91",
		Age:         25,
		Gender:      model.FEMALE,
		Username:    "janedoe",
		Password:    "hashed_password",
	}
	database.GormDB().Create(customer)

	t.Cleanup(func() {
		database.GormDB().Exec("DELETE FROM booked_seat")
		database.GormDB().Exec("DELETE FROM booking")
		database.GormDB().Exec("DELETE FROM customer WHERE id = ?", customer.ID)
	})

	slot := &model.Slot{Name: "Evening", StartTime: "18:00", EndTime: "21:00"}
	database.GormDB().Create(slot)

	screen := &model.Screen{ScreenName: "Screen A"}
	database.GormDB().Create(screen)

	seat := &model.Seat{ScreenID: screen.ID, RowNumber: 5, ColumnNumber: 3, SeatType: "REGULAR"}
	database.GormDB().Create(seat)

	show := &model.Show{
		ScreenID: screen.ID,
		MovieId:  "tt1234567",
		Date:     time.Now().AddDate(0, 0, 1),
		Cost:     250.0,
		SlotId:   slot.Id,
	}
	database.GormDB().Create(show)

	seatStatus := &model.ShowSeatStatus{ShowID: show.Id, SeatID: seat.ID, Status: "AVAILABLE"}
	database.GormDB().Create(seatStatus)

	return repository.NewCustomerBookingRepository(database), &customerBookingFixtures{
		customerID:   int64(customer.ID),
		showID:       show.Id,
		seatStatusID: seatStatus.ID,
	}
}

func TestCustomerBookingRepository_GetSeatStatusesWithSeatByIDs(t *testing.T) {
	repo, fx := setupCustomerBookingDB(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		ids         []int
		wantLen     int
		wantPreload bool
	}{
		{
			name:        "returns seat statuses with Seat preloaded for valid IDs",
			ids:         []int{fx.seatStatusID},
			wantLen:     1,
			wantPreload: true,
		},
		{
			name:        "returns empty slice for non-existent IDs",
			ids:         []int{99999},
			wantLen:     0,
			wantPreload: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			statuses, err := repo.GetSeatStatusesWithSeatByIDs(ctx, tc.ids)
			assert.NoError(t, err)
			assert.Len(t, statuses, tc.wantLen)
			if tc.wantPreload {
				assert.NotZero(t, statuses[0].Seat.ID)
			}
		})
	}
}

func TestCustomerBookingRepository_CreateBookingWithSeatLock(t *testing.T) {
	repo, fx := setupCustomerBookingDB(t)
	ctx := context.Background()
	expiresAt := time.Now().Add(15 * time.Minute)

	t.Run("successfully creates booking and blocks seats when all seats are AVAILABLE", func(t *testing.T) {
		booking, err := repo.CreateBookingWithSeatLock(ctx, fx.customerID, fx.showID, []int{fx.seatStatusID}, 250.0, expiresAt, "REF-001")

		require.NoError(t, err)
		require.NotNil(t, booking)
		assert.NotZero(t, booking.ID)
		assert.Equal(t, "PENDING", booking.Status)
		assert.Equal(t, "REF-001", booking.TransactionID)
		assert.Equal(t, 250.0, booking.TotalPrice)

		var updated model.ShowSeatStatus
		db.GetDB().GormDB().First(&updated, fx.seatStatusID)
		assert.Equal(t, "BLOCKED", updated.Status)
		assert.NotNil(t, updated.LockedUntil)

		var bookedSeats []model.BookedSeat
		db.GetDB().GormDB().Where("booking_id = ?", booking.ID).Find(&bookedSeats)
		assert.Len(t, bookedSeats, 1)
		assert.Equal(t, fx.seatStatusID, bookedSeats[0].ShowSeatStatusID)
	})

	t.Run("returns error when seat ID does not belong to the given show", func(t *testing.T) {
		booking, err := repo.CreateBookingWithSeatLock(ctx, fx.customerID, fx.showID, []int{99999}, 250.0, expiresAt, "REF-INVALID")

		assert.Error(t, err)
		assert.Nil(t, booking)
	})

	t.Run("returns error when seat is not AVAILABLE because it was already blocked", func(t *testing.T) {
		db.GetDB().GormDB().Model(&model.ShowSeatStatus{}).Where("id = ?", fx.seatStatusID).Update("status", "BLOCKED")
		defer db.GetDB().GormDB().Model(&model.ShowSeatStatus{}).Where("id = ?", fx.seatStatusID).Update("status", "AVAILABLE")

		booking, err := repo.CreateBookingWithSeatLock(ctx, fx.customerID, fx.showID, []int{fx.seatStatusID}, 250.0, expiresAt, "REF-BLOCKED")

		assert.Error(t, err)
		assert.Nil(t, booking)
	})

	t.Run("returns error when seat is not AVAILABLE because it is already BOOKED", func(t *testing.T) {
		db.GetDB().GormDB().Model(&model.ShowSeatStatus{}).Where("id = ?", fx.seatStatusID).Update("status", "BOOKED")
		defer db.GetDB().GormDB().Model(&model.ShowSeatStatus{}).Where("id = ?", fx.seatStatusID).Update("status", "AVAILABLE")

		booking, err := repo.CreateBookingWithSeatLock(ctx, fx.customerID, fx.showID, []int{fx.seatStatusID}, 250.0, expiresAt, "REF-BOOKED")

		assert.Error(t, err)
		assert.Nil(t, booking)
	})
}

func TestCustomerBookingRepository_CreateBookingWithSeatLock_Concurrency(t *testing.T) {
	repo, fx := setupCustomerBookingDB(t)
	ctx := context.Background()
	expiresAt := time.Now().Add(15 * time.Minute)

	db.GetDB().GormDB().Model(&model.ShowSeatStatus{}).Where("id = ?", fx.seatStatusID).Update("status", "AVAILABLE")

	var wg sync.WaitGroup
	results := make([]error, 2)

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, err := repo.CreateBookingWithSeatLock(ctx, fx.customerID, fx.showID, []int{fx.seatStatusID}, 250.0, expiresAt, "REF-CONC")
			results[idx] = err
		}(i)
	}

	wg.Wait()

	successCount := 0
	errorCount := 0
	for _, err := range results {
		if err == nil {
			successCount++
		} else {
			errorCount++
		}
	}

	assert.Equal(t, 1, successCount, "exactly one concurrent booking attempt should succeed")
	assert.Equal(t, 1, errorCount, "exactly one concurrent booking attempt should fail with seat unavailable")
}

func TestCustomerBookingRepository_GetBookingByID(t *testing.T) {
	repo, fx := setupCustomerBookingDB(t)
	ctx := context.Background()
	expiresAt := time.Now().Add(15 * time.Minute)

	booking, err := repo.CreateBookingWithSeatLock(ctx, fx.customerID, fx.showID, []int{fx.seatStatusID}, 250.0, expiresAt, "REF-GET")
	require.NoError(t, err)

	tests := []struct {
		name        string
		bookingID   int
		wantErr     bool
		wantStatus  string
		wantPreload bool
	}{
		{
			name:        "returns booking with preloaded Show, Slot, Screen, BookedSeats and Seat for valid ID",
			bookingID:   booking.ID,
			wantErr:     false,
			wantStatus:  "PENDING",
			wantPreload: true,
		},
		{
			name:      "returns error for non-existent booking ID",
			bookingID: 99999,
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := repo.GetBookingByID(ctx, tc.bookingID)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tc.wantStatus, result.Status)
				assert.NotZero(t, result.Show.Id)
				assert.NotEmpty(t, result.BookedSeats)
			}
		})
	}
}

func TestCustomerBookingRepository_ConfirmBooking(t *testing.T) {
	repo, fx := setupCustomerBookingDB(t)
	ctx := context.Background()
	expiresAt := time.Now().Add(15 * time.Minute)

	t.Run("transitions booking from PENDING to CONFIRMED and seats from BLOCKED to BOOKED", func(t *testing.T) {
		db.GetDB().GormDB().Model(&model.ShowSeatStatus{}).Where("id = ?", fx.seatStatusID).Update("status", "AVAILABLE")
		booking, err := repo.CreateBookingWithSeatLock(ctx, fx.customerID, fx.showID, []int{fx.seatStatusID}, 250.0, expiresAt, "REF-CONFIRM")
		require.NoError(t, err)

		err = repo.ConfirmBooking(ctx, booking.ID)
		require.NoError(t, err)

		var updated model.Booking
		db.GetDB().GormDB().First(&updated, booking.ID)
		assert.Equal(t, "CONFIRMED", updated.Status)

		var seatStatus model.ShowSeatStatus
		db.GetDB().GormDB().First(&seatStatus, fx.seatStatusID)
		assert.Equal(t, "BOOKED", seatStatus.Status)
		assert.Nil(t, seatStatus.LockedUntil)
	})

	t.Run("returns error for non-existent booking ID", func(t *testing.T) {
		err := repo.ConfirmBooking(ctx, 99999)
		assert.Error(t, err)
	})

	t.Run("returns error when attempting to confirm an already CONFIRMED booking", func(t *testing.T) {
		db.GetDB().GormDB().Model(&model.ShowSeatStatus{}).Where("id = ?", fx.seatStatusID).Update("status", "AVAILABLE")
		booking, err := repo.CreateBookingWithSeatLock(ctx, fx.customerID, fx.showID, []int{fx.seatStatusID}, 250.0, expiresAt, "REF-DOUBLE-CONFIRM")
		require.NoError(t, err)

		err = repo.ConfirmBooking(ctx, booking.ID)
		require.NoError(t, err)

		err = repo.ConfirmBooking(ctx, booking.ID)
		assert.Error(t, err)
	})
}
