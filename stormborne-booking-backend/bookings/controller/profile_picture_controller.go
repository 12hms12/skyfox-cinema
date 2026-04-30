package controller

import (
	"context"
	"io"
	"net/http"

	"skyfox/bookings/common"
	"skyfox/bookings/constants"
	"skyfox/bookings/dto/response"
	ae "skyfox/error"

	"github.com/gin-gonic/gin"
)

type ProfileImageService interface {
	UploadProfileImage(ctx context.Context, userID uint, imageData []byte, contentType string) (*response.AvatarResponse, error)
	GetProfileImage(ctx context.Context, userID uint) (*response.AvatarResponse, error)
}

type ProfileImageController struct {
	service ProfileImageService
}

func NewProfileImageController(svc ProfileImageService) *ProfileImageController {
	return &ProfileImageController{service: svc}
}

func (h *ProfileImageController) UploadProfileImage(c *gin.Context) {
	if !common.Feature_Upload_Profile_Image {
		c.JSON(http.StatusNotFound, gin.H{constants.ErrorDesc: constants.FeatureNotAvailable})
		return
	}

	userID, ok := extractUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{constants.ErrorDesc: constants.Unauthorized})
		return
	}

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{constants.ErrorDesc: constants.InvalidPayload})
		return
	}
	defer file.Close()

	imageData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{constants.ErrorDesc: constants.UnableToUploadImage})
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	resp, svcErr := h.service.UploadProfileImage(c.Request.Context(), userID, imageData, contentType)
	if svcErr != nil {
		if appErr, ok := svcErr.(*ae.AppError); ok {
			c.AbortWithStatusJSON(appErr.HTTPCode(), appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{constants.ErrorDesc: constants.UnableToUploadImage})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ProfileImageController) GetProfileImage(c *gin.Context) {
	if !common.Feature_Upload_Profile_Image {
		c.JSON(http.StatusNotFound, gin.H{constants.ErrorDesc: constants.FeatureNotAvailable})
		return
	}

	userID, ok := extractUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{constants.ErrorDesc: constants.Unauthorized})
		return
	}

	resp, err := h.service.GetProfileImage(c.Request.Context(), userID)
	if err != nil {
		if appErr, ok := err.(*ae.AppError); ok {
			c.AbortWithStatusJSON(appErr.HTTPCode(), appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{constants.ErrorDesc: constants.UnableToFetchProfileImage})
		return
	}

	c.JSON(http.StatusOK, resp)
}
