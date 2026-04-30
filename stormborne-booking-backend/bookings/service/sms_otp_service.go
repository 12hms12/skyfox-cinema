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
	smsOTPType          = "SMS"
	smsOTPExpiry        = 5 * time.Minute
	smsMaxResends       = 3
	smsResendWindow     = 5 * time.Minute
	smsMaxAttempts      = 5
	smsOTPExpirySeconds = 300
)

type SMSOTPRepository interface {
	Create(ctx context.Context, otp *model.OTPVerification) error
	FindActiveByRecipient(ctx context.Context, recipient, otpType string) (*model.OTPVerification, error)
	InvalidatePreviousOTPs(ctx context.Context, recipient, otpType string) error
	IncrementAttempts(ctx context.Context, id int) error
	MarkUsed(ctx context.Context, id int) error
	FindAll(ctx context.Context) ([]model.OTPVerification, error)
}

type smsOtpService struct {
	repo       SMSOTPRepository
	mu         sync.Mutex
	resendData map[string]*otpResendInfo
}

type otpResendInfo struct {
	Count     int
	FirstTime time.Time
}

func NewSMSOtpService(repo SMSOTPRepository) *smsOtpService {
	return &smsOtpService{
		repo:       repo,
		resendData: make(map[string]*otpResendInfo),
	}
}

func (s *smsOtpService) RequestOTP(ctx context.Context, phoneNumber, ip string) (*response.OTPSendResponse, error) {
	if err := s.repo.InvalidatePreviousOTPs(ctx, phoneNumber, smsOTPType); err != nil {
		logger.Error("failed to invalidate previous SMS OTPs: %v", err)
	}

	code, err := generateOTPCode()
	if err != nil {
		return nil, ae.InternalServerError("OTPGenerationFailed", "failed to generate OTP", err)
	}

	otp := &model.OTPVerification{
		Code:      code,
		Recipient: phoneNumber,
		Type:      smsOTPType,
		Purpose:   "VERIFICATION",
		ExpiresAt: time.Now().Add(smsOTPExpiry),
		IP:        ip,
	}

	if err := s.repo.Create(ctx, otp); err != nil {
		return nil, err
	}

	logger.Info("SMS OTP generated for %s (illusion mode - no real SMS sent)", phoneNumber)

	return &response.OTPSendResponse{
		Success:   true,
		Message:   "OTP sent to phone number",
		ExpiresIn: smsOTPExpirySeconds,
	}, nil
}

func (s *smsOtpService) VerifyOTP(ctx context.Context, phoneNumber, code string) (*response.OTPVerifyResponse, error) {
	otp, err := s.repo.FindActiveByRecipient(ctx, phoneNumber, smsOTPType)
	if err != nil {
		return &response.OTPVerifyResponse{
			Success: false, Message: "no active OTP found for this phone number", Verified: false,
		}, nil
	}

	if time.Now().After(otp.ExpiresAt) {
		return &response.OTPVerifyResponse{
			Success: false, Message: "OTP has expired", Verified: false,
		}, nil
	}

	if otp.Attempts >= smsMaxAttempts {
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
		Success: true, Message: "phone number verified successfully", Verified: true,
	}, nil
}

func (s *smsOtpService) ResendOTP(ctx context.Context, phoneNumber, ip string) (*response.OTPSendResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	info, exists := s.resendData[phoneNumber]

	if !exists {
		s.resendData[phoneNumber] = &otpResendInfo{Count: 1, FirstTime: now}
	} else {
		if now.Sub(info.FirstTime) > smsResendWindow {
			info.Count = 1
			info.FirstTime = now
		} else {
			if info.Count >= smsMaxResends {
				return nil, ae.BadRequestError(
					"ResendLimitExceeded",
					fmt.Sprintf("resend limit exceeded. try again after %d minutes", int(smsResendWindow.Minutes())),
					nil,
				)
			}
			info.Count++
		}
	}

	return s.RequestOTP(ctx, phoneNumber, ip)
}

func (s *smsOtpService) ListAllOTPs(ctx context.Context) (*response.OTPListResponse, error) {
	otps, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]response.OTPItem, len(otps))
	for i, o := range otps {
		items[i] = response.OTPItem{
			ID:        o.ID,
			Code:      o.Code,
			Recipient: o.Recipient,
			Type:      o.Type,
			Purpose:   o.Purpose,
			ExpiresAt: o.ExpiresAt.Format(time.RFC3339),
			IsUsed:    o.IsUsed,
			Attempts:  o.Attempts,
			CreatedAt: o.CreatedAt.Format(time.RFC3339),
		}
	}

	return &response.OTPListResponse{OTPs: items}, nil
}

func generateOTPCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(9000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%04d", n.Int64()+1000), nil
}
