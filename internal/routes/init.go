package routes

import (
	"time"

	"gorm.io/gorm"

	authController "github.com/projsonal/gostock/internal/controller/auth"
	barangController "github.com/projsonal/gostock/internal/controller/barang"
	barangKeluarController "github.com/projsonal/gostock/internal/controller/barang_keluar"
	barangMasukController "github.com/projsonal/gostock/internal/controller/barang_masuk"
	captchaController "github.com/projsonal/gostock/internal/controller/captcha"
	dashboardController "github.com/projsonal/gostock/internal/controller/dashboard"
	gudangController "github.com/projsonal/gostock/internal/controller/gudang"
	laporanController "github.com/projsonal/gostock/internal/controller/laporan"
	maintenanceController "github.com/projsonal/gostock/internal/controller/maintenance"
	pengirimanController "github.com/projsonal/gostock/internal/controller/pengiriman"
	purchaseOrderController "github.com/projsonal/gostock/internal/controller/po"
	roleController "github.com/projsonal/gostock/internal/controller/role"
	securityController "github.com/projsonal/gostock/internal/controller/security"
	stockOpnameController "github.com/projsonal/gostock/internal/controller/stockOpname"
	supplierController "github.com/projsonal/gostock/internal/controller/supplier"
	usersController "github.com/projsonal/gostock/internal/controller/users"
	"github.com/projsonal/gostock/internal/health"
	authRepo "github.com/projsonal/gostock/internal/repositories/auth"
	barangRepo "github.com/projsonal/gostock/internal/repositories/barang"
	barangKeluarRepo "github.com/projsonal/gostock/internal/repositories/barang_keluar"
	barangMasukRepo "github.com/projsonal/gostock/internal/repositories/barang_masuk"
	gudangRepo "github.com/projsonal/gostock/internal/repositories/gudang"
	maintenanceRepo "github.com/projsonal/gostock/internal/repositories/maintenance"
	pengirimanRepo "github.com/projsonal/gostock/internal/repositories/pengiriman"
	purchaseOrderRepo "github.com/projsonal/gostock/internal/repositories/po"
	roleRepo "github.com/projsonal/gostock/internal/repositories/role"
	stockOpnameRepo "github.com/projsonal/gostock/internal/repositories/stockOpname"
	supplierRepo "github.com/projsonal/gostock/internal/repositories/supplier"
	usersRepo "github.com/projsonal/gostock/internal/repositories/users"
	"github.com/projsonal/gostock/pkg/botcheck"
	"github.com/projsonal/gostock/pkg/captcha"
	"github.com/projsonal/gostock/pkg/config"
	"github.com/projsonal/gostock/pkg/geoip"
	"github.com/projsonal/gostock/pkg/otp"
	"github.com/projsonal/gostock/pkg/utils"
	"github.com/projsonal/gostock/pkg/wa"
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
	rMaintenance := maintenanceRepo.New(db)

	// Services lintas modul.
	captchaSvc := captcha.NewService(cfg.Captcha.Secret, time.Duration(cfg.Captcha.TTLMinutes)*time.Minute)
	waOTPSvc := otp.NewService(cfg.WAOTP.Secret, time.Duration(cfg.WAOTP.TTLMinutes)*time.Minute)
	waSender := wa.NewClient(cfg.WhatsApp.APIURL, cfg.WhatsApp.APIKey, cfg.WhatsApp.Sender)
	botCheckSvc := botcheck.NewService(cfg.BotCheck.Secret, time.Duration(cfg.BotCheck.WindowMinutes)*time.Minute)
	geoipSvc := newGeoIPResolver(cfg)

	// Controllers.
	cAuth := authController.New(authController.Params{
		AuthRepo:   rAuth,
		UserRepo:   rUsers,
		RoleRepo:   rRole,
		JWTSvc:     jwtSvc,
		CaptchaSvc: captchaSvc,
		WaOTPSvc:   waOTPSvc,
		WaSender:   waSender,
		Cfg:        cfg,
		GeoipSvc:   geoipSvc,
	})
	cUsers := usersController.New(usersController.Params{
		UserRepo: rUsers,
		RoleRepo: rRole,
		JWTSvc:   jwtSvc,
		WaOTPSvc: waOTPSvc,
		WaSender: waSender,
		Cfg:      cfg,
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
	return geoip.NewHTTPResolver(cfg.GeoIP.BaseURL)
}
