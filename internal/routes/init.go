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
	captchaController "github.com/projsonal/gowms/internal/controller/captcha"
	codController "github.com/projsonal/gowms/internal/controller/cod"
	dashboardController "github.com/projsonal/gowms/internal/controller/dashboard"
	gudangController "github.com/projsonal/gowms/internal/controller/gudang"
	laporanController "github.com/projsonal/gowms/internal/controller/laporan"
	maintenanceController "github.com/projsonal/gowms/internal/controller/maintenance"
	pengirimanController "github.com/projsonal/gowms/internal/controller/pengiriman"
	purchaseOrderController "github.com/projsonal/gowms/internal/controller/po"
	roleController "github.com/projsonal/gowms/internal/controller/role"
	securityController "github.com/projsonal/gowms/internal/controller/security"
	stockOpnameController "github.com/projsonal/gowms/internal/controller/stockOpname"
	supplierController "github.com/projsonal/gowms/internal/controller/supplier"
	usersController "github.com/projsonal/gowms/internal/controller/users"
	"github.com/projsonal/gowms/internal/health"
	assetRepo "github.com/projsonal/gowms/internal/repositories/asset"
	authRepo "github.com/projsonal/gowms/internal/repositories/auth"
	barangRepo "github.com/projsonal/gowms/internal/repositories/barang"
	barangKeluarRepo "github.com/projsonal/gowms/internal/repositories/barang_keluar"
	barangMasukRepo "github.com/projsonal/gowms/internal/repositories/barang_masuk"
	barangRusakRepo "github.com/projsonal/gowms/internal/repositories/barang_rusak"
	codRepo "github.com/projsonal/gowms/internal/repositories/cod"
	gudangRepo "github.com/projsonal/gowms/internal/repositories/gudang"
	maintenanceRepo "github.com/projsonal/gowms/internal/repositories/maintenance"
	pengirimanRepo "github.com/projsonal/gowms/internal/repositories/pengiriman"
	purchaseOrderRepo "github.com/projsonal/gowms/internal/repositories/po"
	roleRepo "github.com/projsonal/gowms/internal/repositories/role"
	stockOpnameRepo "github.com/projsonal/gowms/internal/repositories/stockOpname"
	supplierRepo "github.com/projsonal/gowms/internal/repositories/supplier"
	usersRepo "github.com/projsonal/gowms/internal/repositories/users"
	"github.com/projsonal/gowms/pkg/botcheck"
	"github.com/projsonal/gowms/pkg/captcha"
	"github.com/projsonal/gowms/pkg/config"
	"github.com/projsonal/gowms/pkg/geoip"
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

	GudangController   *gudangController.Controller
	BarangController   *barangController.Controller
	SupplierController *supplierController.Controller

	PurchaseOrderController *purchaseOrderController.Controller
	BarangMasukController   *barangMasukController.Controller
	BarangKeluarController  *barangKeluarController.Controller
	StockOpnameController   *stockOpnameController.Controller
	PengirimanController    *pengirimanController.Controller
	CodController           *codController.Controller
	AssetController         *assetController.Controller
	BarangRusakController   *barangRusakController.Controller
	AppInfoController       *appinfoController.Controller

	LaporanController   *laporanController.Controller
	DashboardController *dashboardController.Controller

	CaptchaController     *captchaController.Controller
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
	rAuth := authRepo.New(db)
	rGudang := gudangRepo.New(db)
	rBarang := barangRepo.New(db)
	rSupplier := supplierRepo.New(db)
	rPO := purchaseOrderRepo.New(db)
	rBarangMasuk := barangMasukRepo.New(db, rPO)
	rBarangKeluar := barangKeluarRepo.New(db)
	rStockOpname := stockOpnameRepo.New(db)
	rPengiriman := pengirimanRepo.New(db)
	rCod := codRepo.New(db)
	rAsset := assetRepo.New(db)
	rBarangRusak := barangRusakRepo.New(db)
	rMaintenance := maintenanceRepo.New(db)

	// Services lintas modul.
	captchaSvc := captcha.NewService(cfg.Captcha.Secret, time.Duration(cfg.Captcha.TTLMinutes)*time.Minute)
	botCheckSvc := botcheck.NewService(cfg.BotCheck.Secret, time.Duration(cfg.BotCheck.WindowMinutes)*time.Minute)
	geoipSvc := newGeoIPResolver(cfg)

	// Controllers.
	cAuth := authController.New(authController.Params{
		AuthRepo:   rAuth,
		UserRepo:   rUsers,
		RoleRepo:   rRole,
		JWTSvc:     jwtSvc,
		CaptchaSvc: captchaSvc,
		Cfg:        cfg,
		GeoipSvc:   geoipSvc,
	})
	cUsers := usersController.New(usersController.Params{
		UserRepo:    rUsers,
		RoleRepo:    rRole,
		AuthRepo:    rAuth,
		JWTSvc:      jwtSvc,
		CaptchaSvc:  captchaSvc,
		StoragePath: cfg.Storage.Path,
	})
	cRole := roleController.New(rRole, jwtSvc)
	cGudang := gudangController.New(rGudang, rRole, jwtSvc)
	cBarang := barangController.New(rBarang, rGudang, rRole, jwtSvc)
	cSupplier := supplierController.New(rSupplier, rRole, jwtSvc)
	cPO := purchaseOrderController.New(rPO, rSupplier, rRole, jwtSvc)
	cBarangMasuk := barangMasukController.New(rBarangMasuk, rBarang, rGudang, rPO, rSupplier, rRole, jwtSvc)
	cBarangKeluar := barangKeluarController.New(rBarangKeluar, rBarang, rGudang, rRole, jwtSvc)
	cStockOpname := stockOpnameController.New(rStockOpname, rBarang, rGudang, rRole, jwtSvc)
	cPengiriman := pengirimanController.New(rPengiriman, rGudang, rBarangKeluar, rRole, jwtSvc)
	cCod := codController.New(rCod, rRole, jwtSvc)
	cAsset := assetController.New(rAsset, rGudang, rRole, jwtSvc)
	cBarangRusak := barangRusakController.New(rBarangRusak, rBarang, rRole, jwtSvc)
	cLaporan := laporanController.New(rBarang, rPO, rBarangMasuk, rBarangKeluar, rStockOpname, rRole, jwtSvc)
	cDashboard := dashboardController.New(rBarang, rGudang, rSupplier, rPO, rBarangMasuk, rBarangKeluar, rStockOpname, rPengiriman, rRole, jwtSvc, db)
	cCaptcha := captchaController.New(captchaSvc)
	cSecurity := securityController.New(botCheckSvc, captchaSvc)
	cMaintenance := maintenanceController.New(rMaintenance, jwtSvc)
	cHealth := health.NewController(health.NewChecker(db, cfg.Storage.Path))

	return &Dependencies{
		Cfg:                     cfg,
		JWTSvc:                  jwtSvc,
		RoleRepo:                rRole,
		GudangRepo:              rGudang,
		MaintenanceRepo:         rMaintenance,
		AuthController:          cAuth,
		UserController:          cUsers,
		RoleController:          cRole,
		GudangController:        cGudang,
		BarangController:        cBarang,
		SupplierController:      cSupplier,
		PurchaseOrderController: cPO,
		BarangMasukController:   cBarangMasuk,
		BarangKeluarController:  cBarangKeluar,
		StockOpnameController:   cStockOpname,
		PengirimanController:    cPengiriman,
		CodController:           cCod,
		AssetController:         cAsset,
		BarangRusakController:   cBarangRusak,
		AppInfoController:       appinfoController.New(),
		LaporanController:       cLaporan,
		DashboardController:     cDashboard,
		CaptchaController:       cCaptcha,
		SecurityController:      cSecurity,
		MaintenanceController:   cMaintenance,
		BotCheckSvc:             botCheckSvc,
		HealthController:        cHealth,
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
