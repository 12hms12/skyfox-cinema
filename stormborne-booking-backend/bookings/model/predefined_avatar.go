package model

// PredefinedAvatar maps to the `predefined_avatars` table that holds selectable avatar URLs.
type PredefinedAvatar struct {
	ID     int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Gender string `gorm:"type:varchar(16);not null" json:"gender"` // "male" | "female" | "neutral"
	URL    string `gorm:"type:text;not null;uniqueIndex" json:"url"`
}

func (PredefinedAvatar) TableName() string {
	return "predefined_avatars"
}
