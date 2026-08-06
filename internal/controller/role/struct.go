package role

import (
	roleRepo "github.com/projsonal/gowms/internal/repositories/role"
	"github.com/projsonal/gowms/pkg/utils"
)

type Controller struct {
	repo   roleRepo.Repository
	jwtSvc *utils.JWTService
}

func New(repo roleRepo.Repository, jwtSvc *utils.JWTService) *Controller {
	return &Controller{repo: repo, jwtSvc: jwtSvc}
}

type CreateRoleRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=50"`
	Description string `json:"description" validate:"max=255"`
}

type UpdatePermissionMatrixRequest struct {
	Items []ModulePermissionItem `json:"items" validate:"required,dive"`
}

type ModulePermissionItem struct {
	Module         string `json:"module" validate:"required"`
	View           bool   `json:"view"`
	Tambah         bool   `json:"tambah"`
	Edit           bool   `json:"edit"`
	ApprovalReject bool   `json:"approval_reject"`
	Print          bool   `json:"print"`
	AssignDelegasi bool   `json:"assign_delegasi"`
}
