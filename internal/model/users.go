// Package model berisi seluruh entitas GORM (tabel database) yang dipakai
// lintas layer repository & controller.
package model

import "time"

type User struct {
	ID                  uint       `json:"id" gorm:"primaryKey"`
	Username            string     `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Email               string     `json:"email" gorm:"uniqueIndex;size:100;not null"`
	PasswordHash        string     `json:"-" gorm:"size:255;not null"`
	FullName            string     `json:"full_name" gorm:"size:100"`
	PhoneNumber         string     `json:"phone_number" gorm:"size:20"`
	RoleID              uint       `json:"role_id" gorm:"not null;index"`
	Role                *Role      `json:"role,omitempty" gorm:"foreignKey:RoleID"`
	IsActive            bool       `json:"is_active" gorm:"default:true"`
	Is2FAEnabled        bool       `json:"is_2fa_enabled" gorm:"column:is_2fa_enabled;default:false"`
	TOTPSecret          string     `json:"-" gorm:"column:totp_secret;size:100"`
	FailedLoginAttempts int        `json:"-" gorm:"default:0"`
	LockedUntil         *time.Time `json:"-"`
	LastLoginAt         *time.Time `json:"last_login_at"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func (User) TableName() string { return "users" }
