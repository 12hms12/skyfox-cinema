package controller

import (
	"context"
	"net/http"
	"strconv"

	"skyfox/bookings/dto/request"
	"skyfox/bookings/dto/response"
	"skyfox/bookings/model"
	"skyfox/common/logger"
	"skyfox/common/middleware/validator"
	ae "skyfox/error"

	"github.com/gin-gonic/gin"
)

type AdminScheduleService interface {
	GetWeeklySchedule(ctx context.Context, startDate string) (*response.WeeklyScheduleResponse, error)
	GetSlots(ctx context.Context) ([]model.Slot, error)
	ScheduleShow(ctx context.Context, req request.ScheduleShowRequest) (*response.ScheduleShowResponse, error)
	DeleteShow(ctx context.Context, showID int) error
	GetScreens(ctx context.Context) ([]response.ScreenOption, error)
	AddScreen(ctx context.Context, req request.AddScreenRequest) (*response.ScreenOption, error)
}

type adminScheduleController struct {
	service AdminScheduleService
}

func NewAdminScheduleController(service AdminScheduleService) *adminScheduleController {
	return &adminScheduleController{service: service}
}

// GetWeeklySchedule godoc
//
//	@Summary		Weekly Admin Schedule
//	@Description	get weekly schedule by start date
//	@Tags			Admin-Schedule
//	@Accept			json
//	@Produce		json
//	@security		BasicAuth
//	@param Authorization header string true "Enter basic auth"
//	@Param startDate query string true "week start date in YYYY-MM-DD"
//	@Success		200	{object}	response.WeeklyScheduleResponse
//	@Failure		400	{object}	ae.AppError
//	@Failure		404	{object}	ae.AppError
//	@Failure		500	{object}	ae.AppError
//	@Router			/admin/schedule/week [get]
func (ac *adminScheduleController) GetWeeklySchedule(c *gin.Context) {
	startDate := c.Query("startDate")
	if startDate == "" {
		err := ae.BadRequestError("MissingStartDate", "startDate query parameter is required", nil)
		c.AbortWithStatusJSON(err.HTTPCode(), err)
		return
	}

	resp, responseError := ac.service.GetWeeklySchedule(c.Request.Context(), startDate)
	if responseError != nil {
		err := responseError.(*ae.AppError)
		logger.Error("%s", err.UnWrap().Error())
		c.AbortWithStatusJSON(err.HTTPCode(), err)
		return
	}

	c.IndentedJSON(http.StatusOK, resp)
}

// GetSlots godoc
//
//	@Summary		Admin Slots
//	@Description	get all slots
//	@Tags			Admin-Schedule
//	@Accept			json
//	@Produce		json
//	@security		BasicAuth
//	@param Authorization header string true "Enter basic auth"
//	@Success		200	{array}		model.Slot
//	@Failure		500	{object}	ae.AppError
//	@Router			/admin/slots [get]
func (ac *adminScheduleController) GetSlots(c *gin.Context) {
	slots, responseError := ac.service.GetSlots(c.Request.Context())
	if responseError != nil {
		err := responseError.(*ae.AppError)
		logger.Error("%s", err.UnWrap().Error())
		c.AbortWithStatusJSON(err.HTTPCode(), err)
		return
	}

	c.IndentedJSON(http.StatusOK, slots)
}

// GetScreens godoc
//
//	@Summary		Admin Screens
//	@Description	get all screens
//	@Tags			Admin-Schedule
//	@Accept			json
//	@Produce		json
//	@security		BasicAuth
//	@param Authorization header string true "Enter basic auth"
//	@Success		200	{array}		response.ScreenOption
//	@Failure		500	{object}	ae.AppError
//	@Router			/admin/screens [get]
func (ac *adminScheduleController) GetScreens(c *gin.Context) {
	screens, responseError := ac.service.GetScreens(c.Request.Context())
	if responseError != nil {
		err := responseError.(*ae.AppError)
		logger.Error("%s", err.UnWrap().Error())
		c.AbortWithStatusJSON(err.HTTPCode(), err)
		return
	}

	c.IndentedJSON(http.StatusOK, screens)
}


