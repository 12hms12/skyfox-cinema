package model

import (
	"time"
)

type Gender string
type AvatarType string

const (
	MALE              Gender = "male"
	FEMALE            Gender = "female"
	PREFER_NOT_TO_SAY Gender = "prefer_not_to"
)

const (
	PREDEFINED AvatarType = "predefined"
	UPLOADED   AvatarType = "uploaded"
)

type Customer struct {
	ID                      uint       `gorm:"primaryKey;autoIncrement"`
	FirstName               string     `gorm:"type:varchar(100);not null"`
	LastName                string     `gorm:"type:varchar(100);not null"`
	Email                   string     `gorm:"type:varchar(150);unique;not null"`
	PhoneNumber             string     `gorm:"type:varchar(20);unique;not null"`
	CountryCode             string     `gorm:"not null"`
	Age                     uint       `gorm:"not null;check:age >= 0;check:age <= 130"`
	Gender                  Gender     `gorm:"type:varchar(100);not null"`
	Username                string     `gorm:"unique;not null;"`
	Password                string     `gorm:"type:varchar(100);not null"`
	AvatarURL               string     `gorm:"type:text"`
	AvatarType              AvatarType `gorm:"type:varchar(100)"`
	IsEmailVerified         bool       `gorm:"type:boolean;default:false"`
	IsPhoneVerified         bool       `gorm:"type:boolean;default:false"`
	PasswordResetToken      *string    `gorm:"type:text"`
	PasswordTokenExpiryTime time.Time  `gorm:"column:password_token_expiry_time"`
	UserRole                string     `gorm:"type:varchar(100)"`
	CreatedAt               time.Time  `gorm:"autoCreateTime"`
	UpdatedAt               time.Time  `gorm:"autoUpdateTime"`
}

func (Customer) TableName() string {
	return "customer"
}
