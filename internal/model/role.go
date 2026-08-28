package model

import "time"

type Role struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex;size:50;not null"`
	Description string    `json:"description" gorm:"size:255"`
	IsSystem    bool      `json:"is_system" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Role) TableName() string { return "roles" }

type Permission struct {
	ID     uint   `json:"id" gorm:"primaryKey"`
	Module string `json:"module" gorm:"size:50;not null;uniqueIndex:idx_module_action"`
	Action string `json:"action" gorm:"size:30;not null;uniqueIndex:idx_module_action"`
}

func (Permission) TableName() string { return "permissions" }

type RolePermission struct {
	ID           uint `json:"id" gorm:"primaryKey"`
	RoleID       uint `json:"role_id" gorm:"uniqueIndex:idx_role_permission"`
	PermissionID uint `json:"permission_id" gorm:"uniqueIndex:idx_role_permission"`
}

func (RolePermission) TableName() string { return "role_permissions" }
