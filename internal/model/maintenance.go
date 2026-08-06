package model

import "time"

type MaintenanceStatus struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	IsActive       bool       `json:"is_active" gorm:"not null;default:false"`
	Message        string     `json:"message" gorm:"size:500"`
	StartedAt      *time.Time `json:"started_at"`
	EstimatedUntil *time.Time `json:"estimated_until"`
	UpdatedBy      *uint      `json:"updated_by"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (MaintenanceStatus) TableName() string { return "maintenance_status" }
