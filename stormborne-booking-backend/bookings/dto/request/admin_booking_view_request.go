package request

type AdminBookingViewRequest struct {
	StartDate string `form:"startDate"`
	EndDate   string `form:"endDate"`
	MovieName string `form:"movieName"`
}
