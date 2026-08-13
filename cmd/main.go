package main

import (
	"log"

	"github.com/projsonal/gowms/docs"
	"github.com/projsonal/gowms/internal/controller/appinfo"
	"github.com/projsonal/gowms/internal/routes"
	"github.com/projsonal/gowms/pkg/config"
)

func main() {
	cfg := config.Load()

	docs.SwaggerInfo.BasePath = "/stockrsd"
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

	// Cetak versi di log startup — cara cepat mengecek dari terminal/log
	// server (tanpa perlu buka browser ke /app/version) apakah binary
	// yang baru saja dijalankan ini benar hasil build source terbaru.
	log.Printf("%s versi %s berjalan di %s (env: %s)", cfg.App.Name, appinfo.CurrentVersion, cfg.App.ListenAddress(), cfg.App.Env)
	if err := app.Listen(cfg.App.ListenAddress()); err != nil {
		log.Fatalf("gagal menjalankan server: %v", err)
	}
}
