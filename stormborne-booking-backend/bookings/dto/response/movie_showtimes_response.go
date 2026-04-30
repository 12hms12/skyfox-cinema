package response

type ShowtimeResponse struct {
	ShowId         int     `json:"showId"`
	ScreenId       int     `json:"screenId"`
	ScreenName     string  `json:"screenName"`
	SlotId         int     `json:"slotId"`
	SlotName       string  `json:"slotName"`
	StartTime      string  `json:"startTime"`
	EndTime        string  `json:"endTime"`
	Cost           float64 `json:"cost"`
	AvailableSeats int64   `json:"availableSeats"`
}
