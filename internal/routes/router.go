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
		AppName:      deps.Cfg.App.Name,
		ErrorHandler: globalErrorHandler,
		// Mitigasi DDoS/slowloris dasar di level aplikasi: batasi ukuran
		// body request & waktu baca/tulis koneksi, supaya klien nakal
		// tidak bisa menahan koneksi terbuka lama atau mengirim payload
		// raksasa untuk menghabiskan resource server.
		BodyLimit:    4 * 1024 * 1024, // 4MB — lebih dari cukup untuk payload JSON biasa
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		// Di produksi, app ini SELALU berjalan di belakang reverse proxy
		// (nginx — lihat panduan CI/CD). Tanpa konfigurasi ini, c.IP()
		// akan selalu mengembalikan alamat nginx (biasanya 127.0.0.1),
		// BUKAN alamat IP pengunjung asli — yang menjelaskan kenapa kolom
		// IP/lokasi login selalu kosong/"-". Dengan ini, Fiber membaca IP
		// asli dari header X-Forwarded-For yang dikirim nginx, TAPI hanya
		// mempercayai header itu kalau permintaan datang dari proxy yang
		// terdaftar di TrustedProxies — mencegah klien memalsukan IP-nya
		// sendiri lewat header itu.
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

	// File lain yang masih diupload ke disk lokal (mis. foto bukti
	// barang_rusak) disajikan statis dari sini. Foto profil user SUDAH
	// TIDAK lewat sini lagi — sekarang disimpan sebagai bytea di database
	// & diserve lewat GET /users/:id/avatar yang wajib login (lihat
	// internal/controller/users/user_controller.go ServeAvatar), karena
	// route statis ini tidak punya proteksi/otentikasi sama sekali. Kalau
	// foto barang_rusak juga dianggap sensitif, pola yang sama (simpan di
	// DB + endpoint ber-auth) bisa dipakai di controller barang_rusak.
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
		// fiber.Error di sini SELALU berasal dari error yang sengaja
		// dilempar handler/router Fiber sendiri (mis. body terlalu besar,
		// method tidak diizinkan) — bukan error internal Go yang bisa
		// membawa detail implementasi (path file, query SQL, dsb). Untuk
		// status 5xx tetap pakai pesan generik; untuk 4xx aman ditampilkan
		// karena pesannya memang ditujukan untuk klien.
		if code < fiber.StatusInternalServerError {
			message = e.Message
		}
	} else {
		// Error non-fiber.Error berarti panic/error Go mentah yang lolos
		// sampai sini — JANGAN pernah balas err.Error() ke klien: bisa
		// membocorkan detail internal (path, driver DB, dsb). Cukup log
		// di server, klien cukup dapat pesan generik.
		log.Printf("router: unhandled error di %s %s: %v", c.Method(), c.Path(), err)
	}
	return utils.Fail(c, code, message, nil)
}
