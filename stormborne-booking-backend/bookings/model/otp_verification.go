package model

import "time"

type OTPVerification struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	Code      string    `gorm:"size:10" json:"code"`
	Recipient string    `gorm:"size:255;not null" json:"recipient"`
	Type      string    `gorm:"size:20;not null;comment:EMAIL or SMS" json:"type"`
	Purpose   string    `gorm:"size:50;default:VERIFICATION;comment:VERIFICATION, RESET_PASSWORD" json:"purpose"`
	ExpiresAt time.Time `json:"expires_at"`
	Attempts  int       `gorm:"default:0" json:"attempts"`
	IsUsed    bool      `gorm:"default:false" json:"is_used"`
	IP        string    `gorm:"size:45" json:"ip"`
	CreatedAt time.Time `json:"created_at"`
}

func (OTPVerification) TableName() string {
	return "otp_verification"
}
