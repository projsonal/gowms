// Package role mengakses tabel roles, permissions, dan role_permissions —
// fondasi RBAC dinamis (matrix akses per modul per role).
package role

import "github.com/projsonal/gostock/internal/model"

// Repository mendefinisikan seluruh operasi data untuk Role & Permission.
// Method HasPermission juga dipakai middleware.PermissionChecker.
type Repository interface {
	FindAll() ([]model.Role, error)
	FindByID(id uint) (*model.Role, error)
	FindByName(name string) (*model.Role, error)
	Create(r *model.Role) error
	Update(r *model.Role) error
	Delete(id uint) error

	FindOrCreatePermission(module, action string) (*model.Permission, error)
	ReplaceRolePermissions(roleID uint, permissionIDs []uint) error
	HasPermission(roleID uint, module, action string) (bool, error)
	GetMatrix(roleID uint) ([]ModulePermission, error)
}

// ModulePermission adalah representasi ringkas 1 baris matrix akses untuk
// satu modul: Lihat | Tambah | Edit | Approval/Reject | Print | Assign/Delegasi.
type ModulePermission struct {
	Module         string `json:"module"`
	View           bool   `json:"view"`
	Tambah         bool   `json:"tambah"`
	Edit           bool   `json:"edit"`
	ApprovalReject bool   `json:"approval_reject"`
	Print          bool   `json:"print"`
	AssignDelegasi bool   `json:"assign_delegasi"`
}
