package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"skyfox/bookings/dto/request"
	"skyfox/bookings/dto/response"
	"skyfox/bookings/model"
	"skyfox/bookings/service"
	repomocks "skyfox/bookings/service/mocks"
	serviceMocks "skyfox/bookings/controller/mocks"
	"skyfox/common/middleware/security"
	"skyfox/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

var serverConf = config.ServerConfig{
	Port:                          8080,
	ReadTimeout:                   5,
	WriteTimeout:                  5,
	GineMode:                      "debug",
	WebClientResetPasswordBaseUrl: "http://localhost:3000/reset-password?token=",
}

func TestSignup_Success(t *testing.T) {
	ctx := context.Background()

	repo := repomocks.NewMockCustomerRepository(t)
	avatarSvc := serviceMocks.NewMockAvatarService(t)
	otpChecker := repomocks.NewMockPhoneOTPChecker(t)

	jwt := security.NewJwtManager(config.TokenConfig{
		SECRET: "test-secret",
		TTL:    3600,
	})

	svc := service.NewCustomerService(repo, jwt, avatarSvc, &serverConf, otpChecker, nil)

	req := request.CustomerSignupRequest{
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "john@example.com",
		PhoneNumber: "+919876543210",
		CountryCode: "IN",
		Age:         25,
		Gender:      "Male",
		Username:    "john123",
		Password:    "password123",
	}

	otpChecker.On("IsRecipientVerified", ctx, "IN+919876543210", "SMS").Return(true, nil)

	repo.EXPECT().
		Create(ctx, mock.AnythingOfType("*model.Customer")).
		Run(func(ctx context.Context, c *model.Customer) {
			assert.True(t, c.IsPhoneVerified)
			c.ID = 1
		}).
		Return(nil)

	avatarSvc.EXPECT().
		AssignRandomAvatar(ctx, uint(1), "Male").
		Return(&response.AvatarResponse{
			UserID:     1,
			AvatarURL:  "https://example.com/avatars/male1.png",
			AvatarType: "predefined",
		}, nil)

	res, err := svc.Signup(ctx, req)

	assert.NoError(t, err)
	assert.Nil(t, res)
}

