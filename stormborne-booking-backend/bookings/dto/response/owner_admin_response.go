package response

type OwnerAdminItemResponse struct {
	ID          uint   `json:"id"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
	CountryCode string `json:"countryCode"`
	Username    string `json:"username"`
	Age         uint   `json:"age"`
	Gender      string `json:"gender"`
	Role        string `json:"role"`
}

type OwnerAdminsListResponse struct {
	Admins []OwnerAdminItemResponse `json:"admins"`
}
