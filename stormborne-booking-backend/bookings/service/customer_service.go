package service

import (
	"context"
	"errors"
	"fmt"
	"skyfox/bookings/dto/request"
	"skyfox/bookings/dto/response"
	"skyfox/bookings/model"
	credentialSecurity "skyfox/bookings/security"
	"skyfox/common/logger"
	"skyfox/common/middleware/security"
	"skyfox/common/middleware/validator"
	"skyfox/config"
	ae "skyfox/error"
	"time"

	"strings"

	"github.com/nyaruka/phonenumbers"
	"golang.org/x/crypto/bcrypt"
	"strconv"
)

type CustomerRepository interface {
	Create(context.Context, *model.Customer) error
	FindByUsername(context.Context, string) (*model.Customer, error)
	FindByEmail(ctx context.Context, email string) (*model.Customer, error)
	FindByResetToken(ctx context.Context, token string) (*model.Customer, error)
	Update(ctx context.Context, customer *model.Customer) error
	FindByID(ctx context.Context, userID uint) (*model.Customer, error)
	UpdateProfile(ctx context.Context, customer *model.Customer) error
}
type resendInfo struct {
	Count     int
	FirstTime time.Time
}
type avatarAssigner interface {
	AssignRandomAvatar(ctx context.Context, userID uint, gender string) (*response.AvatarResponse, error)
}

type PhoneOTPChecker interface {
	IsRecipientVerified(ctx context.Context, recipient, otpType string) (bool, error)
}

type EmailOTPServiceInterface interface {
	RequestOTP(ctx context.Context, email, ip string) (*response.OTPSendResponse, error)
	VerifyOTP(ctx context.Context, email, code string) (*response.OTPVerifyResponse, error)
	IsRecipientVerified(ctx context.Context, recipient, otpType string) (bool, error)
	ResendOTP(ctx context.Context, email, ip string) (*response.OTPSendResponse, error)
}

type pendingProfileUpdate struct {
	FirstName string
	LastName  string
	Email     string
}

type CustomerService struct {
	repo            CustomerRepository
	jwtManager      *security.JwtManager
	cfg             *config.ServerConfig
	avatarService   avatarAssigner
	phoneOTPChecker PhoneOTPChecker
	emailOTPService EmailOTPServiceInterface
	resendData      map[string]*resendInfo
	pendingUpdates  map[uint]*pendingProfileUpdate
}

func NewCustomerService(repo CustomerRepository, jwtManager *security.JwtManager, avatarService avatarAssigner, cfg *config.ServerConfig, phoneOTPChecker PhoneOTPChecker, emailOTPService EmailOTPServiceInterface) *CustomerService {
	return &CustomerService{
		repo:            repo,
		jwtManager:      jwtManager,
		cfg:             cfg,
		avatarService:   avatarService,
		phoneOTPChecker: phoneOTPChecker,
		emailOTPService: emailOTPService,
		resendData:      make(map[string]*resendInfo),
		pendingUpdates:  make(map[uint]*pendingProfileUpdate),
	}
}

func (s *CustomerService) Signup(ctx context.Context, req request.CustomerSignupRequest) (*response.LoginResponse, error) {

	isValidEmail := validator.ValidateEmail(req.Email)
	if !isValidEmail {
		logger.Error("error occurred while parsing email.")
		return nil, ae.InvalidCredentialsError(
			"Invalid email",
			"error occurred while sign up, invalid email",
			nil,
		)
	}

	parsedNumber, err := phonenumbers.Parse(
		req.PhoneNumber,
		strings.ToUpper(req.CountryCode),
	)
	if err != nil {
		logger.Error("error occurred while parsing phone number. %v", err)
		return nil, err
	}

	isValidPhoneNumber := phonenumbers.IsValidNumber(parsedNumber)
	if !isValidPhoneNumber {
		logger.Error("error occurred while validating phone number.")
		return nil, ae.InvalidCredentialsError(
			"Invalid phone number",
			"error occurred while sign up, invalid phone number",
			nil,
		)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("error occurred while hashing password. %v", err)
		return nil, err
	}

	customer := model.Customer{
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		CountryCode: req.CountryCode,
		Age:         req.Age,
		Gender:      req.Gender,
		Username:    req.Username,
		Password:    string(hashedPassword),
		UserRole:    string(security.USER),
	}

	if s.phoneOTPChecker != nil {
		recipient := req.CountryCode + req.PhoneNumber
		verified, err := s.phoneOTPChecker.IsRecipientVerified(ctx, recipient, "SMS")
		if err != nil {
			logger.Error("failed to check phone OTP verification status: %v", err)
		} else if verified {
			customer.IsPhoneVerified = true
		}
	}

	if s.emailOTPService != nil {
		verified, err := s.emailOTPService.IsRecipientVerified(ctx, req.Email, "EMAIL")
		if err != nil {
			logger.Error("failed to check email OTP verification status: %v", err)
		} else if verified {
			customer.IsEmailVerified = true
		}
	}

	if err := s.repo.Create(ctx, &customer); err != nil {
		logger.Error("error occurred while creating customer. %v", err)
		return nil, err
	}

	avatarResp, err := s.avatarService.AssignRandomAvatar(ctx, customer.ID, string(customer.Gender))
	if err != nil {
		logger.Error("failed to assign random avatar during signup, continuing without avatar. %v", err)
	} else {
		customer.AvatarURL = avatarResp.AvatarURL
		customer.AvatarType = model.AvatarType(avatarResp.AvatarType)
	}

	return nil, nil
}

