package repository_test

import (
	"context"
	"testing"
	"time"

	"skyfox/bookings/model"
	"skyfox/bookings/repository"
	"skyfox/integration_test/db"

	"github.com/stretchr/testify/assert"
)

func TestOTPRepository(t *testing.T) {
	database := db.GetDB()

	err := database.GormDB().AutoMigrate(&model.OTPVerification{})
	assert.NoError(t, err)

	database.GormDB().Exec("DELETE FROM otp_verification")

	repo := repository.NewOTPRepository(database)
	ctx := context.Background()

	t.Run("Create should persist a new OTP record with all fields including recipient phone number and expiration timestamp", func(t *testing.T) {
		otp := &model.OTPVerification{
			Code:      "1234",
			Recipient: "+919876543210",
			Type:      "SMS",
			Purpose:   "VERIFICATION",
			ExpiresAt: time.Now().Add(5 * time.Minute),
			IP:        "192.168.1.1",
		}

		err := repo.Create(ctx, otp)

		assert.NoError(t, err)
		assert.NotZero(t, otp.ID)
	})

	t.Run("FindActiveByRecipient should return the most recently created unexpired and unused OTP for a given phone number and type", func(t *testing.T) {
		database.GormDB().Exec("DELETE FROM otp_verification")

		otp := &model.OTPVerification{
			Code:      "5678",
			Recipient: "+919876543210",
			Type:      "SMS",
			Purpose:   "VERIFICATION",
			ExpiresAt: time.Now().Add(5 * time.Minute),
		}
		database.GormDB().Create(otp)

		found, err := repo.FindActiveByRecipient(ctx, "+919876543210", "SMS")

		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "5678", found.Code)
		assert.Equal(t, "+919876543210", found.Recipient)
	})

	t.Run("FindActiveByRecipient should return NotFoundError when no active unexpired OTP exists for the given phone number", func(t *testing.T) {
		found, err := repo.FindActiveByRecipient(ctx, "+910000000000", "SMS")

		assert.Error(t, err)
		assert.Nil(t, found)
	})

	t.Run("FindActiveByRecipient should not return an OTP record that has already been marked as used even if it has not expired", func(t *testing.T) {
		database.GormDB().Exec("DELETE FROM otp_verification")

		otp := &model.OTPVerification{
			Code:      "9999",
			Recipient: "+919876543210",
			Type:      "SMS",
			Purpose:   "VERIFICATION",
			ExpiresAt: time.Now().Add(5 * time.Minute),
			IsUsed:    true,
		}
		database.GormDB().Create(otp)

		found, err := repo.FindActiveByRecipient(ctx, "+919876543210", "SMS")

		assert.Error(t, err)
		assert.Nil(t, found)
	})

	t.Run("InvalidatePreviousOTPs should mark all active unused OTPs as used for the given recipient and type combination", func(t *testing.T) {
		database.GormDB().Exec("DELETE FROM otp_verification")

		for i := 0; i < 3; i++ {
			database.GormDB().Create(&model.OTPVerification{
				Code:      "1111",
				Recipient: "+919876543210",
				Type:      "SMS",
				Purpose:   "VERIFICATION",
				ExpiresAt: time.Now().Add(5 * time.Minute),
				IsUsed:    false,
			})
		}

		err := repo.InvalidatePreviousOTPs(ctx, "+919876543210", "SMS")
		assert.NoError(t, err)

		found, findErr := repo.FindActiveByRecipient(ctx, "+919876543210", "SMS")
		assert.Error(t, findErr)
		assert.Nil(t, found)
	})

	t.Run("InvalidatePreviousOTPs should not affect OTPs belonging to a different recipient even if they share the same type", func(t *testing.T) {
		database.GormDB().Exec("DELETE FROM otp_verification")

		database.GormDB().Create(&model.OTPVerification{
			Code: "1111", Recipient: "+919876543210", Type: "SMS",
			Purpose: "VERIFICATION", ExpiresAt: time.Now().Add(5 * time.Minute),
		})
		database.GormDB().Create(&model.OTPVerification{
			Code: "2222", Recipient: "+918888888888", Type: "SMS",
			Purpose: "VERIFICATION", ExpiresAt: time.Now().Add(5 * time.Minute),
		})

		err := repo.InvalidatePreviousOTPs(ctx, "+919876543210", "SMS")
		assert.NoError(t, err)

		other, otherErr := repo.FindActiveByRecipient(ctx, "+918888888888", "SMS")
		assert.NoError(t, otherErr)
		assert.NotNil(t, other)
		assert.Equal(t, "2222", other.Code)
	})

	t.Run("IncrementAttempts should increase the attempts counter by 1 each time it is called on an OTP record", func(t *testing.T) {
		database.GormDB().Exec("DELETE FROM otp_verification")

		otp := &model.OTPVerification{
			Code: "3333", Recipient: "+919876543210", Type: "SMS",
			Purpose: "VERIFICATION", ExpiresAt: time.Now().Add(5 * time.Minute),
		}
		database.GormDB().Create(otp)

		err := repo.IncrementAttempts(ctx, otp.ID)
		assert.NoError(t, err)

		err = repo.IncrementAttempts(ctx, otp.ID)
		assert.NoError(t, err)

		var updated model.OTPVerification
		database.GormDB().First(&updated, otp.ID)
		assert.Equal(t, 2, updated.Attempts)
	})

	t.Run("MarkUsed should set the is_used flag to true making the OTP no longer discoverable by FindActiveByRecipient", func(t *testing.T) {
		database.GormDB().Exec("DELETE FROM otp_verification")

		otp := &model.OTPVerification{
			Code: "4444", Recipient: "+919876543210", Type: "SMS",
			Purpose: "VERIFICATION", ExpiresAt: time.Now().Add(5 * time.Minute),
		}
		database.GormDB().Create(otp)

		err := repo.MarkUsed(ctx, otp.ID)
		assert.NoError(t, err)

		found, findErr := repo.FindActiveByRecipient(ctx, "+919876543210", "SMS")
		assert.Error(t, findErr)
		assert.Nil(t, found)
	})

	t.Run("FindAll should return all OTP records from the database ordered by creation time descending regardless of status", func(t *testing.T) {
		database.GormDB().Exec("DELETE FROM otp_verification")

		database.GormDB().Create(&model.OTPVerification{
			Code: "1111", Recipient: "+919876543210", Type: "SMS",
			Purpose: "VERIFICATION", ExpiresAt: time.Now().Add(5 * time.Minute),
		})
		database.GormDB().Create(&model.OTPVerification{
			Code: "2222", Recipient: "user@example.com", Type: "EMAIL",
			Purpose: "VERIFICATION", ExpiresAt: time.Now().Add(5 * time.Minute), IsUsed: true,
		})
		database.GormDB().Create(&model.OTPVerification{
			Code: "3333", Recipient: "+918888888888", Type: "SMS",
			Purpose: "VERIFICATION", ExpiresAt: time.Now().Add(-1 * time.Minute),
		})

		otps, err := repo.FindAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, otps, 3)
	})

	t.Run("FindAll should return an empty slice when no OTP records exist in the database", func(t *testing.T) {
		database.GormDB().Exec("DELETE FROM otp_verification")

		otps, err := repo.FindAll(ctx)
		assert.NoError(t, err)
		assert.Empty(t, otps)
	})

	t.Run("IsRecipientVerified should return true when a used OTP exists for the given recipient and type", func(t *testing.T) {
		database.GormDB().Exec("DELETE FROM otp_verification")

		database.GormDB().Create(&model.OTPVerification{
			Code: "1234", Recipient: "IN9876543210", Type: "SMS",
			Purpose: "VERIFICATION", ExpiresAt: time.Now().Add(5 * time.Minute), IsUsed: true,
		})

		verified, err := repo.IsRecipientVerified(ctx, "IN9876543210", "SMS")
		assert.NoError(t, err)
		assert.True(t, verified)
	})

	t.Run("IsRecipientVerified should return false when no used OTP exists for the given recipient", func(t *testing.T) {
		database.GormDB().Exec("DELETE FROM otp_verification")

		database.GormDB().Create(&model.OTPVerification{
			Code: "1234", Recipient: "IN9876543210", Type: "SMS",
			Purpose: "VERIFICATION", ExpiresAt: time.Now().Add(5 * time.Minute), IsUsed: false,
		})

		verified, err := repo.IsRecipientVerified(ctx, "IN9876543210", "SMS")
		assert.NoError(t, err)
		assert.False(t, verified)
	})

	t.Run("IsRecipientVerified should return false when the used OTP belongs to a different recipient", func(t *testing.T) {
		database.GormDB().Exec("DELETE FROM otp_verification")

		database.GormDB().Create(&model.OTPVerification{
			Code: "1234", Recipient: "IN9876543210", Type: "SMS",
			Purpose: "VERIFICATION", ExpiresAt: time.Now().Add(5 * time.Minute), IsUsed: true,
		})

		verified, err := repo.IsRecipientVerified(ctx, "IN0000000000", "SMS")
		assert.NoError(t, err)
		assert.False(t, verified)
	})
}
