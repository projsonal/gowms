package role

import "github.com/projsonal/gowms/internal/model"

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

type ModulePermission struct {
	Module         string `json:"module"`
	View           bool   `json:"view"`
	Tambah         bool   `json:"tambah"`
	Edit           bool   `json:"edit"`
	ApprovalReject bool   `json:"approval_reject"`
	Print          bool   `json:"print"`
	AssignDelegasi bool   `json:"assign_delegasi"`
}
