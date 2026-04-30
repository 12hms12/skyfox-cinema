package response

import "skyfox/bookings/model"

type ShowResponse struct {
	Movie model.Movie `json:"movie"`
	Slot  model.Slot  `json:"slot"`
	Id    int         `json:"id"`
	Date  string      `json:"date"`
	Cost  float64     `json:"cost"`
}

func NewShowResponse(movie model.Movie, slot model.Slot, show model.Show) *ShowResponse {
	return &ShowResponse{
		Movie: movie,
		Slot:  slot,
		Id:    show.Id,
		Date:  show.Date.Format("2006-01-02"),
		Cost:  show.Cost,
	}
}
