package controller

import (
	"net/http"
	"strconv"

	"skyfox/bookings/dto/request"
	"skyfox/bookings/service"
	ae "skyfox/error"

	"github.com/gin-gonic/gin"
)

type ownerAdminController struct {
	service *service.OwnerAdminService
}

func NewOwnerAdminController(service *service.OwnerAdminService) *ownerAdminController {
	return &ownerAdminController{service: service}
}

func (oc *ownerAdminController) ListAdmins(c *gin.Context) {
	result, err := oc.service.ListAdmins(c.Request.Context())
	if err != nil {
		if appErr, ok := err.(*ae.AppError); ok {
			c.AbortWithStatusJSON(appErr.HTTPCode(), appErr)
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (oc *ownerAdminController) AddAdmin(c *gin.Context) {
	var req request.OwnerAddAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := ae.BadRequestError("InvalidRequest", "invalid request body", err)
		c.AbortWithStatusJSON(appErr.HTTPCode(), appErr)
		return
	}

	if err := oc.service.AddAdmin(c.Request.Context(), req); err != nil {
		if appErr, ok := err.(*ae.AppError); ok {
			c.AbortWithStatusJSON(appErr.HTTPCode(), appErr)
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Admin added successfully"})
}

func (oc *ownerAdminController) RemoveAdmin(c *gin.Context) {
	adminIDValue := c.Param("adminId")
	adminID, err := strconv.Atoi(adminIDValue)
	if err != nil || adminID <= 0 {
		appErr := ae.BadRequestError("InvalidAdminID", "invalid admin id", err)
		c.AbortWithStatusJSON(appErr.HTTPCode(), appErr)
		return
	}

	if err := oc.service.RemoveAdmin(c.Request.Context(), uint(adminID)); err != nil {
		if appErr, ok := err.(*ae.AppError); ok {
			c.AbortWithStatusJSON(appErr.HTTPCode(), appErr)
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Admin removed successfully"})
}
