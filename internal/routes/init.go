package routes

import (
	"log"
	"time"

	"gorm.io/gorm"

	appinfoController "github.com/projsonal/gowms/internal/controller/appinfo"
	assetController "github.com/projsonal/gowms/internal/controller/asset_gudang"
	authController "github.com/projsonal/gowms/internal/controller/auth"
	barangController "github.com/projsonal/gowms/internal/controller/barang"
	barangKeluarController "github.com/projsonal/gowms/internal/controller/barang_keluar"
	barangMasukController "github.com/projsonal/gowms/internal/controller/barang_masuk"
	barangRusakController "github.com/projsonal/gowms/internal/controller/barang_rusak"
	barangSerialController "github.com/projsonal/gowms/internal/controller/barang_serial"
	captchaController "github.com/projsonal/gowms/internal/controller/captcha"
	dashboardController "github.com/projsonal/gowms/internal/controller/dashboard"
	gudangController "github.com/projsonal/gowms/internal/controller/gudang"
	humanCheckController "github.com/projsonal/gowms/internal/controller/humancheck"
	laporanController "github.com/projsonal/gowms/internal/controller/laporan"
	maintenanceController "github.com/projsonal/gowms/internal/controller/maintenance"
	notificationController "github.com/projsonal/gowms/internal/controller/notification"
	roleController "github.com/projsonal/gowms/internal/controller/role"
	securityController "github.com/projsonal/gowms/internal/controller/security"
	stockOpnameController "github.com/projsonal/gowms/internal/controller/stockOpname"
	taskController "github.com/projsonal/gowms/internal/controller/task"
	trashController "github.com/projsonal/gowms/internal/controller/trash"
	usersController "github.com/projsonal/gowms/internal/controller/users"
	"github.com/projsonal/gowms/internal/health"
	assetRepo "github.com/projsonal/gowms/internal/repositories/asset"
	assetHistoryRepo "github.com/projsonal/gowms/internal/repositories/asset_history"
	assetPortRepo "github.com/projsonal/gowms/internal/repositories/asset_port"
	authRepo "github.com/projsonal/gowms/internal/repositories/auth"
	barangRepo "github.com/projsonal/gowms/internal/repositories/barang"
	barangKeluarRepo "github.com/projsonal/gowms/internal/repositories/barang_keluar"
	barangMasukRepo "github.com/projsonal/gowms/internal/repositories/barang_masuk"
	barangRusakRepo "github.com/projsonal/gowms/internal/repositories/barang_rusak"
	barangSerialRepo "github.com/projsonal/gowms/internal/repositories/barang_serial"
	gudangRepo "github.com/projsonal/gowms/internal/repositories/gudang"
	maintenanceRepo "github.com/projsonal/gowms/internal/repositories/maintenance"
	notificationRepo "github.com/projsonal/gowms/internal/repositories/notification"
	roleRepo "github.com/projsonal/gowms/internal/repositories/role"
	stockOpnameRepo "github.com/projsonal/gowms/internal/repositories/stockOpname"
	taskRepo "github.com/projsonal/gowms/internal/repositories/task"
	usersRepo "github.com/projsonal/gowms/internal/repositories/users"
	"github.com/projsonal/gowms/pkg/botcheck"
	"github.com/projsonal/gowms/pkg/captcha"
	"github.com/projsonal/gowms/pkg/config"
	"github.com/projsonal/gowms/pkg/geoip"
	"github.com/projsonal/gowms/pkg/humancheck"
	"github.com/projsonal/gowms/pkg/utils"
)

type Dependencies struct {
	Cfg        *config.Config
	JWTSvc     *utils.JWTService
	RoleRepo   roleRepo.Repository
	GudangRepo gudangRepo.Repository

	MaintenanceRepo maintenanceRepo.Repository

	AuthController *authController.Controller
	UserController *usersController.Controller
	RoleController *roleController.Controller

	GudangController *gudangController.Controller
	BarangController *barangController.Controller

	BarangMasukController  *barangMasukController.Controller
	BarangKeluarController *barangKeluarController.Controller
	BarangSerialController *barangSerialController.Controller
	StockOpnameController  *stockOpnameController.Controller
	AssetController        *assetController.Controller
	BarangRusakController  *barangRusakController.Controller
	TaskController         *taskController.Controller
	AppInfoController      *appinfoController.Controller
	TrashController        *trashController.Controller
	NotificationController *notificationController.Controller
	NotificationRepo       notificationRepo.Repository

	LaporanController   *laporanController.Controller
	DashboardController *dashboardController.Controller

	CaptchaController     *captchaController.Controller
	HumanCheckController  *humanCheckController.Controller
	SecurityController    *securityController.Controller
	MaintenanceController *maintenanceController.Controller

	BotCheckSvc *botcheck.Service

	HealthController *health.Controller
}

