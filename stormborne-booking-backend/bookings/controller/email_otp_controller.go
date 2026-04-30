package controller

import (
	"context"
	"net/http"
	"skyfox/bookings/dto/request"
	"skyfox/bookings/dto/response"
	ae "skyfox/error"

	"github.com/gin-gonic/gin"
)

type EmailOTPServiceInterface interface {
	RequestOTP(ctx context.Context, email, ip string) (*response.OTPSendResponse, error)
	VerifyOTP(ctx context.Context, email, code string) (*response.OTPVerifyResponse, error)
	ResendOTP(ctx context.Context, email, ip string) (*response.OTPSendResponse, error)
}

type emailOtpController struct {
	emailService EmailOTPServiceInterface
}

func NewEmailOTPController(emailService EmailOTPServiceInterface) *emailOtpController {
	return &emailOtpController{
		emailService: emailService,
	}
}

// RequestEmailOTP godoc
//
//	@Summary		Request Email OTP
//	@Description	Generate and send OTP to an email address (illusion mode)
//	@Tags			OTP
//	@Accept			json
//	@Produce		json
//	@Param			request	body		request.OTPSendRequest	true	"Email address"
//	@Success		200		{object}	response.OTPSendResponse
//	@Failure		400		{object}	ae.AppError
//	@Router			/otp/email/request [post]
func (ec *emailOtpController) RequestEmailOTP(c *gin.Context) {
	var req request.OTPSendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := ae.BadRequestError("InvalidRequest", "invalid request body", err)
		c.AbortWithStatusJSON(appErr.HTTPCode(), appErr)
		return
	}

	ip := c.ClientIP()
	resp, err := ec.emailService.RequestOTP(c.Request.Context(), req.Recipient, ip)
	if err != nil {
		handleOTPError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, resp)
}

// VerifyEmailOTP godoc
//
//	@Summary		Verify Email OTP
//	@Description	Verify the OTP code sent to an email address
//	@Tags			OTP
//	@Accept			json
//	@Produce		json
//	@Param			request	body		request.OTPVerifyRequest	true	"Email address and OTP code"
//	@Success		200		{object}	response.OTPVerifyResponse
//	@Failure		400		{object}	ae.AppError
//	@Router			/otp/email/verify [post]
func (ec *emailOtpController) VerifyEmailOTP(c *gin.Context) {
	var req request.OTPVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := ae.BadRequestError("InvalidRequest", "invalid request body", err)
		c.AbortWithStatusJSON(appErr.HTTPCode(), appErr)
		return
	}

	resp, err := ec.emailService.VerifyOTP(c.Request.Context(), req.Recipient, req.Code)
	if err != nil {
		handleOTPError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, resp)
}

// ResendEmailOTP godoc
//
//	@Summary		Resend Email OTP
//	@Description	Resend OTP to an email address with rate limiting
//	@Tags			OTP
//	@Accept			json
//	@Produce		json
//	@Param			request	body		request.OTPResendRequest	true	"Email address"
//	@Success		200		{object}	response.OTPSendResponse
//	@Failure		400		{object}	ae.AppError
//	@Router			/otp/email/resend [post]
func (ec *emailOtpController) ResendEmailOTP(c *gin.Context) {
	var req request.OTPResendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := ae.BadRequestError("InvalidRequest", "invalid request body", err)
		c.AbortWithStatusJSON(appErr.HTTPCode(), appErr)
		return
	}

	ip := c.ClientIP()
	resp, err := ec.emailService.ResendOTP(c.Request.Context(), req.Recipient, ip)
	if err != nil {
		handleOTPError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, resp)
}
