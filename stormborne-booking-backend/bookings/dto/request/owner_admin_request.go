package request

import "skyfox/bookings/model"

type OwnerAddAdminRequest struct {
	FirstName   string       `json:"firstName" binding:"required,min=2,max=100"`
	LastName    string       `json:"lastName" binding:"required,min=2,max=100"`
	Email       string       `json:"email" binding:"required,email"`
	PhoneNumber string       `json:"phoneNumber" binding:"required"`
	CountryCode string       `json:"countryCode" binding:"required"`
	Age         uint         `json:"age" binding:"required,gte=1"`
	Gender      model.Gender `json:"gender" binding:"required"`
	Username    string       `json:"username" binding:"required,min=3,max=50"`
	Password    string       `json:"password" binding:"required,min=6,max=100"`
}
