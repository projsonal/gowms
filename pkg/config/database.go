package config

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/constant"
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
	if err := db.AutoMigrate(
		&model.Role{},
		&model.Permission{},
		&model.RolePermission{},
		&model.User{},
		&model.RefreshToken{},

		&model.Kategori{},
		&model.Satuan{},
		&model.Gudang{},

		&model.Barang{},
		&model.BarangMasuk{},
		&model.BarangMasukItem{},
		&model.BarangKeluar{},
		&model.BarangKeluarItem{},
		&model.BarangSerial{},
		&model.BarangStokGudang{},
		&model.StockOpname{},
		&model.StockOpnameItem{},
		&model.Asset{},
		&model.AssetPort{},
		&model.AssetHistory{},
		&model.BarangRusak{},
		&model.MaintenanceStatus{},
		&model.Notification{},
		&model.NotificationRead{},
		&model.NotificationDismissed{},
	); err != nil {
		return err
	}
	if err := ensureAssetPartialUniqueIndexes(db); err != nil {
		return err
	}
	return ensureUsersEmailNotUnique(db)
}

func ensureUsersEmailNotUnique(db *gorm.DB) error {
	stmt := `DO $$
	DECLARE r record;
	BEGIN
		FOR r IN
			SELECT indexname FROM pg_indexes
			WHERE tablename = 'users'
			  AND indexdef ILIKE '%UNIQUE INDEX%'
			  AND indexdef ILIKE '%(email)%'
		LOOP
			EXECUTE 'DROP INDEX IF EXISTS ' || quote_ident(r.indexname);
		END LOOP;
	END $$;`
	return db.Exec(stmt).Error
}

func ensureAssetPartialUniqueIndexes(db *gorm.DB) error {
	stmts := []string{

		`DO $$
		DECLARE r record;
		BEGIN
			FOR r IN
				SELECT indexname FROM pg_indexes
				WHERE tablename = 'assets'
				  AND indexdef ILIKE '%UNIQUE INDEX%'
				  AND indexdef NOT ILIKE '% WHERE %'
				  AND (indexdef ILIKE '%(label_rsd)%' OR indexdef ILIKE '%(kode_ba)%')
			LOOP
				EXECUTE 'DROP INDEX IF EXISTS ' || quote_ident(r.indexname);
			END LOOP;
		END $$;`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_assets_label_rsd_unique ON assets (label_rsd) WHERE label_rsd <> ''`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_assets_kode_ba_unique ON assets (kode_ba) WHERE kode_ba <> ''`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			return err
		}
	}
	return nil
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
