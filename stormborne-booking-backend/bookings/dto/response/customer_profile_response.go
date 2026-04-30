package response

type CustomerProfileResponse struct {
	ID          uint   `json:"id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	CountryCode string `json:"country_code"`
	Age         uint   `json:"age"`
	Gender      string `json:"gender"`
	Username    string `json:"username"`
	AvatarURL   string `json:"avatar_url"`
	AvatarType  string `json:"avatar_type"`
}
