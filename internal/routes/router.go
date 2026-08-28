package routes

import (
	"log"

	"time"

	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/gofiber/swagger"

	"github.com/projsonal/gowms/internal/controller"
	"github.com/projsonal/gowms/internal/middleware"
	"github.com/projsonal/gowms/pkg/utils"
)

func SetupRouter(deps *Dependencies) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:                 deps.Cfg.App.Name,
		ErrorHandler:            globalErrorHandler,
		BodyLimit:               4 * 1024 * 1024,
		ReadTimeout:             15 * time.Second,
		WriteTimeout:            15 * time.Second,
		IdleTimeout:             60 * time.Second,
		EnableTrustedProxyCheck: true,
		TrustedProxies:          deps.Cfg.App.TrustedProxies,
		ProxyHeader:             fiber.HeaderXForwardedFor,
	})

	app.Use(middleware.Recover())
	app.Use(middleware.RequestLogger())
	app.Use(middleware.SecurityHeaders())
	app.Use(middleware.CORS(deps.Cfg))
	app.Use(middleware.GlobalRateLimiter())

	app.Get("/health/live", deps.HealthController.Live)
	app.Get("/health/ready", deps.HealthController.Ready)
	app.Get("/health", deps.HealthController.Health)

	app.Static("/uploads", deps.Cfg.Storage.Path)

	if deps.Cfg.Swagger.Enabled {
		swaggerRoute := app.Group("/swagger")
		if deps.Cfg.Swagger.BasicAuthUser != "" {
			swaggerRoute.Use(middleware.SwaggerBasicAuth(deps.Cfg.Swagger.BasicAuthUser, deps.Cfg.Swagger.BasicAuthPass))
		}
		swaggerRoute.Get("/*", fiberSwagger.HandlerDefault)
	}

	api := app.Group("/stockrsd")

	deps.CaptchaController.RegisterRoutes(api)
	deps.HumanCheckController.RegisterRoutes(api)
	deps.SecurityController.RegisterRoutes(api)
	deps.AppInfoController.RegisterRoutes(api)
	deps.TrashController.RegisterRoutes(api)
	deps.NotificationController.RegisterRoutes(api)

	deps.MaintenanceController.RegisterRoutes(api)

	if deps.Cfg.BotCheck.Enabled {
		api.Use(middleware.BotCheck(deps.BotCheckSvc))
	}

	api.Use("/auth/login", middleware.LoginRateLimiter())

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
		deps.BarangMasukController,
		deps.BarangKeluarController,
		deps.BarangSerialController,
		deps.StockOpnameController,
		deps.AssetController,
		deps.BarangRusakController,
		deps.TaskController,
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
	message := "terjadi kesalahan pada server, silakan coba lagi"
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code

		if code < fiber.StatusInternalServerError {
			message = e.Message
		}
	} else {

		log.Printf("router: unhandled error di %s %s: %v", c.Method(), c.Path(), err)
	}
	return utils.Fail(c, code, message, nil)
}
