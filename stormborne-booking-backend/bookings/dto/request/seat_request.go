package request

type SeatStatusRequest struct {
    ShowID uint `uri:"showId" binding:"required"`
}