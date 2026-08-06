package model

import "time"

type RefreshToken struct {
	ID             uint   `json:"id" gorm:"primaryKey"`
	UserID         uint   `json:"user_id" gorm:"index;not null"`
	TokenHash      string `json:"-" gorm:"size:255;not null"`
	UserAgent      string `json:"user_agent" gorm:"size:255"`
	Browser        string `json:"browser" gorm:"size:50"`
	BrowserVersion string `json:"browser_version" gorm:"size:30"`
	OS             string `json:"os" gorm:"size:50"`
	OSVersion      string `json:"os_version" gorm:"size:30"`
	DeviceType     string `json:"device_type" gorm:"size:20"`

	IPAddress string `json:"ip_address" gorm:"size:64"`
	Location  string `json:"location" gorm:"size:100"`

	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }
