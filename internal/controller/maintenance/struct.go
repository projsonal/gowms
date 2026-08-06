package maintenance

import (
	"time"

	maintenanceRepo "github.com/projsonal/gostock/internal/repositories/maintenance"
	"github.com/projsonal/gostock/pkg/utils"
)

type Controller struct {
	repo   maintenanceRepo.Repository
	jwtSvc *utils.JWTService
}

func New(repo maintenanceRepo.Repository, jwtSvc *utils.JWTService) *Controller {
	return &Controller{repo: repo, jwtSvc: jwtSvc}
}

type SetRequest struct {
	IsActive       bool       `json:"is_active"`
	Message        string     `json:"message" validate:"max=500"`
	EstimatedUntil *time.Time `json:"estimated_until"`
}

type StatusResponse struct {
	IsActive         bool       `json:"is_active"`
	Message          string     `json:"message,omitempty"`
	StartedAt        *time.Time `json:"started_at,omitempty"`
	EstimatedUntil   *time.Time `json:"estimated_until,omitempty"`
	RemainingSeconds int64      `json:"remaining_seconds,omitempty"`
}