func New(db *gorm.DB, cfg *config.Config) *Dependencies {
	jwtSvc := utils.NewJWTService(&cfg.JWT)

	// Repositories.
	rRole := roleRepo.New(db)
	rUsers := usersRepo.New(db)
	rNotification := notificationRepo.New(db)
	rAuth := authRepo.New(db)
	rGudang := gudangRepo.New(db)
	rBarang := barangRepo.New(db)
	rBarangMasuk := barangMasukRepo.New(db)
	rBarangKeluar := barangKeluarRepo.New(db)
	rBarangSerial := barangSerialRepo.New(db)
	rStockOpname := stockOpnameRepo.New(db)
	rAsset := assetRepo.New(db)
	rAssetPort := assetPortRepo.New(db)
	rAssetHistory := assetHistoryRepo.New(db)
	rBarangRusak := barangRusakRepo.New(db)
	rTask := taskRepo.New(db)
	rMaintenance := maintenanceRepo.New(db)

	// Services lintas modul.
	captchaSvc := captcha.NewService(cfg.Captcha.Secret, time.Duration(cfg.Captcha.TTLMinutes)*time.Minute)
	humanCheckSvc := humancheck.NewService(
		cfg.HumanCheck.Secret,
		time.Duration(cfg.HumanCheck.TTLMinutes)*time.Minute,
		time.Duration(cfg.HumanCheck.MinDelaySeconds)*time.Second,
	)
	botCheckSvc := botcheck.NewService(cfg.BotCheck.Secret, time.Duration(cfg.BotCheck.WindowMinutes)*time.Minute)
	geoipSvc := newGeoIPResolver(cfg)

	// Controllers.
	cAuth := authController.New(authController.Params{
		AuthRepo:      rAuth,
		UserRepo:      rUsers,
		RoleRepo:      rRole,
		JWTSvc:        jwtSvc,
		CaptchaSvc:    captchaSvc,
		HumanCheckSvc: humanCheckSvc,
		Cfg:           cfg,
		GeoipSvc:      geoipSvc,
	})
	cUsers := usersController.New(usersController.Params{
		UserRepo:      rUsers,
		RoleRepo:      rRole,
		AuthRepo:      rAuth,
		JWTSvc:        jwtSvc,
		HumanCheckSvc: humanCheckSvc,
		StoragePath:   cfg.Storage.Path,
	})
	cRole := roleController.New(rRole, jwtSvc)
	cGudang := gudangController.New(rGudang, rRole, jwtSvc)
	cBarang := barangController.New(rBarang, rGudang, rRole, jwtSvc)
	cBarangMasuk := barangMasukController.New(rBarangMasuk, rBarang, rGudang, rRole, jwtSvc, rNotification)
	cBarangKeluar := barangKeluarController.New(rBarangKeluar, rBarang, rGudang, rRole, jwtSvc, rNotification)
	cBarangSerial := barangSerialController.New(rBarangSerial, rBarang, rRole, jwtSvc)
	cStockOpname := stockOpnameController.New(rStockOpname, rBarang, rGudang, rRole, jwtSvc, rNotification)
	cAsset := assetController.New(rAsset, rGudang, rAssetPort, rAssetHistory, rUsers, rRole, jwtSvc, rNotification)
	cBarangRusak := barangRusakController.New(rBarangRusak, rBarang, rRole, jwtSvc, cfg.Storage.Path, rNotification)
	cTask := taskController.New(rTask, rRole, jwtSvc)
	cLaporan := laporanController.New(rBarang, rBarangMasuk, rBarangKeluar, rStockOpname, rBarangRusak, rRole, jwtSvc)
	cDashboard := dashboardController.New(rBarang, rGudang, rBarangMasuk, rBarangKeluar, rStockOpname, rRole, jwtSvc, db)
	cCaptcha := captchaController.New(captchaSvc)
	cHumanCheck := humanCheckController.New(humanCheckSvc)
	cSecurity := securityController.New(botCheckSvc, captchaSvc)
	cMaintenance := maintenanceController.New(rMaintenance, jwtSvc, rNotification)
	cHealth := health.NewController(health.NewChecker(db, cfg.Storage.Path))

	return &Dependencies{
		Cfg:                    cfg,
		JWTSvc:                 jwtSvc,
		RoleRepo:               rRole,
		GudangRepo:             rGudang,
		MaintenanceRepo:        rMaintenance,
		AuthController:         cAuth,
		UserController:         cUsers,
		RoleController:         cRole,
		GudangController:       cGudang,
		BarangController:       cBarang,
		BarangMasukController:  cBarangMasuk,
		BarangKeluarController: cBarangKeluar,
		BarangSerialController: cBarangSerial,
		StockOpnameController:  cStockOpname,
		AssetController:        cAsset,
		BarangRusakController:  cBarangRusak,
		TaskController:         cTask,
		AppInfoController:      appinfoController.New(cfg, jwtSvc, rMaintenance, rNotification),
		TrashController:        trashController.New(db, jwtSvc),
		NotificationController: notificationController.New(rNotification, jwtSvc),
		NotificationRepo:       rNotification,
		LaporanController:      cLaporan,
		DashboardController:    cDashboard,
		CaptchaController:      cCaptcha,
		HumanCheckController:   cHumanCheck,
		SecurityController:     cSecurity,
		MaintenanceController:  cMaintenance,
		BotCheckSvc:            botCheckSvc,
		HealthController:       cHealth,
	}
}

func newGeoIPResolver(cfg *config.Config) geoip.Resolver {
	if !cfg.GeoIP.Enabled {
		return geoip.NoopResolver{}
	}

	resolver, err := geoip.NewHTTPResolver(cfg.GeoIP.BaseURL)
	if err != nil {
		log.Printf("geoip: konfigurasi GEOIP_BASE_URL tidak valid, fallback ke NoopResolver: %v", err)
		return geoip.NoopResolver{}
	}
	return resolver
}
