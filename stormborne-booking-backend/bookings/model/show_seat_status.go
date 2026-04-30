package model

import "time"

type ShowSeatStatus struct {
	ID      int `gorm:"primaryKey"`

	ShowID  int
	Show    Show `gorm:"foreignKey:ShowID"`

	SeatID  int
	Seat    Seat `gorm:"foreignKey:SeatID"`

	Status      string     // AVAILABLE, BLOCKED, BOOKED
	LockedUntil *time.Time

	BookedSeats []BookedSeat
}

func (ShowSeatStatus) TableName() string {
	return "show_seat_status"
}