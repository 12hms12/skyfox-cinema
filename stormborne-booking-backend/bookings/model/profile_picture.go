package model

import "time"

type ProfilePicture struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	UserID      uint      `gorm:"uniqueIndex;not null"`
	ImageData   string    `gorm:"type:text;not null"`
	ContentType string    `gorm:"type:varchar(50);not null;default:'image/jpeg'"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

func (ProfilePicture) TableName() string {
	return "profile_pictures"
}
