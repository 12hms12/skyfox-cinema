package response

type SeatResponse struct {
	SeatID    uint    `json:"seatId"`
	Label     string  `json:"label"`     // e.g. "A1", "B10"
	Row       string  `json:"row"`       // e.g. "A"
	Column    int     `json:"column"`    // e.g. 1
	SeatType  string  `json:"seatType"`  // "REGULAR", "PREMIUM"
	Status    string  `json:"status"`    // "available" | "sold"
	BasePrice float64 `json:"basePrice"` // includes weekend surcharge if applicable
	IsWeekend bool    `json:"isWeekend"`
}

type SeatStatusResponse struct {
	ShowID   uint           `json:"showId"`
	ShowDate string         `json:"showDate"`
	Seats    []SeatResponse `json:"seats"`
}