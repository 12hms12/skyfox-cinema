package repository

import (
	"context"
	"errors"
	"fmt"
	"skyfox/bookings/database/common"
	"skyfox/bookings/model"
	ae "skyfox/error"
	"strings"

	"gorm.io/gorm"
)

type customerRepository struct {
	*common.BaseDB
}

func NewCustomerRepository(db *common.BaseDB) *customerRepository {
	return &customerRepository{
		BaseDB: db,
	}
}

func (repo customerRepository) Create(ctx context.Context, c *model.Customer) error {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	result := dbCtx.Create(c)

	if result.Error != nil {

		errMsg := result.Error.Error()
		if strings.Contains(errMsg, "SQLSTATE 23505") {

			msg := "duplicate record"

			if strings.Contains(errMsg, "uni_customer_email") {
				msg = "email already exists"
			}

			if strings.Contains(errMsg, "uni_customer_phone_number") {
				msg = "phone number already exists"
			}

			if strings.Contains(errMsg, "uni_customer_username") {
				msg = "username already exists"
			}

			return ae.UnProcessableError(
				"DuplicateRecord",
				msg,
				result.Error,
			)
		}

		return ae.InternalServerError(
			"CustomerCreationFailed",
			"failed to create online customer",
			result.Error,
		)
	}

	return nil
}

func (repo customerRepository) FindByUsername(ctx context.Context, username string) (*model.Customer, error) {
	var customer model.Customer

	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	if result := dbCtx.Where("username = ?", username).First(&customer); result.Error != nil {
		fmt.Println("DB error", result.Error)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {

			return nil, ae.NotFoundError("CustomerNotFound", "No customer found with given username", result.Error)
		}
		return nil, ae.UnProcessableError("CustomerFetchFailed", "Failed to fetch customer by username", result.Error)
	}
	return &customer, nil

}

func (repo *customerRepository) FindByEmail(ctx context.Context, email string) (*model.Customer, error) {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	var customer model.Customer
	result := dbCtx.Where("email = ?", email).First(&customer)
	if result.Error != nil {
		fmt.Println("DB error", result.Error)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {

			return nil, ae.NotFoundError("CustomerNotFound", "No customer found with given email", result.Error)
		}
		return nil, ae.UnProcessableError("CustomerFetchFailed", "Failed to fetch customer by email", result.Error)
	}
	return &customer, nil
}

func (repo *customerRepository) FindByResetToken(ctx context.Context, token string) (*model.Customer, error) {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	var customer model.Customer
	result := dbCtx.Where("password_reset_token = ?", token).First(&customer)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {

			return nil, ae.NotFoundError("TokenNotFound", "No customer found with given token", result.Error)
		}
		return nil, ae.UnProcessableError("TokenFetchFailed", "Failed to fetch customer by token", result.Error)
	}
	return &customer, nil
}

func (repo *customerRepository) Update(ctx context.Context, customer *model.Customer) error {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()
	return dbCtx.Save(customer).Error
}
func (repo *customerRepository) FindByID(ctx context.Context, userID uint) (*model.Customer, error) {
	var customer model.Customer

	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	if result := dbCtx.First(&customer, userID); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ae.NotFoundError("CustomerNotFound", "No customer found with given ID", result.Error)
		}
		return nil, ae.UnProcessableError("CustomerFetchFailed", "Failed to fetch customer by ID", result.Error)
	}
	return &customer, nil
}

func (repo *customerRepository) FindAllByRole(ctx context.Context, role string) ([]model.Customer, error) {
	var customers []model.Customer

	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	if result := dbCtx.Where("user_role = ?", role).Order("id ASC").Find(&customers); result.Error != nil {
		return nil, ae.UnProcessableError("CustomerFetchFailed", "Failed to fetch customers by role", result.Error)
	}

	return customers, nil
}

func (repo *customerRepository) DeleteByID(ctx context.Context, userID uint) error {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	result := dbCtx.Delete(&model.Customer{}, userID)
	if result.Error != nil {
		return ae.UnProcessableError("CustomerDeleteFailed", "Failed to delete customer", result.Error)
	}

	if result.RowsAffected == 0 {
		return ae.NotFoundError("CustomerNotFound", "No customer found with given ID", gorm.ErrRecordNotFound)
	}

	return nil
}


func (repo *customerRepository) UpdateProfile(ctx context.Context, customer *model.Customer) error {
    dbCtx, cancel := repo.WithContext(ctx)
    defer cancel()

    result := dbCtx.Model(customer).Updates(customer)
	
    if result.Error != nil {
        return ae.InternalServerError("ProfileUpdateFailed", "failed to update profile", result.Error)
    }
    return nil
}