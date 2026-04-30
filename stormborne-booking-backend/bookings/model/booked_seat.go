package model

type BookedSeat struct {
	ID               int `gorm:"primaryKey"`

	BookingID        int
	Booking          Booking `gorm:"foreignKey:BookingID"`

	ShowSeatStatusID int
	ShowSeatStatus   ShowSeatStatus `gorm:"foreignKey:ShowSeatStatusID"`
}

func (BookedSeat) TableName() string {
	return "booked_seat"
}