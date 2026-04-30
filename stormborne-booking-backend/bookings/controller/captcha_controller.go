package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"skyfox/bookings/dto/request"
	"skyfox/bookings/dto/response"
	"skyfox/bookings/service"
)

type CaptchaController struct {
	captchaService service.CaptchaServiceInterface
}

func NewCaptchaController(captchaService service.CaptchaServiceInterface) *CaptchaController {
	return &CaptchaController{captchaService: captchaService}
}

// GenerateCaptcha godoc
// @Summary      Generate slide captcha
// @Description  Returns a new slide captcha: background image, tile image and initial positions
// @Tags         captcha
// @Produce      json
// @Success      200  {object}  response.CaptchaGenerateResponse
// @Failure      500  {object}  map[string]string
// @Router       /api/captcha/generate [get]
func (c *CaptchaController) GenerateCaptcha(ctx *gin.Context) {
	captchaID, masterImage, tileImage, tileX, tileY, tileWidth, tileHeight, err := c.captchaService.Generate()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate captcha"})
		return
	}

	ctx.JSON(http.StatusOK, response.CaptchaGenerateResponse{
		CaptchaID:   captchaID,
		MasterImage: masterImage,
		TileImage:   tileImage,
		TileX:       tileX,
		TileY:       tileY,
		TileWidth:   tileWidth,
		TileHeight:  tileHeight,
	})
}

// VerifyCaptcha godoc
// @Summary      Verify slide captcha
// @Description  Verifies the user's slider position against the stored target X
// @Tags         captcha
// @Accept       json
// @Produce      json
// @Param        body  body      request.CaptchaVerifyRequest  true  "Captcha verify payload"
// @Success      200   {object}  map[string]bool
// @Failure      400   {object}  map[string]string
// @Router       /api/captcha/verify [post]
func (c *CaptchaController) VerifyCaptcha(ctx *gin.Context) {
	var req request.CaptchaVerifyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	success := c.captchaService.Verify(req.CaptchaID, req.SliderPositionX)
	if !success {
		ctx.JSON(http.StatusOK, gin.H{"success": false})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
