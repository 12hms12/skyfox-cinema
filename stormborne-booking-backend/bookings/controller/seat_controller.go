package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"skyfox/bookings/common"
	"skyfox/bookings/dto/request"
	"skyfox/bookings/service"
)

type SeatController struct {
	seatService service.SeatService
}

func NewSeatController(seatService service.SeatService) *SeatController {
	return &SeatController{seatService: seatService}
}

// GET /shows/:showId/seats
// GetSeatStatus godoc
// @Summary      Get seat map for a show
// @Description  Returns all seats with availability status and pricing for a given show. Weekend shows (Sat/Sun) have a ₹50 surcharge applied to base price.
// @Tags         seats
// @Accept       json
// @Produce      json
// @Param        showId  path      int  true  "Show ID"  example(1)
// @Success      200     {object}  response.SeatStatusResponse
// @Failure      400     {object}  map[string]string  "Invalid showId"
// @Failure      404     {object}  map[string]string  "Feature not available"
// @Failure      500     {object}  map[string]string  "Internal server error"
// @Router       /shows/{showId}/seats [get]
func (sc *SeatController) GetSeatStatus(ctx *gin.Context) {
	if !common.Feature_SeatMap {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "feature not available"})
		return
	}

	var req request.SeatStatusRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := sc.seatService.GetSeatStatus(ctx.Request.Context(), req.ShowID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}