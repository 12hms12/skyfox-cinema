package response

type AvatarResponse struct {
	UserID     uint   `json:"user_id"`
	AvatarURL  string `json:"avatar_url"`
	AvatarType string `json:"avatar_type"`
	UpdatedAt  string `json:"updated_at,omitempty"`
}

type PredefinedAvatarResponse struct {
	ID     int64  `json:"id"`
	Gender string `json:"gender"`
	URL    string `json:"url"`
}

type PredefinedAvatarListResponse struct {
	Items []PredefinedAvatarResponse `json:"items"`
	Total int                        `json:"total"`
}
