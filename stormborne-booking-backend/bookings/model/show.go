package model

import "time"

type Show struct {
	Id           int       `json:"id" gorm:"primaryKey"`
	ScreenID     int       `gorm:"uniqueIndex:idx_show_screen_date_slot"`
	Screen       Screen    `gorm:"foreignKey:ScreenID"`
	MovieId      string    `json:"movieId"`
	Movie        Movie     `gorm:"-"`
	SlotId       int       `gorm:"uniqueIndex:idx_show_screen_date_slot"`
	Slot         Slot      `json:"slot" gorm:"foreignKey:SlotId"`
	Date         time.Time `json:"date" gorm:"type:timestamp;uniqueIndex:idx_show_screen_date_slot"`
	Cost         float64   `json:"cost"`
	Pricings     []ShowPricing
	SeatStatuses []ShowSeatStatus
	Bookings     []Booking
}

func (Show) TableName() string {
	return "show"
}
