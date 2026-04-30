package controller

import (
	"context"
	"net/http"
	"skyfox/bookings/common"
	"skyfox/bookings/dto/request"
	"skyfox/bookings/dto/response"
	"skyfox/common/logger"
	ae "skyfox/error"
	"strconv"

	"github.com/gin-gonic/gin"
)

type BookingViewService interface {
	GetAllBookings(ctx context.Context, filter request.AdminBookingViewRequest) (*response.AdminBookingViewResponse, error)
	GetAllBookingsCustomer(
		ctx context.Context,
		filter request.CustomerBookingViewRequest,
	) (*response.AdminBookingViewResponse, error)
}

type BookingViewController struct {
	bookingViewService BookingViewService
}

func NewBookingViewController(bookingViewService BookingViewService) *BookingViewController {
	return &BookingViewController{
		bookingViewService: bookingViewService,
	}
}

// GetBookings godoc
//
//	@Summary		Get All Bookings
//	@Description	Get all bookings with optional filters for admin
//	@Tags			Admin
//	@security		BearerAuth
//	@param			Authorization	header		string	true	"Enter bearer token"
//	@Param			startDate		query		string	false	"Filter by start date (YYYY-MM-DD)"
//	@Param			endDate			query		string	false	"Filter by end date (YYYY-MM-DD)"
//	@Param			movieName		query		string	false	"Filter by movie name"
//	@Produce		json
//	@Success		200	{object}	response.AdminBookingViewResponse
//	@Failure		400	{object}	ae.AppError
//	@Failure		500	{object}	ae.AppError
//	@Failure		503	{object}	ae.AppError
//	@Router			/admin/view-bookings [GET]
func (bvc *BookingViewController) GetBookings(c *gin.Context) {
	if !common.Feature_AdminViewBookings {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, ae.NewAppError(http.StatusServiceUnavailable, "FeatureDisabled", "admin view booking feature is currently disabled", nil))
		return
	}

	var filter request.AdminBookingViewRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		logger.Error("bad request: %v", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, ae.BadRequestError("BadRequest", err.Error(), err))
		return
	}

	result, err := bvc.bookingViewService.GetAllBookings(c.Request.Context(), filter)
	if err != nil {
		appErr := err.(*ae.AppError)
		logger.Error("error fetching bookings: %v", appErr)
		c.AbortWithStatusJSON(appErr.HTTPCode(), appErr)
		return
	}

	c.IndentedJSON(http.StatusOK, result)
}

func (bvc *BookingViewController) GetAllBookingsCustomer(c *gin.Context) {
	if !common.Feature_AdminViewBookings {
		c.AbortWithStatusJSON(
			http.StatusServiceUnavailable,
			ae.NewAppError(
				http.StatusServiceUnavailable,
				"FeatureDisabled",
				"admin view booking feature is currently disabled",
				nil,
			),
		)
		return
	}

	var filter request.CustomerBookingViewRequest

	// Bind query params (movieName, startDate, endDate etc.)
	if err := c.ShouldBindQuery(&filter); err != nil {
		logger.Error("bad request: %v", err)
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			ae.BadRequestError("BadRequest", err.Error(), err),
		)
		return
	}

	// Get customer id from URL param
	onlineCustomerIDStr := c.Param("onlineCustomerId")
	if onlineCustomerIDStr != "" {
		id, err := strconv.Atoi(onlineCustomerIDStr)
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				ae.BadRequestError("InvalidCustomerId", "invalid customer id", err),
			)
			return
		}
		filter.OnlineCustomerID = uint(id)
	}

	result, err := bvc.bookingViewService.GetAllBookingsCustomer(
		c.Request.Context(),
		filter,
	)

	if err != nil {
		appErr := err.(*ae.AppError)
		logger.Error("error fetching bookings: %v", appErr)
		c.AbortWithStatusJSON(appErr.HTTPCode(), appErr)
		return
	}

	c.IndentedJSON(http.StatusOK, result)
}