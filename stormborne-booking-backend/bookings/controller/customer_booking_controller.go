package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	ae "skyfox/error"

	"skyfox/bookings/common"
	"skyfox/bookings/dto/request"
	"skyfox/bookings/service"
	"skyfox/common/logger"
)

type CustomerBookingController struct {
	customerBookingService service.CustomerBookingService
}

func NewCustomerBookingController(svc service.CustomerBookingService) *CustomerBookingController {
	return &CustomerBookingController{customerBookingService: svc}
}

func (c *CustomerBookingController) CreateBooking(ctx *gin.Context) {
	if !common.Feature_CustomerBooking {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "feature not available"})
		return
	}

	var req request.CreateCustomerBookingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	resp, err := c.customerBookingService.CreateBooking(ctx.Request.Context(), userID.(int64), req.ShowID, req.SeatIDs)
	if err != nil {
		respondWithAppError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

func (c *CustomerBookingController) ProcessPayment(ctx *gin.Context) {
	if !common.Feature_CustomerBooking {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "feature not available"})
		return
	}

	bookingID, err := strconv.Atoi(ctx.Param("bookingId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id"})
		return
	}

	var req request.PaymentRequest
	if bindErr := ctx.ShouldBindJSON(&req); bindErr != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": bindErr.Error()})
		return
	}

	resp, serviceErr := c.customerBookingService.ProcessPayment(ctx.Request.Context(), bookingID, req.Success)
	if serviceErr != nil {
		respondWithAppError(ctx, serviceErr)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (c *CustomerBookingController) GetBookingDetails(ctx *gin.Context) {
	if !common.Feature_CustomerBooking {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "feature not available"})
		return
	}

	bookingID, err := strconv.Atoi(ctx.Param("bookingId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id"})
		return
	}

	resp, serviceErr := c.customerBookingService.GetBookingDetails(ctx.Request.Context(), bookingID)
	if serviceErr != nil {
		respondWithAppError(ctx, serviceErr)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func respondWithAppError(ctx *gin.Context, err error) {
	if appErr, ok := err.(*ae.AppError); ok {
		ctx.JSON(appErr.HTTPCode(), gin.H{"error": appErr.Error()})
		return
	}
	ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func (c *CustomerBookingController) CheckedIn(ctx *gin.Context) {
	idParam := ctx.Param("bookingId")

	bookingID, err := strconv.Atoi(idParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid booking id",
		})
		return
	}

	responseError := c.customerBookingService.CheckIn(ctx.Request.Context(), bookingID)
	if responseError != nil {
		err := responseError.(*ae.AppError)
		logger.Error("error occurred. %v", err)
		ctx.AbortWithStatusJSON(err.HTTPCode(), err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "checked in successfully",
	})
}

