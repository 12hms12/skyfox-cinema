package request

type OTPSendRequest struct {
	Recipient string `json:"recipient" binding:"required"`
}

type OTPVerifyRequest struct {
	Recipient string `json:"recipient" binding:"required"`
	Code      string `json:"code" binding:"required,len=4"`
}

type OTPResendRequest struct {
	Recipient string `json:"recipient" binding:"required"`
}
