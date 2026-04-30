package testdata

import (
	"skyfox/bookings/model"
	"skyfox/common/middleware/security"
	"skyfox/config"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var HashedPasswordForDummyUsers string
var DummyJwtManager *security.JwtManager

var DummyUsersPassword string = "foobar"

func GetTestJwtToken(customer model.Customer) string {
	token, err := DummyJwtManager.GenerateToken(int64(customer.ID),security.Role(customer.UserRole))
	if err != nil {
		panic("err creating test token")
	}
	return token;
}

var DummyUsers = []model.Customer{
	{
		ID:          1,
		FirstName:   "Johny",
		LastName:    "Doe",
		Email:       "johny@example.com",
		PhoneNumber: "9999999999",
		CountryCode: "+91",
		Age:         25,
		Gender:      model.MALE,
		Username:    "johny_doe",
		Password:    HashedPasswordForDummyUsers,
		AvatarType:  model.PREDEFINED,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		UserRole:    "user",
	},
	{
		ID:          2,
		FirstName:   "Jane",
		LastName:    "Smith",
		Email:       "jane@example.com",
		PhoneNumber: "8888888888",
		CountryCode: "+91",
		Age:         28,
		Gender:      model.FEMALE,
		Username:    "jane_smith",
		Password:    HashedPasswordForDummyUsers,
		AvatarType:  model.UPLOADED,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		UserRole:    "user",
	},
}

func init() {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(DummyUsersPassword), bcrypt.DefaultCost)
	HashedPasswordForDummyUsers = string(hashedPassword)
	DummyJwtManager = security.NewJwtManager(config.TokenConfig{
        TTL:    1800,
        SECRET: "test",
    })
}