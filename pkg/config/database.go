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
		&model.CodTransaction{},
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
	return ensureAssetPartialUniqueIndexes(db)
}

// ensureAssetPartialUniqueIndexes menegakkan keunikan label_rsd/kode_ba di
// tabel assets lewat PARTIAL unique index, bukan unique index penuh.
//
// Latar belakang: label_rsd hanya diisi untuk aset berkoordinat
// (tiang/odc/ont/odp/olt) dan kode_ba hanya diisi untuk aset transportasi —
// jadi salah satu kolom SELALU kosong ("") tergantung jenis aset. Unique
// index penuh ala `gorm:"uniqueIndex"` memperlakukan "" sebagai nilai biasa
// (BUKAN seperti NULL yang boleh berulang di Postgres), sehingga aset kedua
// dari jenis manapun akan "bentrok" dengan aset pertama yang kolom
// satunya sama-sama kosong — walau labelnya sendiri belum pernah dipakai.
// Ini penyebab utama error "nomor label aset ini kebetulan sudah dipakai"
// saat menambah/mengedit aset di Manajemen Gudang.
//
// Solusinya: index unik hanya berlaku untuk baris yang kolomnya TIDAK
// kosong (`WHERE label_rsd <> ''` / `WHERE kode_ba <> ''`). Baris dengan
// nilai kosong tidak ikut ditegakkan keunikannya sama sekali — sesuai
// maksud aslinya (kolom itu memang "tidak relevan" untuk jenis aset
// tersebut, bukan bagian dari data yang perlu unik).
//
// Dijalankan tiap startup, idempotent (aman dipanggil berkali-kali):
// pertama membuang unique index PENUH lama (peninggalan tag
// `gorm:"uniqueIndex"` sebelumnya, kalau skema DB sempat kebuat dengan
// versi lama), baru membuat partial unique index yang benar.
func ensureAssetPartialUniqueIndexes(db *gorm.DB) error {
	stmts := []string{
		// Buang index unik PENUH lama di kolom label_rsd/kode_ba, apa pun
		// namanya — dicari lewat definisi index-nya, bukan ditebak namanya,
		// supaya tetap aman kalau GORM/versi lama memberi nama berbeda.
		// Partial unique index yang kita buat sendiri di bawah TIDAK ikut
		// kebuang karena predicate "WHERE" tidak match filter ini.
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
