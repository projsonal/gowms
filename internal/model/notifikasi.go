package model

import "time"

type Notification struct {
	ID uint `json:"id" gorm:"primaryKey"`

	UserID     *uint  `json:"user_id" gorm:"index"`
	RoleTarget string `json:"role_target" gorm:"size:20;index"`

	Type    string `json:"type" gorm:"size:30;not null"`
	Title   string `json:"title" gorm:"size:150;not null"`
	Message string `json:"message" gorm:"size:500"`

	LinkHref string `json:"link_href" gorm:"size:255"`

	CreatedAt time.Time `json:"created_at"`
}

func (Notification) TableName() string { return "notifications" }

type NotificationRead struct {
	NotificationID uint      `json:"notification_id" gorm:"primaryKey"`
	UserID         uint      `json:"user_id" gorm:"primaryKey"`
	ReadAt         time.Time `json:"read_at"`
}

func (NotificationRead) TableName() string { return "notification_reads" }

type NotificationDismissed struct {
	NotificationID uint      `json:"notification_id" gorm:"primaryKey"`
	UserID         uint      `json:"user_id" gorm:"primaryKey"`
	DismissedAt    time.Time `json:"dismissed_at"`
}

func (NotificationDismissed) TableName() string { return "notification_dismissed" }