// AddScreen godoc
//
//	@Summary		Add Screen
//	@Description	create a new screen for scheduling
//	@Tags			Admin-Schedule
//	@Accept			json
//	@Produce		json
//	@security		BasicAuth
//	@param Authorization header string true "Enter basic auth"
//	@Param request body request.AddScreenRequest false "Add screen request"
//	@Success		201	{object}	response.ScreenOption
//	@Failure		400	{object}	ae.AppError
//	@Failure		500	{object}	ae.AppError
//	@Router			/admin/screens [post]
func (ac *adminScheduleController) AddScreen(c *gin.Context) {
	req := request.AddScreenRequest{}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Error("bad request %v", validator.HandleStructValidationError(err))
			c.AbortWithStatusJSON(http.StatusBadRequest, validator.HandleStructValidationError(err))
			return
		}
	}

	screen, responseError := ac.service.AddScreen(c.Request.Context(), req)
	if responseError != nil {
		err := responseError.(*ae.AppError)
		logger.Error("%s", err.UnWrap().Error())
		c.AbortWithStatusJSON(err.HTTPCode(), err)
		return
	}

	c.IndentedJSON(http.StatusCreated, screen)
}

// ScheduleShow godoc
//
//	@Summary		Schedule Show
//	@Description	schedule a movie to a screen and slot on a date
//	@Tags			Admin-Schedule
//	@Accept			json
//	@Produce		json
//	@security		BasicAuth
//	@param Authorization header string true "Enter basic auth"
//	@Param request body request.ScheduleShowRequest true "Schedule show request"
//	@Success		201	{object}	response.ScheduleShowResponse
//	@Failure		400	{object}	ae.AppError
//	@Failure		404	{object}	ae.AppError
//	@Failure		500	{object}	ae.AppError
//	@Router			/admin/shows [post]
func (ac *adminScheduleController) ScheduleShow(c *gin.Context) {
	var req request.ScheduleShowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("bad request %v", validator.HandleStructValidationError(err))
		c.AbortWithStatusJSON(http.StatusBadRequest, validator.HandleStructValidationError(err))
		return
	}

	resp, responseError := ac.service.ScheduleShow(c.Request.Context(), req)
	if responseError != nil {
		err := responseError.(*ae.AppError)
		logger.Error("%s", err.UnWrap().Error())
		c.AbortWithStatusJSON(err.HTTPCode(), err)
		return
	}

	c.IndentedJSON(http.StatusCreated, resp)
}

// DeleteShow godoc
//
//	@Summary		Delete Show
//	@Description	delete a scheduled show by id
//	@Tags			Admin-Schedule
//	@Accept			json
//	@Produce		json
//	@security		BasicAuth
//	@param Authorization header string true "Enter basic auth"
//	@Param showId path int true "Show ID"
//	@Success		200	{object}	map[string]string
//	@Failure		400	{object}	ae.AppError
//	@Failure		404	{object}	ae.AppError
//	@Failure		500	{object}	ae.AppError
//	@Router			/admin/shows/{showId} [delete]
func (ac *adminScheduleController) DeleteShow(c *gin.Context) {
	showID, parseErr := strconv.Atoi(c.Param("showId"))
	if parseErr != nil || showID <= 0 {
		err := ae.BadRequestError("InvalidShowID", "showId must be a positive integer", parseErr)
		c.AbortWithStatusJSON(err.HTTPCode(), err)
		return
	}

	responseError := ac.service.DeleteShow(c.Request.Context(), showID)
	if responseError != nil {
		err := responseError.(*ae.AppError)
		logger.Error("%s", err.UnWrap().Error())
		c.AbortWithStatusJSON(err.HTTPCode(), err)
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "show deleted successfully"})
}
