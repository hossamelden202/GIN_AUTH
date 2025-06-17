package model

import (
	"time"
	"gorm.io/gorm"
)

type Users struct {
	gorm.Model 

	Username              string
	Name                  string
	Email                 string
	IsEmailVerified       bool
	VerificationCode      string
	Phone                 string
	VerificationExpiresAt time.Time
	PasswordHash          string
	Role                  string
	ProfileImageUrl       string
	CoverImageUrl         string
	Bio                   string
	IsVerified            bool
	Gender                string
	Birthday              time.Time
	IsActive              *bool `gorm:"default:true"`
	LastLoginAt           time.Time
}
