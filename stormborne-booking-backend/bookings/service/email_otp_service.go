package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"skyfox/bookings/dto/response"
	"skyfox/bookings/model"
	"skyfox/common/logger"
	ae "skyfox/error"
	"sync"
	"time"
)

const (
	emailOTPType          = "EMAIL"
	emailOTPExpiry        = 5 * time.Minute
	emailMaxResends       = 3
	emailResendWindow     = 5 * time.Minute
	emailMaxAttempts      = 5
	emailOTPExpirySeconds = 300
)

type EmailOTPRepository interface {
	Create(ctx context.Context, otp *model.OTPVerification) error
	FindActiveByRecipient(ctx context.Context, recipient, otpType string) (*model.OTPVerification, error)
	InvalidatePreviousOTPs(ctx context.Context, recipient, otpType string) error
	IncrementAttempts(ctx context.Context, id int) error
	MarkUsed(ctx context.Context, id int) error
	FindAll(ctx context.Context) ([]model.OTPVerification, error)
}

type emailOtpService struct {
	repo       EmailOTPRepository
	mu         sync.Mutex
	resendData map[string]*emailResendInfo
}

type emailResendInfo struct {
	Count     int
	FirstTime time.Time
}

func NewEmailOtpService(repo EmailOTPRepository) *emailOtpService {
	return &emailOtpService{
		repo:       repo,
		resendData: make(map[string]*emailResendInfo),
	}
}

func (s *emailOtpService) RequestOTP(ctx context.Context, email, ip string) (*response.OTPSendResponse, error) {
	if err := s.repo.InvalidatePreviousOTPs(ctx, email, emailOTPType); err != nil {
		logger.Error("failed to invalidate previous email OTPs: %v", err)
	}

	code, err := generateEmailOTPCode()
	if err != nil {
		return nil, ae.InternalServerError("OTPGenerationFailed", "failed to generate OTP", err)
	}

	otp := &model.OTPVerification{
		Code:      code,
		Recipient: email,
		Type:      emailOTPType,
		Purpose:   "VERIFICATION",
		ExpiresAt: time.Now().Add(emailOTPExpiry),
		IP:        ip,
	}

	if err := s.repo.Create(ctx, otp); err != nil {
		return nil, err
	}

	logger.Info("Email OTP generated for %s (illusion mode - no real email sent)", email)

	return &response.OTPSendResponse{
		Success:   true,
		Message:   "OTP sent to email address",
		ExpiresIn: emailOTPExpirySeconds,
	}, nil
}

func (s *emailOtpService) VerifyOTP(ctx context.Context, email, code string) (*response.OTPVerifyResponse, error) {
	otp, err := s.repo.FindActiveByRecipient(ctx, email, emailOTPType)
	if err != nil {
		return &response.OTPVerifyResponse{
			Success: false, Message: "no active OTP found for this email address", Verified: false,
		}, nil
	}

	if time.Now().After(otp.ExpiresAt) {
		return &response.OTPVerifyResponse{
			Success: false, Message: "OTP has expired", Verified: false,
		}, nil
	}

	if otp.Attempts >= emailMaxAttempts {
		return &response.OTPVerifyResponse{
			Success: false, Message: "maximum verification attempts exceeded", Verified: false,
		}, nil
	}

	_ = s.repo.IncrementAttempts(ctx, otp.ID)

	if otp.Code != code {
		return &response.OTPVerifyResponse{
			Success: false, Message: "incorrect OTP", Verified: false,
		}, nil
	}

	if err := s.repo.MarkUsed(ctx, otp.ID); err != nil {
		return nil, err
	}

	return &response.OTPVerifyResponse{
		Success: true, Message: "email address verified successfully", Verified: true,
	}, nil
}

func (s *emailOtpService) ResendOTP(ctx context.Context, email, ip string) (*response.OTPSendResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	info, exists := s.resendData[email]

	if !exists {
		s.resendData[email] = &emailResendInfo{Count: 1, FirstTime: now}
	} else {
		if now.Sub(info.FirstTime) > emailResendWindow {
			info.Count = 1
			info.FirstTime = now
		} else {
			if info.Count >= emailMaxResends {
				return nil, ae.BadRequestError(
					"ResendLimitExceeded",
					fmt.Sprintf("resend limit exceeded. try again after %d minutes", int(emailResendWindow.Minutes())),
					nil,
				)
			}
			info.Count++
		}
	}

	return s.RequestOTP(ctx, email, ip)
}

func generateEmailOTPCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(9000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%04d", n.Int64()+1000), nil
}


func (s *emailOtpService) IsRecipientVerified(ctx context.Context, recipient, otpType string) (bool, error) {
	otp, err := s.repo.FindActiveByRecipient(ctx, recipient, otpType)
	if err != nil {
		return false, nil
	}
	return otp.IsUsed, nil
}