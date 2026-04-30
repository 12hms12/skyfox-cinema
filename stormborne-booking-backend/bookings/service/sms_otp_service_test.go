package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"skyfox/bookings/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSMSOTPRepository struct {
	mock.Mock
}

func (m *MockSMSOTPRepository) Create(ctx context.Context, otp *model.OTPVerification) error {
	args := m.Called(ctx, otp)
	return args.Error(0)
}

func (m *MockSMSOTPRepository) FindActiveByRecipient(ctx context.Context, recipient, otpType string) (*model.OTPVerification, error) {
	args := m.Called(ctx, recipient, otpType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.OTPVerification), args.Error(1)
}

func (m *MockSMSOTPRepository) InvalidatePreviousOTPs(ctx context.Context, recipient, otpType string) error {
	args := m.Called(ctx, recipient, otpType)
	return args.Error(0)
}

func (m *MockSMSOTPRepository) IncrementAttempts(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSMSOTPRepository) MarkUsed(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSMSOTPRepository) FindAll(ctx context.Context) ([]model.OTPVerification, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.OTPVerification), args.Error(1)
}

func TestSMSRequestOTP(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		phone     string
		ip        string
		setupMock func(repo *MockSMSOTPRepository)
		wantErr   bool
		wantMsg   string
	}{
		{
			name:  "should generate a new SMS OTP and store it in database after invalidating all previous active OTPs for the same phone number",
			phone: "+919876543210",
			ip:    "192.168.1.1",
			setupMock: func(repo *MockSMSOTPRepository) {
				repo.On("InvalidatePreviousOTPs", ctx, "+919876543210", "SMS").Return(nil)
				repo.On("Create", ctx, mock.AnythingOfType("*model.OTPVerification")).Return(nil)
			},
			wantErr: false,
			wantMsg: "OTP sent to phone number",
		},
		{
			name:  "should still generate a new OTP even when invalidation of previous OTPs fails due to a database error",
			phone: "+919876543210",
			ip:    "10.0.0.1",
			setupMock: func(repo *MockSMSOTPRepository) {
				repo.On("InvalidatePreviousOTPs", ctx, "+919876543210", "SMS").Return(errors.New("db timeout"))
				repo.On("Create", ctx, mock.AnythingOfType("*model.OTPVerification")).Return(nil)
			},
			wantErr: false,
			wantMsg: "OTP sent to phone number",
		},
		{
			name:  "should return an error when the OTP record cannot be persisted to the database",
			phone: "+919876543210",
			ip:    "10.0.0.1",
			setupMock: func(repo *MockSMSOTPRepository) {
				repo.On("InvalidatePreviousOTPs", ctx, "+919876543210", "SMS").Return(nil)
				repo.On("Create", ctx, mock.AnythingOfType("*model.OTPVerification")).Return(errors.New("insert failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockSMSOTPRepository)
			tt.setupMock(repo)

			svc := NewSMSOtpService(repo)
			resp, err := svc.RequestOTP(ctx, tt.phone, tt.ip)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.True(t, resp.Success)
				assert.Equal(t, tt.wantMsg, resp.Message)
				assert.Equal(t, smsOTPExpirySeconds, resp.ExpiresIn)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestSMSVerifyOTP(t *testing.T) {
	ctx := context.Background()

	activeOTP := &model.OTPVerification{
		ID:        1,
		Code:      "1234",
		Recipient: "+919876543210",
		Type:      "SMS",
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Attempts:  0,
		IsUsed:    false,
	}

	expiredOTP := &model.OTPVerification{
		ID:        2,
		Code:      "5678",
		Recipient: "+919876543210",
		Type:      "SMS",
		ExpiresAt: time.Now().Add(-1 * time.Minute),
		Attempts:  0,
	}

	maxedOutOTP := &model.OTPVerification{
		ID:        3,
		Code:      "9999",
		Recipient: "+919876543210",
		Type:      "SMS",
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Attempts:  5,
	}

	tests := []struct {
		name         string
		phone        string
		code         string
		setupMock    func(repo *MockSMSOTPRepository)
		wantVerified bool
		wantMsg      string
	}{
		{
			name:  "should verify successfully when the correct 4-digit code is provided for an active unexpired OTP with remaining attempts",
			phone: "+919876543210",
			code:  "1234",
			setupMock: func(repo *MockSMSOTPRepository) {
				repo.On("FindActiveByRecipient", ctx, "+919876543210", "SMS").Return(activeOTP, nil)
				repo.On("IncrementAttempts", ctx, 1).Return(nil)
				repo.On("MarkUsed", ctx, 1).Return(nil)
			},
			wantVerified: true,
			wantMsg:      "phone number verified successfully",
		},
		{
			name:  "should reject verification when the entered OTP code does not match the stored code for the active OTP record",
			phone: "+919876543210",
			code:  "0000",
			setupMock: func(repo *MockSMSOTPRepository) {
				repo.On("FindActiveByRecipient", ctx, "+919876543210", "SMS").Return(activeOTP, nil)
				repo.On("IncrementAttempts", ctx, 1).Return(nil)
			},
			wantVerified: false,
			wantMsg:      "incorrect OTP",
		},
		{
			name:  "should reject verification when no active unexpired OTP record exists for the given phone number",
			phone: "+919999999999",
			code:  "1234",
			setupMock: func(repo *MockSMSOTPRepository) {
				repo.On("FindActiveByRecipient", ctx, "+919999999999", "SMS").Return(nil, errors.New("not found"))
			},
			wantVerified: false,
			wantMsg:      "no active OTP found for this phone number",
		},
		{
			name:  "should reject verification when the OTP has passed its 5-minute expiration window even if the code is correct",
			phone: "+919876543210",
			code:  "5678",
			setupMock: func(repo *MockSMSOTPRepository) {
				repo.On("FindActiveByRecipient", ctx, "+919876543210", "SMS").Return(expiredOTP, nil)
			},
			wantVerified: false,
			wantMsg:      "OTP has expired",
		},
		{
			name:  "should reject verification when the maximum of 5 verification attempts has been exhausted for this OTP",
			phone: "+919876543210",
			code:  "9999",
			setupMock: func(repo *MockSMSOTPRepository) {
				repo.On("FindActiveByRecipient", ctx, "+919876543210", "SMS").Return(maxedOutOTP, nil)
			},
			wantVerified: false,
			wantMsg:      "maximum verification attempts exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockSMSOTPRepository)
			tt.setupMock(repo)

			svc := NewSMSOtpService(repo)
			resp, err := svc.VerifyOTP(ctx, tt.phone, tt.code)

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.wantVerified, resp.Verified)
			assert.Contains(t, resp.Message, tt.wantMsg)
			repo.AssertExpectations(t)
		})
	}
}

func TestSMSResendOTP(t *testing.T) {
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
			repo := new(MockSMSOTPRepository)

			repo.On("InvalidatePreviousOTPs", ctx, "+919876543210", "SMS").Return(nil)
			repo.On("Create", ctx, mock.AnythingOfType("*model.OTPVerification")).Return(nil)

			svc := NewSMSOtpService(repo)

			var lastErr error
			for i := 0; i < tt.callCount; i++ {
				_, lastErr = svc.ResendOTP(ctx, "+919876543210", "10.0.0.1")
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

func TestSMSListAllOTPs(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		setupMock func(repo *MockSMSOTPRepository)
		wantCount int
		wantErr   bool
	}{
		{
			name: "should return all OTP records from the database ordered by creation time for the illusion dashboard",
			setupMock: func(repo *MockSMSOTPRepository) {
				repo.On("FindAll", ctx).Return([]model.OTPVerification{
					{ID: 1, Code: "1234", Recipient: "+919876543210", Type: "SMS", ExpiresAt: time.Now(), CreatedAt: time.Now()},
					{ID: 2, Code: "5678", Recipient: "user@test.com", Type: "EMAIL", ExpiresAt: time.Now(), CreatedAt: time.Now()},
				}, nil)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "should return an empty list when no OTP records have been generated yet",
			setupMock: func(repo *MockSMSOTPRepository) {
				repo.On("FindAll", ctx).Return([]model.OTPVerification{}, nil)
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "should propagate the database error when the repository fails to fetch OTP records",
			setupMock: func(repo *MockSMSOTPRepository) {
				repo.On("FindAll", ctx).Return([]model.OTPVerification(nil), errors.New("db error"))
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockSMSOTPRepository)
			tt.setupMock(repo)

			svc := NewSMSOtpService(repo)
			resp, err := svc.ListAllOTPs(ctx)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Len(t, resp.OTPs, tt.wantCount)
			}
			repo.AssertExpectations(t)
		})
	}
}
