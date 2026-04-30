package response

type CaptchaGenerateResponse struct{
	CaptchaID string `json:"captcha_id"`
	MasterImage string `json:"master_image"`
	TileImage string `json:"tile_image"`
	TileX int `json:"tile_x"`
	TileY int `json:"tile_y"`
	TileWidth int `json:"tile_width"`
	TileHeight int `json:"tile_height"`
}