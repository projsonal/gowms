package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/projsonal/gowms/docs"
	"github.com/projsonal/gowms/internal/controller/appinfo"
	"github.com/projsonal/gowms/internal/routes"
	"github.com/projsonal/gowms/pkg/config"
)

func main() {
	cfg := config.Load()

	docs.SwaggerInfo.BasePath = "/wms-rsd"
	docs.SwaggerInfo.Title = cfg.App.Name
	docs.SwaggerInfo.Version = appinfo.CurrentVersion

	db := config.NewDatabase(cfg)
	if err := config.AutoMigrate(db); err != nil {
		log.Fatalf("gagal migrasi database: %v", err)
	}
	if err := config.SeedDefaultRoles(db); err != nil {
		log.Fatalf("gagal seed role default: %v", err)
	}
	if err := config.SeedDefaultPermissions(db); err != nil {
		log.Fatalf("gagal seed izin default: %v", err)
	}

	deps := routes.New(db, cfg)
	app := routes.SetupRouter(deps)

	pingInterval := 60 * time.Second
	if raw := os.Getenv("ASSET_PING_INTERVAL_SECONDS"); raw != "" {
		if secs, err := strconv.Atoi(raw); err == nil && secs >= 0 {
			pingInterval = time.Duration(secs) * time.Second
		} else {
			log.Printf("ASSET_PING_INTERVAL_SECONDS tidak valid (%q), pakai default 60 detik", raw)
		}
	}
	deps.AssetController.StartAutoPingScheduler(pingInterval)

	log.Printf("%s versi %s berjalan di %s (env: %s)", cfg.App.Name, appinfo.CurrentVersion, cfg.App.ListenAddress(), cfg.App.Env)
	if err := app.Listen(cfg.App.ListenAddress()); err != nil {
		log.Fatalf("gagal menjalankan server: %v", err)
	}
}
