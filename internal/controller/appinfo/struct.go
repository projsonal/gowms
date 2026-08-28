package appinfo

import (
	maintenanceRepo "github.com/projsonal/gowms/internal/repositories/maintenance"
	notificationRepo "github.com/projsonal/gowms/internal/repositories/notifikasi"
	"github.com/projsonal/gowms/pkg/config"
	"github.com/projsonal/gowms/pkg/utils"
)

type Controller struct {
	cfg             *config.Config
	jwtSvc          *utils.JWTService
	maintenanceRepo maintenanceRepo.Repository
	notifRepo       notificationRepo.Repository
}

func New(
	cfg *config.Config,
	jwtSvc *utils.JWTService,
	maintenanceRepo maintenanceRepo.Repository,
	notifRepo notificationRepo.Repository,
) *Controller {
	return &Controller{
		cfg:             cfg,
		jwtSvc:          jwtSvc,
		maintenanceRepo: maintenanceRepo,
		notifRepo:       notifRepo,
	}
}
