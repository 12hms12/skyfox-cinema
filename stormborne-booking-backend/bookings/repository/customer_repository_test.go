package repository_test

import (
	"context"
	"skyfox/bookings/model"
	"skyfox/bookings/repository"
	"skyfox/integration_test/db"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomerRepository(t *testing.T) {

	database := db.GetDB()

	err := database.GormDB().AutoMigrate(&model.Customer{})
	assert.NoError(t, err)

	repo := repository.NewCustomerRepository(database)
	ctx := context.Background()

	resetDB := func() {
		database.GormDB().Exec("DELETE FROM customer")
	}

	seedCustomer := model.Customer{
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "john@example.com",
		PhoneNumber: "9999999999",
		CountryCode: "+91",
		Age:         25,
		Gender:      model.MALE,
		Username:    "john_doe",
		Password:    "password",
	}

	t.Run("Create Success", func(t *testing.T) {
		resetDB()

		customer := seedCustomer

		err := repo.Create(ctx, &customer)

		assert.NoError(t, err)
		assert.NotZero(t, customer.ID)
	})

	t.Run("Create Duplicate Email", func(t *testing.T) {
		resetDB()

		customerWithDuplicateEmail := model.Customer{
			FirstName:   "John",
			LastName:    "Doe",
			Email:       "john@example.com",
			PhoneNumber: "9999999998",
			CountryCode: "+91",
			Age:         25,
			Gender:      model.MALE,
			Username:    "john_doe1",
			Password:    "password",
		}
		database.GormDB().Create(&customerWithDuplicateEmail)

		dup := seedCustomer
		dup.Username = "different"

		err := repo.Create(ctx, &dup)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email already exists")
	})

	t.Run("Create Duplicate Phone", func(t *testing.T) {
		resetDB()

		customer := seedCustomer
		database.GormDB().Create(&customer)

		dup := seedCustomer
		dup.Email = "new@example.com"
		dup.Username = "newuser"

		err := repo.Create(ctx, &dup)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "phone number already exists")
	})

	t.Run("Create Duplicate Username", func(t *testing.T) {
		resetDB()

		customer := seedCustomer
		database.GormDB().Create(&customer)

		dup := seedCustomer
		dup.Email = "another@example.com"
		dup.PhoneNumber = "8888888888"

		err := repo.Create(ctx, &dup)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "username already exists")
	})

	t.Run("FindByEmail Success", func(t *testing.T) {
		resetDB()

		database.GormDB().Create(&seedCustomer)

		customer, err := repo.FindByEmail(ctx, "john@example.com")

		assert.NoError(t, err)
		assert.Equal(t, "john@example.com", customer.Email)
	})

	t.Run("FindByEmail Not Found", func(t *testing.T) {
		resetDB()

		customer, err := repo.FindByEmail(ctx, "notfound@example.com")

		assert.Error(t, err)
		assert.Nil(t, customer)
	})

	t.Run("FindByUsername Success", func(t *testing.T) {
		resetDB()

		database.GormDB().Create(&seedCustomer)

		customer, err := repo.FindByUsername(ctx, "john_doe")

		assert.NoError(t, err)
		assert.Equal(t, "john_doe", customer.Username)
	})

	t.Run("FindByUsername Not Found", func(t *testing.T) {
		resetDB()

		customer, err := repo.FindByUsername(ctx, "invalid")

		assert.Error(t, err)
		assert.Nil(t, customer)
	})

	t.Run("FindByResetToken Success", func(t *testing.T) {
		resetDB()

		token := "reset-token"

		customer := seedCustomer
		customer.PasswordResetToken = &token

		database.GormDB().Create(&customer)

		result, err := repo.FindByResetToken(ctx, token)

		assert.NoError(t, err)
		assert.Equal(t, "john@example.com", result.Email)
	})

	t.Run("FindByResetToken NotFound", func(t *testing.T) {
		resetDB()

		customer, err := repo.FindByResetToken(ctx, "invalid")

		assert.Error(t, err)
		assert.Nil(t, customer)
	})

	t.Run("Update Success", func(t *testing.T) {
		resetDB()

		customer := seedCustomer
		database.GormDB().Create(&customer)

		found, err := repo.FindByEmail(ctx, "john@example.com")
		assert.NoError(t, err)

		found.Password = "new-password"
		found.PasswordResetToken = nil

		err = repo.Update(ctx, found)

		assert.NoError(t, err)

		updated, err := repo.FindByEmail(ctx, "john@example.com")
		assert.NoError(t, err)

		assert.Equal(t, "new-password", updated.Password)
		assert.Nil(t, updated.PasswordResetToken)
	})

	t.Run("Update Insert When Not Exist", func(t *testing.T) {

		resetDB()

		customer := seedCustomer

		err := repo.Update(ctx, &customer)

		assert.NoError(t, err)
		assert.NotZero(t, customer.ID)
	})
	t.Run("FindByID Success", func(t *testing.T) {
		resetDB()

		customer := seedCustomer
		err := database.GormDB().Create(&customer).Error
		assert.NoError(t, err)

		foundCustomer, err := repo.FindByID(ctx, customer.ID)

		assert.NoError(t, err)
		assert.NotNil(t, foundCustomer)
		assert.Equal(t, customer.ID, foundCustomer.ID)
		assert.Equal(t, customer.Email, foundCustomer.Email)
		assert.Equal(t, customer.Username, foundCustomer.Username)
	})

	t.Run("FindByID Not Found", func(t *testing.T) {
		resetDB()

		foundCustomer, err := repo.FindByID(ctx, 9999)

		assert.Error(t, err)
		assert.Nil(t, foundCustomer)
		assert.Contains(t, err.Error(), "No customer found with given ID")
	})

}
