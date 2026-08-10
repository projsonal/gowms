package task

import (
	"github.com/projsonal/gowms/internal/repositories/role"
	taskRepo "github.com/projsonal/gowms/internal/repositories/task"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

const Module = constant.ModuleTaskManagement

type Controller struct {
	repo     taskRepo.Repository
	roleRepo role.Repository
	jwtSvc   *utils.JWTService
}

func New(repo taskRepo.Repository, roleRepo role.Repository, jwtSvc *utils.JWTService) *Controller {
	return &Controller{repo: repo, roleRepo: roleRepo, jwtSvc: jwtSvc}
}

type TaskRequest struct {
	Title       string `json:"title" validate:"required,max=150"`
	Description string `json:"description" validate:"max=500"`
	AssignedTo  uint   `json:"assigned_to" validate:"required"`
	DueDate     string `json:"due_date" validate:"required"` // format YYYY-MM-DD
	Priority    string `json:"priority" validate:"required,oneof=rendah sedang tinggi"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=baru proses selesai"`
}

type SummaryResponse struct {
	Total    int64 `json:"total"`
	Proses   int64 `json:"proses"`
	Terlambat int64 `json:"terlambat"`
	Selesai  int64 `json:"selesai"`
}
