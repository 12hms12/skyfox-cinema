package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"skyfox/bookings/model"
	"skyfox/bookings/service/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEmailRequestOTP(t *testing.T) {
	ctx := context.Background()


	tests := []struct {
		name      string
		email     string
		ip        string
		setupMock func(repo *mocks.MockEmailOTPRepository)
		wantErr   bool
		wantMsg   string
	}{
		{
			name:  "should generate a new email OTP and store it in database after invalidating all previous active OTPs for the same email",
			email: "user@example.com",
			ip:    "192.168.1.1",
			setupMock: func(repo *mocks.MockEmailOTPRepository) {
				repo.On("InvalidatePreviousOTPs", ctx, "user@example.com", "EMAIL").Return(nil)
				repo.On("Create", ctx, mock.AnythingOfType("*model.OTPVerification")).Return(nil)
			},
			wantErr: false,
			wantMsg: "OTP sent to email address",
		},
		{
			name:  "should still generate a new OTP even when invalidation of previous OTPs fails due to a database error",
			email: "user@example.com",
			ip:    "10.0.0.1",
			setupMock: func(repo *mocks.MockEmailOTPRepository) {
				repo.On("InvalidatePreviousOTPs", ctx, "user@example.com", "EMAIL").Return(errors.New("db timeout"))
				repo.On("Create", ctx, mock.AnythingOfType("*model.OTPVerification")).Return(nil)
			},
			wantErr: false,
			wantMsg: "OTP sent to email address",
		},
		{
			name:  "should return an error when the OTP record cannot be persisted to the database",
			email: "user@example.com",
			ip:    "10.0.0.1",
			setupMock: func(repo *mocks.MockEmailOTPRepository) {
				repo.On("InvalidatePreviousOTPs", ctx, "user@example.com", "EMAIL").Return(nil)
				repo.On("Create", ctx, mock.AnythingOfType("*model.OTPVerification")).Return(errors.New("insert failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockEmailOTPRepository(t)
			tt.setupMock(repo)

			svc := NewEmailOtpService(repo)
			resp, err := svc.RequestOTP(ctx, tt.email, tt.ip)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.True(t, resp.Success)
				assert.Equal(t, tt.wantMsg, resp.Message)
				assert.Equal(t, emailOTPExpirySeconds, resp.ExpiresIn)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestEmailVerifyOTP(t *testing.T) {
	ctx := context.Background()

	activeOTP := &model.OTPVerification{
		ID:        1,
		Code:      "1234",
		Recipient: "user@example.com",
		Type:      "EMAIL",
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Attempts:  0,
		IsUsed:    false,
	}

	expiredOTP := &model.OTPVerification{
		ID:        2,
		Code:      "5678",
		Recipient: "user@example.com",
		Type:      "EMAIL",
		ExpiresAt: time.Now().Add(-1 * time.Minute),
		Attempts:  0,
	}

	maxedOutOTP := &model.OTPVerification{
		ID:        3,
		Code:      "9999",
		Recipient: "user@example.com",
		Type:      "EMAIL",
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Attempts:  5,
	}

	tests := []struct {
		name         string
		email        string
		code         string
		setupMock    func(repo *mocks.MockEmailOTPRepository)
		wantVerified bool
		wantMsg      string
	}{
		{
			name:  "should verify successfully when the correct 4-digit code is provided for an active unexpired OTP",
			email: "user@example.com",
			code:  "1234",
			setupMock: func(repo *mocks.MockEmailOTPRepository) {
				repo.On("FindActiveByRecipient", ctx, "user@example.com", "EMAIL").Return(activeOTP, nil)
				repo.On("IncrementAttempts", ctx, 1).Return(nil)
				repo.On("MarkUsed", ctx, 1).Return(nil)
			},
			wantVerified: true,
			wantMsg:      "email address verified successfully",
		},
		{
			name:  "should reject verification when the entered OTP code does not match the stored code",
			email: "user@example.com",
			code:  "0000",
			setupMock: func(repo *mocks.MockEmailOTPRepository) {
				repo.On("FindActiveByRecipient", ctx, "user@example.com", "EMAIL").Return(activeOTP, nil)
				repo.On("IncrementAttempts", ctx, 1).Return(nil)
			},
			wantVerified: false,
			wantMsg:      "incorrect OTP",
		},
		{
			name:  "should reject verification when no active unexpired OTP record exists for the given email",
			email: "nobody@example.com",
			code:  "1234",
			setupMock: func(repo *mocks.MockEmailOTPRepository) {
				repo.On("FindActiveByRecipient", ctx, "nobody@example.com", "EMAIL").Return(nil, errors.New("not found"))
			},
			wantVerified: false,
			wantMsg:      "no active OTP found for this email address",
		},
		{
			name:  "should reject verification when the OTP has passed its 5-minute expiration window",
			email: "user@example.com",
			code:  "5678",
			setupMock: func(repo *mocks.MockEmailOTPRepository) {
				repo.On("FindActiveByRecipient", ctx, "user@example.com", "EMAIL").Return(expiredOTP, nil)
			},
			wantVerified: false,
			wantMsg:      "OTP has expired",
		},
		{
			name:  "should reject verification when the maximum of 5 verification attempts has been exhausted",
			email: "user@example.com",
			code:  "9999",
			setupMock: func(repo *mocks.MockEmailOTPRepository) {
				repo.On("FindActiveByRecipient", ctx, "user@example.com", "EMAIL").Return(maxedOutOTP, nil)
			},
			wantVerified: false,
			wantMsg:      "maximum verification attempts exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockEmailOTPRepository(t)
			tt.setupMock(repo)

			svc := NewEmailOtpService(repo)
			resp, err := svc.VerifyOTP(ctx, tt.email, tt.code)

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.wantVerified, resp.Verified)
			assert.Contains(t, resp.Message, tt.wantMsg)
			repo.AssertExpectations(t)
		})
	}
}

func TestEmailResendOTP(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		callCount int
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "should successfully resend OTP on the first resend attempt within the 5-minute rate limit window",
			callCount: 1,
			wantErr:   false,
		},
		{
			name:      "should successfully resend OTP on the second resend attempt within the 5-minute rate limit window",
			callCount: 2,
			wantErr:   false,
		},
		{
			name:      "should successfully resend OTP on the third and final allowed resend attempt within the 5-minute window",
			callCount: 3,
			wantErr:   false,
		},
		{
			name:      "should reject the fourth resend attempt within the same 5-minute window because the 3-resend rate limit is exceeded",
			callCount: 4,
			wantErr:   true,
			errMsg:    "resend limit exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockEmailOTPRepository(t)

			repo.On("InvalidatePreviousOTPs", ctx, "user@example.com", "EMAIL").Return(nil)
			repo.On("Create", ctx, mock.AnythingOfType("*model.OTPVerification")).Return(nil)

			svc := NewEmailOtpService(repo)

			var lastErr error
			for i := 0; i < tt.callCount; i++ {
				_, lastErr = svc.ResendOTP(ctx, "user@example.com", "10.0.0.1")
			}

			if tt.wantErr {
				assert.Error(t, lastErr)
				assert.Contains(t, lastErr.Error(), tt.errMsg)
			} else {
				assert.NoError(t, lastErr)
			}
		})
	}
}
