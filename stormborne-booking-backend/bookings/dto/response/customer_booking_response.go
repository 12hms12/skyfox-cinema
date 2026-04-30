package response

import "time"

type BookingSeatInfo struct {
	SeatID   int     `json:"seatId"`
	Label    string  `json:"label"`
	SeatType string  `json:"seatType"`
	Price    float64 `json:"price"`
}

type CreateCustomerBookingResponse struct {
	BookingID        int               `json:"bookingId"`
	BookingReference string            `json:"bookingReference"`
	Seats            []BookingSeatInfo `json:"seats"`
	TotalPrice       float64           `json:"totalPrice"`
	Status           string            `json:"status"`
	ExpiresAt        time.Time         `json:"expiresAt"`
}

type PaymentResponse struct {
	PaymentID string  `json:"paymentId"`
	BookingID int     `json:"bookingId"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Status    string  `json:"status"`
}

type BookingDetailsResponse struct {
	BookingID        int               `json:"bookingId"`
	BookingReference string            `json:"bookingReference"`
	MovieID          string            `json:"movieId"`
	ShowTime         string            `json:"showTime"`
	ScreenName       string            `json:"screenName"`
	Seats            []BookingSeatInfo `json:"seats"`
	TotalPrice       float64           `json:"totalPrice"`
	PaymentStatus    string            `json:"paymentStatus"`
	BookingStatus    string            `json:"bookingStatus"`
}
