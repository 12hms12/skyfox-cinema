package response

import "strings"

type BookingConfirmationResponse struct {
	Id             int     `json:"id"`
	CustomerName   string  `json:"customerName" validate:"required,min=3,max=50,alpha"`
	ShowDate       string  `json:"showDate" validate:"required,datetime=2006-05-20"`
	StartTime      string  `json:"startTime"`
	AmountPaid     float64 `json:"amountPaid" validate:"required,gte=0.0"`
	NoOfSeats      int     `json:"noOfSeats" validate:"required,gte=1"`
	MoviePosterUrl string  `json:"moviePosterUrl"`
	Genre 		   string	`json:"genre"`
	ImdbRating	   string	`json:"imdbRating"`
}

func NewBookingConfirmationResponse(id int, customerName string, showDate string, startTime string, amount float64, noOfSeats int, posterUrl string, genre string, imdbrating string) *BookingConfirmationResponse {
	return &BookingConfirmationResponse{
		Id:             id,
		CustomerName:   customerName,
		ShowDate:       strings.Split(showDate, "T")[0],
		StartTime:      startTime,
		AmountPaid:     amount,
		NoOfSeats:      noOfSeats,
		MoviePosterUrl: posterUrl,
		Genre: 			genre,
		ImdbRating:		imdbrating,
	}
}
