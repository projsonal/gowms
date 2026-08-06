package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gostock/internal/controller"
	"github.com/projsonal/gostock/internal/middleware"
	"github.com/projsonal/gostock/pkg/utils"
)

func SetupRouter(deps *Dependencies) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      deps.Cfg.App.Name,
		ErrorHandler: globalErrorHandler,
	})

	app.Use(middleware.Recover())
	app.Use(middleware.RequestLogger())
	app.Use(middleware.CORS(deps.Cfg))

	app.Get("/health/live", deps.HealthController.Live)
	app.Get("/health/ready", deps.HealthController.Ready)
	app.Get("/health", deps.HealthController.Health)

	api := app.Group("/stockrsd")

	deps.CaptchaController.RegisterRoutes(api)
	deps.SecurityController.RegisterRoutes(api)

	deps.MaintenanceController.RegisterRoutes(api)

	api.Use(middleware.BotCheck(deps.BotCheckSvc))

	api.Use("/auth/login", middleware.LoginRateLimiter())

	// Modul yang TETAP bisa diakses walau mode maintenance aktif: auth
	// (supaya user termasuk super_admin tetap bisa login) & manajemen
	// user/role (tugas administratif, bukan operasional harian).
	alwaysOn := []controller.RouteRegistrar{
		deps.AuthController,
		deps.UserController,
		deps.RoleController,
	}
	for _, r := range alwaysOn {
		r.RegisterRoutes(api)
	}

	operational := api.Group("/", middleware.MaintenanceMode(deps.MaintenanceRepo, deps.JWTSvc))
	operationalRegistrars := []controller.RouteRegistrar{
		deps.GudangController,
		deps.BarangController,
		deps.SupplierController,
		deps.PurchaseOrderController,
		deps.BarangMasukController,
		deps.BarangKeluarController,
		deps.StockOpnameController,
		deps.PengirimanController,
		deps.LaporanController,
		deps.DashboardController,
	}
	for _, r := range operationalRegistrars {
		r.RegisterRoutes(operational)
	}

	app.Use(utils.NotFound)

	return app
}

func globalErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	return utils.Fail(c, code, err.Error(), nil)
}
