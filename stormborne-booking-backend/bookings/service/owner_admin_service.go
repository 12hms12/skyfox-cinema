package service

import (
	"context"
	"net/http"
	// "strings"

	"skyfox/bookings/dto/request"
	"skyfox/bookings/dto/response"
	"skyfox/bookings/model"
	"skyfox/common/middleware/security"
	"skyfox/common/middleware/validator"
	ae "skyfox/error"

	// "github.com/nyaruka/phonenumbers"
	"golang.org/x/crypto/bcrypt"
)

type OwnerAdminRepository interface {
	Create(context.Context, *model.Customer) error
	FindByUsername(context.Context, string) (*model.Customer, error)
	FindByEmail(context.Context, string) (*model.Customer, error)
	FindByID(context.Context, uint) (*model.Customer, error)
	FindAllByRole(context.Context, string) ([]model.Customer, error)
	DeleteByID(context.Context, uint) error
}

type OwnerAdminService struct {
	repo OwnerAdminRepository
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	appErr, ok := err.(*ae.AppError)
	if !ok {
		return false
	}
	return appErr.HTTPCode() == http.StatusNotFound
}

func NewOwnerAdminService(repo OwnerAdminRepository) *OwnerAdminService {
	return &OwnerAdminService{repo: repo}
}

func (s *OwnerAdminService) ListAdmins(ctx context.Context) (*response.OwnerAdminsListResponse, error) {
	admins, err := s.repo.FindAllByRole(ctx, string(security.ADMIN))
	if err != nil {
		return nil, err
	}

	items := make([]response.OwnerAdminItemResponse, 0, len(admins))
	for _, a := range admins {
		items = append(items, response.OwnerAdminItemResponse{
			ID:          a.ID,
			FirstName:   a.FirstName,
			LastName:    a.LastName,
			Email:       a.Email,
			PhoneNumber: a.PhoneNumber,
			CountryCode: a.CountryCode,
			Username:    a.Username,
			Age:         a.Age,
			Gender:      string(a.Gender),
			Role:        a.UserRole,
		})
	}

	return &response.OwnerAdminsListResponse{Admins: items}, nil
}

func (s *OwnerAdminService) AddAdmin(ctx context.Context, req request.OwnerAddAdminRequest) error {
	if !validator.ValidateEmail(req.Email) {
		return ae.BadRequestError("InvalidEmail", "invalid email", nil)
	}

	// parsedNumber, err := phonenumbers.Parse(req.PhoneNumber, strings.ToUpper(req.CountryCode))
	// if err != nil || !phonenumbers.IsValidNumber(parsedNumber) {
	// 	return ae.BadRequestError("InvalidPhoneNumber", "invalid phone number", err)
	// }

	if existingByUsername, err := s.repo.FindByUsername(ctx, req.Username); err == nil && existingByUsername != nil {
		return ae.UnProcessableError("DuplicateRecord", "username already exists", nil)
	} else if err != nil && !isNotFoundError(err) {
		return err
	}
	if existingByEmail, err := s.repo.FindByEmail(ctx, req.Email); err == nil && existingByEmail != nil {
		return ae.UnProcessableError("DuplicateRecord", "email already exists", nil)
	} else if err != nil && !isNotFoundError(err) {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return ae.InternalServerError("PasswordHashFailed", "failed to hash password", err)
	}

	admin := &model.Customer{
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		Email:           req.Email,
		PhoneNumber:     req.PhoneNumber,
		CountryCode:     req.CountryCode,
		Age:             req.Age,
		Gender:          req.Gender,
		Username:        req.Username,
		Password:        string(hashedPassword),
		UserRole:        string(security.ADMIN),
		IsEmailVerified: true,
		IsPhoneVerified: true,
	}

	if err := s.repo.Create(ctx, admin); err != nil {
		return err
	}

	return nil
}

func (s *OwnerAdminService) RemoveAdmin(ctx context.Context, adminID uint) error {
	admin, err := s.repo.FindByID(ctx, adminID)
	if err != nil {
		return err
	}

	if admin.UserRole != string(security.ADMIN) {
		return ae.NewAppError(http.StatusForbidden, "Forbidden", "only admin users can be removed", nil)
	}

	return s.repo.DeleteByID(ctx, adminID)
}
