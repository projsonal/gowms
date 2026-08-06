package main

import (
	"log"

	"github.com/projsonal/gostock/internal/routes"
	"github.com/projsonal/gostock/pkg/config"
)

func main() {
	cfg := config.Load()

	db := config.NewDatabase(cfg)
	if err := config.AutoMigrate(db); err != nil {
		log.Fatalf("gagal migrasi database: %v", err)
	}
	if err := config.SeedDefaultRoles(db); err != nil {
		log.Fatalf("gagal seed role default: %v", err)
	}

	deps := routes.New(db, cfg)
	app := routes.SetupRouter(deps)

	log.Printf("%s berjalan di %s (env: %s)", cfg.App.Name, cfg.App.ListenAddress(), cfg.App.Env)
	if err := app.Listen(cfg.App.ListenAddress()); err != nil {
		log.Fatalf("gagal menjalankan server: %v", err)
	}
}
