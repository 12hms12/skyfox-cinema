package response

import (
	"skyfox/bookings/model"
)

type RevenueResponse struct {
	GrossRevenue float64       `json:"grossRevenue"`
	Shows        []RevenueShow `json:"shows"`
}

type RevenueShow struct {
	Id         int         `json:"id"`
	Date       string    `json:"date"`
	ShowTime   string      `json:"showTime"`
	Movie      model.Movie `json:"movie"`
	Revenue    float64     `json:"revenue"`
	TicketSold uint        `json:"ticketSold"`
}
