package controller

import (
	"context"
	"net/http"
	"skyfox/bookings/dto/response"
	"skyfox/bookings/repository"
	"skyfox/common/logger"
	ae "skyfox/error"
	"time"

	"github.com/gin-gonic/gin"
)

type RevenueService interface {
	RevenueBy(ctx context.Context, revenueQuery *repository.RevenueQuery) (*response.RevenueResponse, error)
}

type RevenueController struct {
	revenueService RevenueService
}

func NewRevenueController(revenueService RevenueService) *RevenueController {
	return &RevenueController{
		revenueService: revenueService,
	}
}

// Revenue godoc
//
//	@Summary		Revenue
//	@Description	Get revenue by date or movie ID
//	@Tags			Revenue
//	@Accept			json
//	@Produce		json
//	@param          Authorization header    string  true    "Enter basic auth"
//	@Param			startDate	  query		string	false	"filter revenue by start date"
//	@Param			endDate	  	  query		string	false	"filter revenue by end date"
//	@Param			movieId	      query		string	false	"filter revenue by movie ID"
//	@Param			genre	      query		string	false	"filter revenue by genre"
//	@Param			showTime	      query		string	false	"filter revenue by show time"
//	@Success 		200 		  {object} 	response.RevenueResponse
//	@Failure		400		      {object}	ae.AppError
//	@Router			/revenue [get]
func (rh *RevenueController) GetRevenue(c *gin.Context) {

	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	movieId := c.Query("movieId")
	genre := c.Query("genre")
	showTime := c.Query("showTime")
	layout := "2006-01-02"

	parsedStartDate, _ := time.Parse(layout, startDate)
	parsedEndDate, _ := time.Parse(layout, endDate)

	revenue, responseError := rh.revenueService.RevenueBy(c.Request.Context(), &repository.RevenueQuery{
		StartDate: &parsedStartDate,
		EndDate:   &parsedEndDate,
		MovieId:   movieId,
		Genre:     genre,
		ShowTime: showTime,
	})

	if responseError != nil {
		err := responseError.(*ae.AppError)
		logger.Error("%s", err.UnWrap().Error())
		c.AbortWithStatusJSON(err.HTTPCode(), err)
		return
	}

	c.JSON(http.StatusOK, revenue)
}
