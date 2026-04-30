package response

type OTPSendResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ExpiresIn int    `json:"expires_in_seconds"`
}

type OTPVerifyResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Verified bool   `json:"verified"`
}

type OTPItem struct {
	ID        int    `json:"id"`
	Code      string `json:"code"`
	Recipient string `json:"recipient"`
	Type      string `json:"type"`
	Purpose   string `json:"purpose"`
	ExpiresAt string `json:"expires_at"`
	IsUsed    bool   `json:"is_used"`
	Attempts  int    `json:"attempts"`
	CreatedAt string `json:"created_at"`
}

type OTPListResponse struct {
	OTPs []OTPItem `json:"otps"`
}
