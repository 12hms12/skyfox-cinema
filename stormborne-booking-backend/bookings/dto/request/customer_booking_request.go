package request

type CreateCustomerBookingRequest struct {
	ShowID  int   `json:"showId" binding:"required,gte=1"`
	SeatIDs []int `json:"seatIds" binding:"required,min=1"`
}

type PaymentRequest struct {
	Success bool `json:"success"`
}