func TestSignup_InvalidEmail(t *testing.T) {
	ctx := context.Background()

	repo := repomocks.NewMockCustomerRepository(t)
	avatarSvc := serviceMocks.NewMockAvatarService(t)

	jwt := security.NewJwtManager(config.TokenConfig{
		SECRET: "test-secret",
		TTL:    3600,
	})

	svc := service.NewCustomerService(repo, jwt, avatarSvc, &serverConf, nil, nil)

	req := request.CustomerSignupRequest{
		Email:       "invalid-email",
		PhoneNumber: "+919876543210",
		CountryCode: "IN",
		Password:    "password",
	}

	res, err := svc.Signup(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestSignup_InvalidPhone(t *testing.T) {
	ctx := context.Background()

	repo := repomocks.NewMockCustomerRepository(t)
	avatarSvc := serviceMocks.NewMockAvatarService(t)

	jwt := security.NewJwtManager(config.TokenConfig{
		SECRET: "test-secret",
		TTL:    3600,
	})

	svc := service.NewCustomerService(repo, jwt, avatarSvc, &serverConf, nil, nil)

	req := request.CustomerSignupRequest{
		Email:       "john@example.com",
		PhoneNumber: "12345",
		CountryCode: "IN",
		Password:    "password",
	}

	res, err := svc.Signup(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestSignup_RepoError(t *testing.T) {
	ctx := context.Background()

	repo := repomocks.NewMockCustomerRepository(t)
	avatarSvc := serviceMocks.NewMockAvatarService(t)
	otpChecker := repomocks.NewMockPhoneOTPChecker(t)

	jwt := security.NewJwtManager(config.TokenConfig{
		SECRET: "test-secret",
		TTL:    3600,
	})

	svc := service.NewCustomerService(repo, jwt, avatarSvc, &serverConf, otpChecker, nil)

	req := request.CustomerSignupRequest{
		Email:       "john@example.com",
		PhoneNumber: "+919876543210",
		CountryCode: "IN",
		Password:    "password",
	}

	otpChecker.On("IsRecipientVerified", ctx, "IN+919876543210", "SMS").Return(false, nil)

	repo.EXPECT().
		Create(ctx, mock.Anything).
		Return(errors.New("db error"))

	res, err := svc.Signup(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestLogin_Success(t *testing.T) {
	ctx := context.Background()

	repo := repomocks.NewMockCustomerRepository(t)
	avatarSvc := serviceMocks.NewMockAvatarService(t)

	jwt := security.NewJwtManager(config.TokenConfig{
		SECRET: "test-secret",
		TTL:    3600,
	})

	svc := service.NewCustomerService(repo, jwt, avatarSvc, &serverConf, nil, nil)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)

	customer := model.Customer{
		ID:         1,
		Username:   "john123",
		Email:      "john@example.com",
		Password:   string(hashed),
		AvatarURL:  "url",
		AvatarType: "png",
		Gender:     "Male",
		Age:        25,
	}

	repo.EXPECT().
		FindByEmail(ctx, customer.Email).
		Return(&customer, nil)

	req := request.CustomerLoginRequest{
		Email:    "john@example.com",
		Password: "pass",
	}

	res, err := svc.Login(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, res.Token)
	assert.Equal(t, "john123", res.Username)
	assert.Equal(t, "url", res.AvatarURL)
	assert.EqualValues(t, "Male", res.Gender)
	assert.EqualValues(t, 25, res.Age)
}

func TestCustomerService(t *testing.T) {
	ctx := context.Background()

	jwt := security.NewJwtManager(config.TokenConfig{
		SECRET: "test-secret",
		TTL:    3600,
	})

	avatarSvc := serviceMocks.NewMockAvatarService(t)

	t.Run("ForgotPassword", func(t *testing.T) {

		t.Run("Success", func(t *testing.T) {
			repo := repomocks.NewMockCustomerRepository(t)

			customer := &model.Customer{Email: "john@example.com"}

			repo.On("FindByEmail", mock.Anything, "john@example.com").Return(customer, nil).Once()
			repo.On("Update", mock.Anything, mock.AnythingOfType("*model.Customer")).Return(nil).Once()

			service := service.NewCustomerService(repo, jwt, avatarSvc, &serverConf, nil, nil)

			resetLink, err := service.ForgotPassword(ctx, "john@example.com")

			assert.NoError(t, err)
			assert.NotEmpty(t, resetLink)
			assert.Contains(t, resetLink, "reset-password?token=")
			assert.NotNil(t, customer.PasswordResetToken)
			assert.WithinDuration(t, time.Now().Add(30*time.Minute), customer.PasswordTokenExpiryTime, time.Minute)

			repo.AssertExpectations(t)
		})

		t.Run("CustomerNotFound", func(t *testing.T) {
			repo := repomocks.NewMockCustomerRepository(t)
			service := service.NewCustomerService(repo, jwt, avatarSvc, &serverConf, nil, nil)

			repo.On("FindByEmail", mock.Anything, "missing@example.com").Return(nil, errors.New("email not found")).Once()

			resetLink, err := service.ForgotPassword(ctx, "missing@example.com")

			assert.Error(t, err)
			assert.Empty(t, resetLink)

			repo.AssertExpectations(t)
		})

		t.Run("UpdateFails", func(t *testing.T) {
			repo := repomocks.NewMockCustomerRepository(t)
			service := service.NewCustomerService(repo, jwt, avatarSvc, &serverConf, nil, nil)

			customer := &model.Customer{Email: "john@example.com"}

			repo.On("FindByEmail", mock.Anything, "john@example.com").Return(customer, nil).Once()
			repo.On("Update", mock.Anything, mock.AnythingOfType("*model.Customer")).Return(errors.New("update failed")).Once()

			resetLink, err := service.ForgotPassword(ctx, "john@example.com")

			assert.Error(t, err)
			assert.Empty(t, resetLink)

			repo.AssertExpectations(t)
		})

		t.Run("ResendLimitExceeded", func(t *testing.T) {
			repo := repomocks.NewMockCustomerRepository(t)
			service := service.NewCustomerService(repo, jwt, avatarSvc, &serverConf, nil, nil)

			customer := &model.Customer{
				Email: "john.doe@example.com",
			}

			repo.On("FindByEmail", mock.Anything, "john.doe@example.com").Return(customer, nil).Times(4)
			repo.On("Update", mock.Anything, mock.AnythingOfType("*model.Customer")).Return(nil).Times(3)

			link1, err1 := service.ForgotPassword(ctx, "john.doe@example.com")
			assert.NoError(t, err1)
			assert.NotEmpty(t, link1)

			link2, err2 := service.ForgotPassword(ctx, "john.doe@example.com")
			assert.NoError(t, err2)
			assert.NotEmpty(t, link2)

			link3, err3 := service.ForgotPassword(ctx, "john.doe@example.com")
			assert.NoError(t, err3)
			assert.NotEmpty(t, link3)

			link4, err4 := service.ForgotPassword(ctx, "john.doe@example.com")
			assert.Error(t, err4)
			assert.Empty(t, link4)
			assert.Contains(t, err4.Error(), "resend limit exceeded")

			repo.AssertExpectations(t)
		})
	})

	t.Run("ResetPassword", func(t *testing.T) {

		token := "reset-token-123"

		t.Run("Success", func(t *testing.T) {
			repo := repomocks.NewMockCustomerRepository(t)
			service := service.NewCustomerService(repo, jwt, avatarSvc, &serverConf, nil, nil)

			customer := &model.Customer{
				Email:                   "john@example.com",
				PasswordResetToken:      &token,
				PasswordTokenExpiryTime: time.Now().Add(30 * time.Minute),
			}

			repo.On("FindByResetToken", mock.Anything, token).Return(customer, nil).Once()
			repo.On("Update", mock.Anything, mock.AnythingOfType("*model.Customer")).Return(nil).Once()

			err := service.ResetPassword(ctx, token, "new-password")

			assert.NoError(t, err)

			err = bcrypt.CompareHashAndPassword(
				[]byte(customer.Password),
				[]byte("new-password"),
			)

			assert.NoError(t, err)
			assert.Nil(t, customer.PasswordResetToken)
			assert.True(t, customer.PasswordTokenExpiryTime.IsZero())

			repo.AssertExpectations(t)
		})

		t.Run("TokenNotFound", func(t *testing.T) {
			repo := repomocks.NewMockCustomerRepository(t)
			service := service.NewCustomerService(repo, jwt, avatarSvc, &serverConf, nil, nil)

			repo.On("FindByResetToken", mock.Anything, token).Return(nil, errors.New("token not found")).Once()

			err := service.ResetPassword(ctx, token, "new-password")

			assert.Error(t, err)

			repo.AssertExpectations(t)
		})

		t.Run("TokenExpired", func(t *testing.T) {
			repo := repomocks.NewMockCustomerRepository(t)
			service := service.NewCustomerService(repo, jwt, avatarSvc, &serverConf, nil, nil)

			customer := &model.Customer{
				Email:                   "john@example.com",
				PasswordResetToken:      &token,
				PasswordTokenExpiryTime: time.Now().Add(-1 * time.Hour),
			}

			repo.On("FindByResetToken", mock.Anything, token).Return(customer, nil).Once()

			err := service.ResetPassword(ctx, token, "new-password")

			assert.Error(t, err)

			repo.AssertExpectations(t)
		})

		t.Run("UpdateFails", func(t *testing.T) {
			repo := repomocks.NewMockCustomerRepository(t)
			service := service.NewCustomerService(repo, jwt, avatarSvc, &serverConf, nil, nil)

			customer := &model.Customer{
				Email:                   "john@example.com",
				PasswordResetToken:      &token,
				PasswordTokenExpiryTime: time.Now().Add(30 * time.Minute),
			}

			repo.On("FindByResetToken", mock.Anything, token).Return(customer, nil).Once()
			repo.On("Update", mock.Anything, mock.AnythingOfType("*model.Customer")).Return(errors.New("update failed")).Once()

			err := service.ResetPassword(ctx, token, "new-password")

			assert.Error(t, err)

			repo.AssertExpectations(t)
		})
	})
}