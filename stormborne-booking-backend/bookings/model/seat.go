package model

type Seat struct {
	ID           int    `gorm:"primaryKey"`
	ScreenID     int
	Screen Screen `gorm:"foreignKey:ScreenID"`

	RowNumber    int
	ColumnNumber int
	SeatType     string
}

func (Seat) TableName() string {
	return "seat"
}