package repository

import (
	"context"
	"errors"
	"skyfox/bookings/database/common"
	"skyfox/bookings/model"
	ae "skyfox/error"

	"gorm.io/gorm"
)

type otpRepository struct {
	*common.BaseDB
}

func NewOTPRepository(db *common.BaseDB) *otpRepository {
	return &otpRepository{BaseDB: db}
}

func (repo *otpRepository) Create(ctx context.Context, otp *model.OTPVerification) error {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	result := dbCtx.Create(otp)
	if result.Error != nil {
		return ae.InternalServerError("OTPCreationFailed", "failed to create OTP", result.Error)
	}
	return nil
}

func (repo *otpRepository) FindActiveByRecipient(ctx context.Context, recipient, otpType string) (*model.OTPVerification, error) {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	var otp model.OTPVerification
	result := dbCtx.
		Where("recipient = ? AND type = ? AND is_used = false AND expires_at > (NOW() AT TIME ZONE 'UTC')", recipient, otpType).
		Order("created_at DESC").
		First(&otp)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ae.NotFoundError("OTPNotFound", "no active OTP found", result.Error)
		}
		return nil, ae.InternalServerError("OTPFetchFailed", "failed to fetch OTP", result.Error)
	}
	return &otp, nil
}

func (repo *otpRepository) InvalidatePreviousOTPs(ctx context.Context, recipient, otpType string) error {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	result := dbCtx.
		Model(&model.OTPVerification{}).
		Where("recipient = ? AND type = ? AND is_used = false", recipient, otpType).
		Update("is_used", true)

	if result.Error != nil {
		return ae.InternalServerError("OTPInvalidationFailed", "failed to invalidate previous OTPs", result.Error)
	}
	return nil
}

func (repo *otpRepository) IncrementAttempts(ctx context.Context, id int) error {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	result := dbCtx.
		Model(&model.OTPVerification{}).
		Where("id = ?", id).
		UpdateColumn("attempts", gorm.Expr("attempts + 1"))

	if result.Error != nil {
		return ae.InternalServerError("OTPUpdateFailed", "failed to increment OTP attempts", result.Error)
	}
	return nil
}

func (repo *otpRepository) MarkUsed(ctx context.Context, id int) error {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	result := dbCtx.
		Model(&model.OTPVerification{}).
		Where("id = ?", id).
		Update("is_used", true)

	if result.Error != nil {
		return ae.InternalServerError("OTPUpdateFailed", "failed to mark OTP as used", result.Error)
	}
	return nil
}

func (repo *otpRepository) IsRecipientVerified(ctx context.Context, recipient, otpType string) (bool, error) {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	var count int64
	result := dbCtx.
		Model(&model.OTPVerification{}).
		Where("recipient = ? AND type = ? AND is_used = true", recipient, otpType).
		Count(&count)

	if result.Error != nil {
		return false, ae.InternalServerError("OTPCheckFailed", "failed to check OTP verification status", result.Error)
	}
	return count > 0, nil
}

func (repo *otpRepository) FindAll(ctx context.Context) ([]model.OTPVerification, error) {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	var otps []model.OTPVerification
	result := dbCtx.Order("created_at DESC").Find(&otps)
	if result.Error != nil {
		return nil, ae.InternalServerError("OTPFetchFailed", "failed to fetch OTPs", result.Error)
	}
	return otps, nil
}
