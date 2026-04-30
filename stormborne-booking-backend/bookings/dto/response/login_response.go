package response

import "skyfox/bookings/model"

type LoginResponse struct {
	Token      string           `json:"token"`
	Username   string           `json:"username"`
	AvatarURL  string           `json:"avatar_url"`
	AvatarType model.AvatarType `json:"avatar_type"`
	Gender     model.Gender     `json:"gender"`
	Age        uint             `json:"age"`
	Role       string           `json:"role"`
	Id         string           `json:"id"`
}
