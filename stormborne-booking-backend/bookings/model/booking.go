package model

import "time"

type Booking struct {
	ID            int       `gorm:"primaryKey"`
	TransactionID string

	OnlineCustomerID uint
	Customer   Customer `gorm:"foreignKey:OnlineCustomerID"`

	ShowID int
	Show   Show `gorm:"foreignKey:ShowID"`

	TotalPrice float64
	Status     string    // PENDING, CONFIRMED, CANCELLED

	IsCheckedIn bool `gorm:"default:false"`

	ExpiresAt  time.Time
	CreatedAt  time.Time

	BookedSeats []BookedSeat
}

func (Booking) TableName() string {
	return "booking"
}