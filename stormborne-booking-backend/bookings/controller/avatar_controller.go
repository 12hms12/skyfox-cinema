package controller

import (
	"context"
	"net/http"
	"skyfox/bookings/constants"
	"skyfox/bookings/dto/request"
	"skyfox/bookings/dto/response"
	ae "skyfox/error"

	"github.com/gin-gonic/gin"
)

type AvatarService interface {
	UpdateAvatar(ctx context.Context, userID uint, req request.UpdateAvatarRequest) (*response.AvatarResponse, error)
	GetUserAvatar(ctx context.Context, userID uint) (*response.AvatarResponse, error)
	AssignRandomAvatar(ctx context.Context, userID uint, gender string) (*response.AvatarResponse, error)
	ListPredefinedAvatars(ctx context.Context, gender string) (*response.PredefinedAvatarListResponse, error)
}

type AvatarController struct {
	avatarService AvatarService
}

func NewAvatarController(svc AvatarService) *AvatarController {
	return &AvatarController{avatarService: svc}
}

func extractUserID(c *gin.Context) (uint, bool) {
	uidVal, ok := c.Get("userID")
	if !ok {
		return 0, false
	}
	switch v := uidVal.(type) {
	case uint:
		return v, true
	case int64:
		if v < 0 {
			return 0, false
		}
		return uint(v), true
	case int:
		if v < 0 {
			return 0, false
		}
		return uint(v), true
	default:
		return 0, false
	}
}

// UpdateMyAvatar godoc
//
//	@Summary		Update my avatar
//	@Description	Update the avatar for the authenticated user
//	@Tags			Avatar
//	@Accept			json
//	@Produce		json
//	@Param			request	body	request.UpdateAvatarRequest	true	"Update Avatar Request"
//	@Success		200	{object}	response.AvatarResponse
//	@Failure		400	{object}	ae.AppError
//	@Failure		401	{object}	ae.AppError
//	@Failure		500	{object}	ae.AppError
//	@Router			/avatar [put]
//	@Security		BearerAuth
func (h *AvatarController) UpdateMyAvatar(c *gin.Context) {
	userID, ok := extractUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{constants.ErrorDesc: constants.Unauthorized})
		return
	}

	var req request.UpdateAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{constants.ErrorDesc: constants.InvalidPayload})
		return
	}

	resp, err := h.avatarService.UpdateAvatar(c.Request.Context(), userID, req)
	if err != nil {
		if appErr, ok := err.(*ae.AppError); ok {
			c.AbortWithStatusJSON(appErr.HTTPCode(), appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{constants.ErrorDesc: constants.UnableToFetchAvatar})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetMyAvatar godoc
//
//	@Summary		Get my avatar
//	@Description	Retrieve the avatar of the authenticated user
//	@Tags			Avatar
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.AvatarResponse
//	@Failure		401	{object}	ae.AppError
//	@Failure		500	{object}	ae.AppError
//	@Router			/avatar [get]
//	@Security		BearerAuth
func (h *AvatarController) GetMyAvatar(c *gin.Context) {
	userID, ok := extractUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{constants.ErrorDesc: constants.Unauthorized})
		return
	}

	resp, err := h.avatarService.GetUserAvatar(c.Request.Context(), userID)
	if err != nil {
		if appErr, ok := err.(*ae.AppError); ok {
			c.AbortWithStatusJSON(appErr.HTTPCode(), appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{constants.ErrorDesc: constants.UnableToFetchAvatar})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListPredefinedAvatars godoc
//
//	@Summary		List predefined avatars
//	@Description	Get predefined avatars optionally filtered by gender
//	@Tags			Avatar
//	@Accept			json
//	@Produce		json
//	@Param			gender	query	string	false	"Gender filter"
//	@Success		200	{object}	response.PredefinedAvatarListResponse
//	@Failure		500	{object}	ae.AppError
//	@Router			/avatars/predefined [get]
func (h *AvatarController) ListPredefinedAvatars(c *gin.Context) {
	gender := c.Query("gender")

	resp, err := h.avatarService.ListPredefinedAvatars(c.Request.Context(), gender)
	if err != nil {
		if appErr, ok := err.(*ae.AppError); ok {
			c.AbortWithStatusJSON(appErr.HTTPCode(), appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{constants.ErrorDesc: constants.UnableToFetchAvatar})
		return
	}
	c.JSON(http.StatusOK, resp)
}
