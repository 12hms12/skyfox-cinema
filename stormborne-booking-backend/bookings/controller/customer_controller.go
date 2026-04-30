package controller

import (
	"context"
	"net/http"
	"skyfox/bookings/dto/request"
	"skyfox/bookings/dto/response"
	"skyfox/bookings/service"
	"skyfox/common/logger"
	"skyfox/bookings/common"
	ae "skyfox/error"

	"github.com/gin-gonic/gin"
)

type CustomerInterface interface {
	Signup(context.Context, request.CustomerSignupRequest) (*response.LoginResponse, error)
	Login(context.Context, request.CustomerLoginRequest) (*response.LoginResponse, error)
	ForgotPassword(ctx context.Context, email string) (string, error)
	ResetPassword(ctx context.Context, token string, password string) error
	VerifyResetToken(ctx context.Context, token string) error
	GetProfile(ctx context.Context, userID uint) (*response.CustomerProfileResponse, error)
	UpdateProfile(ctx context.Context, userID uint, req request.CustomerUpdateProfileRequest) (*response.UpdateProfileResponse, error)
}

type onlineCustomerController struct {
	customerService CustomerInterface
	captchaService  service.CaptchaServiceInterface
}

func NewOnlineCustomerController(customerService CustomerInterface, captchaService service.CaptchaServiceInterface) *onlineCustomerController {
	return &onlineCustomerController{
		customerService: customerService,
		captchaService:  captchaService,
	}
}

// Signup godoc
//
//	@Summary		Customer Signup
//	@Description	Register new online customer
//	@Tags			OnlineCustomer
//	@Accept			json
//	@Produce		json
//	@Param			request	body	request.CustomerSignupRequest	true	"Signup Request"
//	@Success		201	{object}	model.Customer
//	@Failure		400	{object}	ae.AppError
//	@Failure		500	{object}	ae.AppError
//	@Router			/customer/signup [post]
func (oc *onlineCustomerController) SignupController(c *gin.Context) {
	var req request.CustomerSignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := ae.BadRequestError(
			"InvalidRequest",
			"invalid request body",
			err,
		)
		c.AbortWithStatusJSON(appErr.HTTPCode(), appErr)
		return
	}

	if common.Feature_Captcha && !oc.captchaService.Verify(req.CaptchaID, req.SliderPositionX) {
		appErr := ae.BadRequestError(
			"InvalidCaptcha",
			"captcha verification failed",
			nil,
		)
		c.AbortWithStatusJSON(appErr.HTTPCode(), appErr)
		return
	}

	_, responseError := oc.customerService.Signup(c.Request.Context(), req)
	if responseError != nil {
		appErr, ok := responseError.(*ae.AppError)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "internal server error",
			})
			return
		}
		if unwrapped := appErr.UnWrap(); unwrapped != nil {
			logger.Error("%s", unwrapped.Error())
		}
		c.AbortWithStatusJSON(appErr.HTTPCode(), appErr)
		return
	}
	c.IndentedJSON(http.StatusCreated, gin.H{"message": "Account created successfully"})
}

// Login godoc
//
//	@Summary		Customer Login
//	@Description	Authenticate online customer
//	@Tags			OnlineCustomer
//	@Accept			json
//	@Produce		json
//	@Param			request	body	request.CustomerLoginRequest	true	"Login Request"
//	@Success		200	{object}	response.LoginResponse
//	@Failure		400	{object}	ae.AppError
//	@Failure		401	{object}	ae.AppError
//	@Failure		500	{object}	ae.AppError
//	@Router			/login [post]
func (oc *onlineCustomerController) LoginController(c *gin.Context) {
	var req request.CustomerLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := ae.BadRequestError(
			"InvalidRequest",
			"invalid request body",
			err,
		)
		c.AbortWithStatusJSON(appErr.HTTPCode(), appErr)
		return
	}
	loginResponse, responseError := oc.customerService.Login(c.Request.Context(), req)
	if responseError != nil {
		appErr, ok := responseError.(*ae.AppError)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "internal server error",
			})
			return
		}
		if unwrapped := appErr.UnWrap(); unwrapped != nil {
			logger.Error("%s", unwrapped.Error())
		}
		c.AbortWithStatusJSON(appErr.HTTPCode(), appErr)
		return
	}
	c.IndentedJSON(http.StatusOK, loginResponse)
}

type forgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type resetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"password" binding:"required,min=6"`
}

// ForgotPassword handles the initial request to reset a password.
//
//	@Summary		Forgot Password
//	@Description	Validates email and sends a password reset link to the user
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		forgotPasswordRequest	true	"Email of the user"
//	@Success		200		{object}	map[string]interface{}	"Email Found in DB"
//	@Failure		400		{object}	map[string]interface{}	"Invalid request or Forgot password failed"
//	@Router			/forgot-password [POST]
func (o *onlineCustomerController) ForgotPassword(c *gin.Context) {
	var req forgotPasswordRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request",
			"error":   err.Error(),
		})
		return
	}
	resetLink, err1 := o.customerService.ForgotPassword(c.Request.Context(), req.Email)
	if err1 != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err1.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"resetLink": resetLink,
		"message":   "Email Found in DB",
	})
}

// ResetPassword updates the user's password using a valid token.
//
//	@Summary		Reset Password
//	@Description	Verifies the reset token and updates the password in the database
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		resetPasswordRequest	true	"Token and new password"
//	@Success		200		{object}	map[string]interface{}	"Password reset successful"
//	@Failure		400		{object}	map[string]interface{}	"Invalid request or Reset password failed"
//	@Router			/reset-password [POST]
func (o *onlineCustomerController) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request",
			"error":   err.Error(),
		})
		return
	}
	err1 := o.customerService.ResetPassword(
		c.Request.Context(),
		req.Token,
		req.NewPassword,
	)
	if err1 != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Reset password failed",
			"error":   err1.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Password reset successful",
	})
}

// VerifyResetToken checks if the provided reset token is valid.
//
//	@Summary		Verify Reset Token
//	@Description	Validates the password reset token before allowing the user to reset the password
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			token	query		string	true	"Password reset token"
//	@Success		200		{object}	map[string]interface{}	"Valid token"
//	@Failure		400		{object}	map[string]interface{}	"Token missing or invalid"
//	@Router			/verify-reset-token [GET]
func (o *onlineCustomerController) VerifyResetToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Token missing",
		})
		return
	}
	err := o.customerService.VerifyResetToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Valid token",
	})
}

// ResetPasswordPage handles the request to display or verify the password reset token.
//
//	@Summary		View Reset Password Page
//	@Description	Retrieves the reset token from the URL query parameters to initiate the password reset process
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			token	query		string					true	"Reset Token"
//	@Success		200		{object}	map[string]interface{}	"Token received successfully"
//	@Failure		400		{object}	map[string]interface{}	"Token missing or invalid"
//	@Router			/reset-password [GET]
func (o *onlineCustomerController) ResetPasswordPage(c *gin.Context) {
	token := c.Query("token")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Token received",
		"token":   token,
	})
}

// GetProfileController returns the profile of the authenticated online customer.
//
// @Summary      View Profile
// @Description  Returns the profile details of the currently authenticated user
// @Tags         OnlineCustomer
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Profile fetched successfully"
// @Failure      400  {object}  map[string]interface{}  "Failed to fetch profile"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      500  {object}  map[string]interface{}  "Invalid user ID or server error"
// @Router       /customer/profile [get]
func (o *onlineCustomerController) GetProfileController(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized"})
		return
	}

	userID, ok := userIDVal.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Invalid user ID"})
		return
	}

	profile, err := o.customerService.GetProfile(c.Request.Context(), uint(userID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Failed to fetch profile",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Profile fetched successfully",
		"data":    profile,
	})
}


// UpdateProfileController updates the profile of the authenticated online customer.
//
// @Summary      Update Profile
// @Description  Updates editable profile fields (first name, last name, email) of the authenticated user.
//
//	Pass only the fields you want to update. Leave fields as empty string ("") to skip updating them.
//
// @Tags         OnlineCustomer
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header    string                              true  "Bearer <token>"
// @Param        request        body      request.CustomerUpdateProfileRequest true  "Update Profile Request"
// @Success      200            {object}  map[string]interface{}              "Profile updated successfully"
// @Failure      400            {object}  ae.AppError                         "Invalid request or validation error"
// @Failure      401            {object}  map[string]interface{}              "Unauthorized"
// @Failure      422            {object}  ae.AppError                         "Email already in use"
// @Failure      500            {object}  map[string]interface{}              "Internal server error"
// @Router       /profile [put]
func (o *onlineCustomerController) UpdateProfileController(c *gin.Context) {
    userIDVal, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized"})
        return
    }

    userID, ok := userIDVal.(int64)
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Invalid user ID"})
        return
    }

    var req request.CustomerUpdateProfileRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        appErr := ae.BadRequestError("InvalidRequest", "invalid request body", err)
        c.AbortWithStatusJSON(appErr.HTTPCode(), appErr)
        return
    }

    resp, err := o.customerService.UpdateProfile(c.Request.Context(), uint(userID), req)
    if err != nil {
        appErr, ok := err.(*ae.AppError)
        if !ok {
            c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
            return
        }
        if unwrapped := appErr.UnWrap(); unwrapped != nil {
            logger.Error("%s", unwrapped.Error())
        }
        c.AbortWithStatusJSON(appErr.HTTPCode(), appErr)
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": resp.Message,
    })
}