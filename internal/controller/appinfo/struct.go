package appinfo

import (
	maintenanceRepo "github.com/projsonal/gowms/internal/repositories/maintenance"
	notificationRepo "github.com/projsonal/gowms/internal/repositories/notification"
	"github.com/projsonal/gowms/pkg/config"
	"github.com/projsonal/gowms/pkg/utils"
)

// Controller menangani endpoint info aplikasi (versi, changelog) DAN alur
// "Cek Update"/"Update Sekarang" di Settings > Sistem (lihat
// update_controller.go). Butuh:
//   - cfg: kredensial repo GitHub & path skrip deploy (lihat
//     pkg/config SelfUpdateConfig)
//   - jwtSvc: melindungi endpoint update (harus login; TriggerUpdate
//     khusus super_admin) — lihat RegisterRoutes
//   - maintenanceRepo: menyalakan Mode Pemeliharaan otomatis selama
//     proses update berjalan
//   - notifRepo: memberi tahu Super Admin lain begitu update
//     selesai/gagal
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