func (s *CustomerService) Login(
	ctx context.Context,
	req request.CustomerLoginRequest,
) (*response.LoginResponse, error) {

	if strings.TrimSpace(req.Email) == "" &&
		strings.TrimSpace(req.Username) == "" {
		return nil, ae.BadRequestError(
			"BadRequest",
			"email or username is required",
			nil,
		)
	}

	if strings.TrimSpace(req.Password) == "" {
		return nil, ae.BadRequestError(
			"BadRequest",
			"password is required",
			nil,
		)
	}

	var (
		customer *model.Customer
		err      error
	)

	if strings.TrimSpace(req.Email) != "" {
		customer, err = s.repo.FindByEmail(ctx, req.Email)
	} else {
		customer, err = s.repo.FindByUsername(ctx, req.Username)
	}

	if customer == nil || customer.ID == 0 || err != nil {
		logger.Error("error occurred while fetching customer: %v", err)
		return nil, ae.InvalidCredentialsError(
			"WrongCredentials",
			"email/username or password is wrong",
			nil,
		)
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(customer.Password),
		[]byte(req.Password),
	)
	if err != nil {
		return nil, ae.InvalidCredentialsError(
			"WrongCredentials",
			"email/username or password is wrong",
			nil,
		)
	}

	token, err := s.jwtManager.GenerateToken(
		int64(customer.ID),
		security.Role(customer.UserRole),
	)
	if err != nil {
		logger.Error("error occurred while generating token: %v", err)
		return nil, ae.InternalServerError(
			"TokenGenerationFailed",
			"failed to generate token",
			err,
		)
	}

	res := &response.LoginResponse{
		Token:      token,
		Username:   customer.Username,
		AvatarURL:  customer.AvatarURL,
		AvatarType: customer.AvatarType,
		Gender:     customer.Gender,
		Age:        customer.Age,
		Role:       string(customer.UserRole),
		Id:         strconv.FormatUint(uint64(customer.ID), 10),
	}

	return res, nil
}

func (s *CustomerService) ForgotPassword(ctx context.Context, email string) (string, error) {
	customer, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	now := time.Now()
	info, exists := s.resendData[email]

	if !exists {
		s.resendData[email] = &resendInfo{
			Count:     1,
			FirstTime: now,
		}
	} else {
		if now.Sub(info.FirstTime) > 5*time.Minute {
			info.Count = 1
			info.FirstTime = now
		} else {
			if info.Count >= 3 {
				return "", errors.New("resend limit exceeded.Try after 5 minutes")
			}
			info.Count++
		}
	}

	token, err := credentialSecurity.GeneratePasswordResetToken()
	if err != nil {
		fmt.Print("Generate Token Failed")
		return "", err
	}
	customer.PasswordResetToken = &token
	customer.PasswordTokenExpiryTime = time.Now().Add(30 * time.Minute)

	err1 := s.repo.Update(ctx, customer)
	if err1 != nil {
		return "", err1
	}

	resetLink := s.cfg.WebClientResetPasswordBaseUrl + token

	return resetLink, nil
}

func (s *CustomerService) ResetPassword(ctx context.Context, token string, password string) error {
	Customer, err := s.repo.FindByResetToken(ctx, token)

	if err != nil {
		fmt.Println("invalid token")
		return err
	}
	if time.Now().After(Customer.PasswordTokenExpiryTime) {
		fmt.Println("Token expired")
		return errors.New("token has expired")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	Customer.Password = string(hashedPassword)
	Customer.PasswordResetToken = nil
	Customer.PasswordTokenExpiryTime = time.Time{}
	return s.repo.Update(ctx, Customer)
}

func (s *CustomerService) VerifyResetToken(ctx context.Context, token string) error {

	customer, err := s.repo.FindByResetToken(ctx, token)

	if err != nil {
		return err
	}

	if time.Now().After(customer.PasswordTokenExpiryTime) {
		return errors.New("token expired")
	}

	return nil
}
func (s *CustomerService) GetProfile(ctx context.Context, userID uint) (*response.CustomerProfileResponse, error) {
	Customer, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &response.CustomerProfileResponse{
		ID:          Customer.ID,
		FirstName:   Customer.FirstName,
		LastName:    Customer.LastName,
		Email:       Customer.Email,
		PhoneNumber: Customer.PhoneNumber,
		CountryCode: Customer.CountryCode,
		Age:         Customer.Age,
		Gender:      string(Customer.Gender),
		Username:    Customer.Username,
		AvatarURL:   Customer.AvatarURL,
		AvatarType:  string(Customer.AvatarType),
	}, nil
}

func (s *CustomerService) UpdateProfile(ctx context.Context, userID uint, req request.CustomerUpdateProfileRequest) (*response.UpdateProfileResponse, error) {
	customer, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, ae.UnProcessableError("UserNotFound", "user not found", err)
	}

	customer.Age = req.Age
	customer.FirstName = req.FirstName
	customer.LastName = req.LastName
	customer.Email = req.Email

	if err := s.repo.UpdateProfile(ctx, customer); err != nil {
		return nil, err
	}

	return &response.UpdateProfileResponse{
		Message: "Profile updated successfully",
	}, nil
}

func validateName(name, field string) error {
	if strings.TrimSpace(name) == "" {
		return ae.BadRequestError("InvalidName", field+" cannot be empty", nil)
	}
	if strings.Contains(name, " ") {
		return ae.BadRequestError("InvalidName", field+" cannot contain spaces", nil)
	}
	for _, ch := range name {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')) {
			return ae.BadRequestError("InvalidName", field+" must contain alphabets only", nil)
		}
	}
	return nil
}
