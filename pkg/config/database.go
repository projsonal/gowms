package config

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/projsonal/gostock/internal/model"
	"github.com/projsonal/gostock/pkg/constant"
)

func NewDatabase(cfg *Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		log.Fatalf("gagal konek database: %v", err)
	}
	return db
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Role{},
		&model.Permission{},
		&model.RolePermission{},
		&model.User{},
		&model.RefreshToken{},

		&model.Kategori{},
		&model.Satuan{},
		&model.Gudang{},
		&model.Rak{},

		&model.Barang{},
		&model.Supplier{},
		&model.PurchaseOrder{},
		&model.PurchaseOrderItem{},
		&model.BarangMasuk{},
		&model.BarangMasukItem{},
		&model.BarangKeluar{},
		&model.BarangKeluarItem{},
		&model.StockOpname{},
		&model.StockOpnameItem{},
		&model.Pengiriman{},
		&model.PengirimanTrackingPoint{},
		&model.MaintenanceStatus{},
	)
}

func SeedDefaultRoles(db *gorm.DB) error {
	defaults := []model.Role{
		{Name: constant.RoleSuperAdmin, Description: "Akses penuh seluruh modul sistem", IsSystem: true},
		{Name: constant.RoleAdmin, Description: "Kelola operasional gudang & pengguna", IsSystem: true},
		{Name: constant.RoleKaryawan, Description: "Role default akun self-register", IsSystem: true},
	}

	for _, r := range defaults {
		var existing model.Role
		err := db.Where("name = ?", r.Name).First(&existing).Error
		if err == nil {
			continue
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}
		if err := db.Create(&r).Error; err != nil {
			return err
		}
	}
	return nil
}
