package model

type Screen struct {
	ID         int    `gorm:"primaryKey"`
	ScreenName string
	
	Seats      []Seat `gorm:"foreignKey:ScreenID"`
	Shows      []Show `gorm:"foreignKey:ScreenID"`
}

func (Screen) TableName() string {
	return "screen"
}