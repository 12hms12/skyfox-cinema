package response

type WeeklyScheduleResponse struct {
	Schedule []ScheduleItem `json:"schedule"`
}

type ScheduleItem struct {
	ID       int                 `json:"id"`
	Date     string              `json:"date"`
	ScreenID int                 `json:"screenId"`
	SlotID   int                 `json:"slotId"`
	Cost     float64             `json:"cost"`
	Movie    ScheduledMovieBrief `json:"movie"`
}

type ScreenOption struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ScheduledMovieBrief struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ScheduleShowResponse struct {
	ID       int    `json:"id"`
	MovieID  string `json:"movieId"`
	ScreenID int    `json:"screenId"`
	SlotID   int    `json:"slotId"`
	Date     string `json:"date"`
	Cost     float64 `json:"cost"`
	Rated    string `json:"rated"`
}
