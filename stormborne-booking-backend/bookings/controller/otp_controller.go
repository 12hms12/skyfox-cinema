package controller

import (
	"context"
	"net/http"
	"skyfox/bookings/dto/request"
	"skyfox/bookings/dto/response"
	ae "skyfox/error"

	"github.com/gin-gonic/gin"
)

type SMSOTPServiceInterface interface {
	RequestOTP(ctx context.Context, phoneNumber, ip string) (*response.OTPSendResponse, error)
	VerifyOTP(ctx context.Context, phoneNumber, code string) (*response.OTPVerifyResponse, error)
	ResendOTP(ctx context.Context, phoneNumber, ip string) (*response.OTPSendResponse, error)
	ListAllOTPs(ctx context.Context) (*response.OTPListResponse, error)
}

type otpController struct {
	smsService SMSOTPServiceInterface
}

func NewOTPController(smsService SMSOTPServiceInterface) *otpController {
	return &otpController{
		smsService: smsService,
	}
}

// RequestSMSOTP godoc
//
//	@Summary		Request SMS OTP
//	@Description	Generate and send OTP to a phone number (illusion mode)
//	@Tags			OTP
//	@Accept			json
//	@Produce		json
//	@Param			request	body		request.OTPSendRequest	true	"Phone number"
//	@Success		200		{object}	response.OTPSendResponse
//	@Failure		400		{object}	ae.AppError
//	@Router			/otp/sms/request [post]
func (oc *otpController) RequestSMSOTP(c *gin.Context) {
	var req request.OTPSendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := ae.BadRequestError("InvalidRequest", "invalid request body", err)
		c.AbortWithStatusJSON(appErr.HTTPCode(), appErr)
		return
	}

	ip := c.ClientIP()
	resp, err := oc.smsService.RequestOTP(c.Request.Context(), req.Recipient, ip)
	if err != nil {
		handleOTPError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, resp)
}

// VerifySMSOTP godoc
//
//	@Summary		Verify SMS OTP
//	@Description	Verify the OTP code sent to a phone number
//	@Tags			OTP
//	@Accept			json
//	@Produce		json
//	@Param			request	body		request.OTPVerifyRequest	true	"Phone number and OTP code"
//	@Success		200		{object}	response.OTPVerifyResponse
//	@Failure		400		{object}	ae.AppError
//	@Router			/otp/sms/verify [post]
func (oc *otpController) VerifySMSOTP(c *gin.Context) {
	var req request.OTPVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := ae.BadRequestError("InvalidRequest", "invalid request body", err)
		c.AbortWithStatusJSON(appErr.HTTPCode(), appErr)
		return
	}

	resp, err := oc.smsService.VerifyOTP(c.Request.Context(), req.Recipient, req.Code)
	if err != nil {
		handleOTPError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, resp)
}

// ResendSMSOTP godoc
//
//	@Summary		Resend SMS OTP
//	@Description	Resend OTP to a phone number with rate limiting
//	@Tags			OTP
//	@Accept			json
//	@Produce		json
//	@Param			request	body		request.OTPResendRequest	true	"Phone number"
//	@Success		200		{object}	response.OTPSendResponse
//	@Failure		400		{object}	ae.AppError
//	@Router			/otp/sms/resend [post]
func (oc *otpController) ResendSMSOTP(c *gin.Context) {
	var req request.OTPResendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := ae.BadRequestError("InvalidRequest", "invalid request body", err)
		c.AbortWithStatusJSON(appErr.HTTPCode(), appErr)
		return
	}

	ip := c.ClientIP()
	resp, err := oc.smsService.ResendOTP(c.Request.Context(), req.Recipient, ip)
	if err != nil {
		handleOTPError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, resp)
}

// ListAllOTPs godoc
//
//	@Summary		List All OTPs
//	@Description	Returns all OTPs for the illusion dashboard (no auth required)
//	@Tags			OTP
//	@Produce		json
//	@Success		200	{object}	response.OTPListResponse
//	@Failure		500	{object}	ae.AppError
//	@Router			/otp/all [get]
func (oc *otpController) ListAllOTPs(c *gin.Context) {
	resp, err := oc.smsService.ListAllOTPs(c.Request.Context())
	if err != nil {
		handleOTPError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, resp)
}

func handleOTPError(c *gin.Context, err error) {
	appErr, ok := err.(*ae.AppError)
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "internal server error",
		})
		return
	}
	c.AbortWithStatusJSON(appErr.HTTPCode(), appErr)
}
