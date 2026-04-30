package model

type ShowPricing struct {
	ID       int     `gorm:"primaryKey"`
	ShowID   int
	Show     Show    `gorm:"foreignKey:ShowID"`

	SeatType string
	Price    float64
}

func (ShowPricing) TableName() string {
	return "show_pricing"
}