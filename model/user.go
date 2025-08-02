package model

import (
	"time"
	"gorm.io/gorm"
)

type Users struct {
	ID                    uint      `gorm:"column:id;primaryKey"`
	CreatedAt             time.Time `gorm:"column:created_at"`
	UpdatedAt             time.Time `gorm:"column:updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"column:deleted_at"`
	Username              string    `gorm:"column:username"`
	Name                  string    `gorm:"column:name"`
	Email                 string    `gorm:"column:email"`
	IsEmailVerified       bool      `gorm:"column:is_email_verified"`
	VerificationCode      string    `gorm:"column:verification_code"`
	VerificationExpiresAt time.Time `gorm:"column:verification_expires_at"`
	Phone                 string    `gorm:"column:phone"`
	PasswordHash          string    `gorm:"column:password_hash"`
	Role                  string    `gorm:"column:role"`
	ProfileImageUrl       string    `gorm:"column:profile_image_url"`
	CoverImageUrl         string    `gorm:"column:cover_image_url"`
	Bio                   string    `gorm:"column:bio"`
	IsVerified            bool      `gorm:"column:is_verified"`
	Gender                string    `gorm:"column:gender"`
	Birthday              time.Time `gorm:"column:birthday"`
	IsActive              *bool     `gorm:"column:is_active;default:true"`
	Tfa_verifed           bool      `gorm:"column:tfa_verifed;default:false"`
	Login_codes           string    `gorm:"column:login_codes"`
	Login_codes_set       bool      `gorm:"column:login_codes_set;default:false"`
	Tfa_code              string     `gorm:"column:tfa_code"` 
	Token_version         int     `gorm:"column:token_version"`
	Provider              string   `gorm:"column:provider"`
}
func (Users) TableName() string {
	return "users" // 👈 exact match with your actual Postgres table name
}
func (DeviceRecord) TableName() string {
	return "device_record"
}

func (OldPassword)TableName()string{
	return "old_password"
}
type DeviceRecord struct {
	ID        uint      `gorm:"column:id;primaryKey"`
	UserID    uint      `gorm:"column:userid;not null"`
	City      string    `gorm:"column:city"`
	Region    string    `gorm:"column:region"`
	Country   string    `gorm:"column:country"`
	Locale    string    `gorm:"column:locale"`
	Lat       float64   `gorm:"column:lat"`
	Lon       float64   `gorm:"column:lon"`
	ZipCode   string    `gorm:"column:zipcode"`
	LastLogin time.Time `gorm:"column:last_login"`
	Browser   string    `gorm:"column:browser"`
}
type OldPassword struct{
	ID        uint      `gorm:"column:id;primaryKey"`
UserID uint `gorm:"column:user_id;not null"`
Password string `gorm:"column:password"`
}
type Session struct{
			
			Jti string
			UserID int
			IsActive bool
			IssuedAT time.Time
			DeviceInfoId int
			ExpireAt time.Time
}