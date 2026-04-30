package request

type ScheduleShowRequest struct {
	MovieID  string `json:"movieId" binding:"required"`
	ScreenID int    `json:"screenId" binding:"required,gt=0"`
	SlotID   int    `json:"slotId" binding:"required,gt=0"`
	Date     string `json:"date" binding:"required"`
	Cost     float64 `json:"cost" binding:"required,gt=0"`
	Rated    string `json:"rated"`
}

type AddScreenRequest struct {
	Name string `json:"name"`
}
