package main

import (
	"log"

	// Import package docs (generated swaggo/swag) yang mendaftarkan spec
	// Swagger ke registry global lewat init(). TANPA baris ini, halaman
	// /swagger/index.html tampil kosong/putih total — swagger-ui berhasil
	// dimuat tapi tidak menemukan spec apa pun untuk dirender karena
	// docs.go tidak pernah ter-link ke binary (package Go yang tidak
	// diimpor di mana pun tidak akan dikompilasi masuk). Dipakai juga di
	// bawah untuk mengisi BasePath supaya tombol "Try it out" di Swagger
	// UI memanggil endpoint yang benar (/stockrsd/..., bukan /...).
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
