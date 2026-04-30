package request

import (
	"skyfox/bookings/model"
)

type CustomerSignupRequest struct {
	FirstName       string       `json:"firstName" binding:"required,min=2,max=100"`
	LastName        string       `json:"lastName" binding:"required,min=2,max=100"`
	Email           string       `json:"email" binding:"required,email"`
	PhoneNumber     string       `json:"phoneNumber" binding:"required"`
	CountryCode     string       `json:"countryCode" binding:"required"`
	Age             uint         `json:"age" binding:"required,gte=1"`
	Gender          model.Gender `json:"gender" binding:"required"`
	Username        string       `json:"username" binding:"required,min=3,max=50"`
	Password        string       `json:"password" binding:"required,min=6,max=100"`
	CaptchaID       string       `json:"captcha_id" binding:"required"`
	SliderPositionX float64      `json:"slider_position_x" binding:"required"`
}

type CustomerLoginRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password" binding:"required"`
}

type CustomerBookingViewRequest struct {
	StartDate        string `form:"startDate"`
	EndDate          string `form:"endDate"`
	MovieName        string `form:"movieName"`
	OnlineCustomerID uint   `form:"onlineCustomerId"`
}

type CustomerUpdateProfileRequest struct {
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
    Email     string `json:"email"`
    Age       uint   `json:"age"`
}