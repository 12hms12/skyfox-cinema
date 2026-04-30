package request

type PaymentGatewayRequest struct {
	CardNumber string `json:"cardNumber" binding:"required"`

	ExpiryDate string `json:"expiryDate" binding:"required"`

	Name string `json:"name" binding:"required"`

	CVV int `json:"cvv" binding:"required"`
}
