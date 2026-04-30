package request

type CaptchaVerifyRequest struct {
	CaptchaID       string  `json:"captcha_id" binding:"required"`
	SliderPositionX float64 `json:"slider_position_x" binding:"required"`
}
