package response

type AdminBookingViewItem struct {
	ID               int       `json:"id"`
	TransactionID    string    `json:"transactionId"`
	OnlineCustomerID uint      `json:"onlineCustomerId"`
	ShowID           int       `json:"showId"`
	MovieName        string    `json:"movieName"`
	ShowDate         string    `json:"showDate"`
	StartTime        string    `json:"startTime"`
	TotalPrice       float64   `json:"totalPrice"`
	Status           string    `json:"status"`
}

type AdminBookingViewResponse struct {
	Bookings []AdminBookingViewItem `json:"bookings"`
	Total    int                    `json:"total"`
}

func NewAdminBookingViewResponse(bookings []AdminBookingViewItem) *AdminBookingViewResponse {
	return &AdminBookingViewResponse{
		Bookings: bookings,
		Total:    len(bookings),
	}
}