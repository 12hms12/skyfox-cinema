package request


type UpdateAvatarRequest struct {
	PredefinedAvatarID int64 `json:"predefined_avatar_id,omitempty"`

	AvatarURL string `json:"avatar_url,omitempty"`
	
	AvatarType string `json:"avatar_type,omitempty"`
}