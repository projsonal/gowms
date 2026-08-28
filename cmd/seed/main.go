package seed

import (
	"fmt"
	"log"
	"time"

	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/config"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
	"gorm.io/gorm"
)

const roleNameQuery = "name = ?"

var modules = []string{
	constant.ModuleDashboard, "kategori", "satuan", constant.ModuleManajemenGudang, "rak", "barang", constant.ModuleKelolaBarang,
	constant.ModuleBarangMasuk, constant.ModuleBarangKeluar,
	constant.ModuleStockOpname,
	constant.ModuleLaporan, constant.ModuleManajemenUser, constant.ModuleSettings, "notifikasi",
	constant.ModuleAsetGudang, constant.ModuleBarangRusak,
}
var actions = []string{constant.ActionView, constant.ActionTambah, constant.ActionEdit, constant.ActionApprovalReject, constant.ActionPrint, constant.ActionAssignDelegasi}

func upsertUser(db *gorm.DB, username, email, fullname, password, roleName string) uint {
	var role model.Role
	if err := db.Where(roleNameQuery, roleName).First(&role).Error; err != nil {
		log.Fatalf("role %s not found: %v", roleName, err)
	}
	hash, err := utils.HashPassword(password)
	if err != nil {
		log.Fatalf("hash err: %v", err)
	}
	var u model.User
	err = db.Where("username = ?", username).First(&u).Error
	if err == gorm.ErrRecordNotFound {
		u = model.User{Username: username, PasswordHash: hash, RoleID: role.ID, IsActive: true}
		if err := db.Create(&u).Error; err != nil {
			log.Fatalf("create err: %v", err)
		}
		fmt.Printf("[CREATED USER] %s role=%s id=%d\n", username, roleName, u.ID)
	} else {
		u.PasswordHash = hash
		u.RoleID = role.ID
		u.IsActive = true
		db.Save(&u)
	}
	return u.ID
}

func seedPermissions(db *gorm.DB) {
	for _, m := range modules {
		for _, a := range actions {
			var p model.Permission
			err := db.Where("module = ? AND action = ?", m, a).First(&p).Error
			if err == gorm.ErrRecordNotFound {
				p = model.Permission{Module: m, Action: a}
				db.Create(&p)
			}
		}
	}
	fmt.Println("[OK] permissions seeded")
}

func grantAll(db *gorm.DB, roleName string) {
	var role model.Role
	db.Where(roleNameQuery, roleName).First(&role)
	var perms []model.Permission
	db.Find(&perms)
	for _, p := range perms {
		var rp model.RolePermission
		if err := db.Where("role_id = ? AND permission_id = ?", role.ID, p.ID).First(&rp).Error; err == gorm.ErrRecordNotFound {
			db.Create(&model.RolePermission{RoleID: role.ID, PermissionID: p.ID})
		}
	}
	fmt.Printf("[OK] granted all to %s\n", roleName)
}

func grantView(db *gorm.DB, roleName string) {
	var role model.Role
	db.Where(roleNameQuery, roleName).First(&role)
	var perms []model.Permission
	db.Where("action = ?", constant.ActionView).Find(&perms)
	for _, p := range perms {
		var rp model.RolePermission
		if err := db.Where("role_id = ? AND permission_id = ?", role.ID, p.ID).First(&rp).Error; err == gorm.ErrRecordNotFound {
			db.Create(&model.RolePermission{RoleID: role.ID, PermissionID: p.ID})
		}
	}
	fmt.Printf("[OK] granted view to %s\n", roleName)
}

func findOrCreateKategori(db *gorm.DB, nama string) model.Kategori {
	var kat model.Kategori
	if db.Where("nama = ?", nama).First(&kat).Error == gorm.ErrRecordNotFound {
		kat = model.Kategori{Nama: nama}
		db.Create(&kat)
	}
	return kat
}

func findOrCreateSatuan(db *gorm.DB, nama, singkatan string) model.Satuan {
	var sat model.Satuan
	if db.Where("nama = ?", nama).First(&sat).Error == gorm.ErrRecordNotFound {
		sat = model.Satuan{Nama: nama, Singkatan: singkatan}
		db.Create(&sat)
	}
	return sat
}

func findOrCreateGudang(db *gorm.DB, nama, alamat string) model.Gudang {
	var g model.Gudang
	if db.Where("nama = ?", nama).First(&g).Error == gorm.ErrRecordNotFound {
		g = model.Gudang{Nama: nama, Alamat: alamat}
		db.Create(&g)
	}
	return g
}

func findOrCreateGudangDenganKode(db *gorm.DB, nama, alamat, kode string) model.Gudang {
	var g model.Gudang
	if db.Where("nama = ?", nama).First(&g).Error == gorm.ErrRecordNotFound {
		g = model.Gudang{Nama: nama, Alamat: alamat, Kode: kode}
		db.Create(&g)
		return g
	}
	if g.Kode == "" {
		g.Kode = kode
		db.Save(&g)
	}
	return g
}

func seedBarang(db *gorm.DB, katID, satID uint) {
	barangList := []model.Barang{
		{KodeBarang: "BRG-001", Nama: "Sarung Tangan Steril M", KategoriID: katID, SatuanID: satID, StokMinimum: 50, HargaBeli: 45000, Stok: 245, IsActive: true},
		{KodeBarang: "BRG-002", Nama: "Masker Bedah 3 Ply", KategoriID: katID, SatuanID: satID, StokMinimum: 100, HargaBeli: 28000, Stok: 89, IsActive: true},
		{KodeBarang: "BRG-003", Nama: "Alcohol Swab", KategoriID: katID, SatuanID: satID, StokMinimum: 200, HargaBeli: 12000, Stok: 1240, IsActive: true},
		{KodeBarang: "BRG-004", Nama: "Handsanitizer 500ml", KategoriID: katID, SatuanID: satID, StokMinimum: 100, HargaBeli: 25000, Stok: 320, IsActive: true},
	}
	for i := range barangList {
		var e model.Barang
		if db.Where("kode_barang = ?", barangList[i].KodeBarang).First(&e).Error == gorm.ErrRecordNotFound {
			db.Create(&barangList[i])
		}
	}
}

func seedBarangMasuk(db *gorm.DB, g1, g2 model.Gudang, adminRef uint, now time.Time) {
	for i := 0; i < 12; i++ {
		created := now.AddDate(0, -(i / 3), -(i * 5))
		gid := g1.ID
		if i%2 == 1 {
			gid = g2.ID
		}
		bm := model.BarangMasuk{
			NomorPenerimaan: fmt.Sprintf("IN-%04d", 2340+i),
			GudangID:        gid,
			Status:          constant.StatusBMSelesai,
			Tanggal:         created,
			Catatan:         "Data seed",
			DiterimaOleh:    &adminRef,
			CreatedAt:       created,
			UpdatedAt:       created,
		}
		db.Create(&bm)
	}
}

func seedBarangKeluar(db *gorm.DB, g2 model.Gudang, adminRef uint, now time.Time) {
	for i := 0; i < 9; i++ {
		created := now.AddDate(0, -(i / 3), -(i * 7))
		bk := model.BarangKeluar{
			NomorPengeluaran: fmt.Sprintf("OUT-%04d", 1120+i),
			GudangID:         g2.ID,
			Status:           constant.StatusBKSelesai,
			Tanggal:          created,
			Keperluan:        "Distribusi ke RS",
			Penerima:         "RS Hasan Sadikin",
			DikeluarkanOleh:  &adminRef,
			CreatedAt:        created,
			UpdatedAt:        created,
		}
		db.Create(&bk)
	}
}

func seedStockOpname(db *gorm.DB, g1 model.Gudang, now time.Time) {
	for i := 0; i < 3; i++ {
		created := now.AddDate(0, 0, -i*10)
		so := model.StockOpname{
			NomorOpname: fmt.Sprintf("SO-%03d", 441+i),
			GudangID:    g1.ID,
			Status:      constant.StatusSOSelesai,
			Tanggal:     created,
			CreatedAt:   created,
			UpdatedAt:   created,
		}
		db.Create(&so)
	}
}

func seedSampleData(db *gorm.DB, adminID uint) {
	var c int64
	db.Model(&model.BarangMasuk{}).Count(&c)
	if c > 0 {
		fmt.Println("[SKIP] sample already present")
		return
	}

	kat := findOrCreateKategori(db, "Alkes")
	sat := findOrCreateSatuan(db, "Box", "bx")
	g1 := findOrCreateGudangDenganKode(db, "Gudang 1 Cimahi", "Cimahi -7.02090991881257, 107.64954118009624", "BBU")
	g2 := findOrCreateGudangDenganKode(db, "Gudang 2 Bandung Lt.3", "Bandung Lt.3 -6.922156836137817, 107.61645745311088", "MAHANG")
	seedBarang(db, kat.ID, sat.ID)

	now := time.Now()
	seedBarangMasuk(db, g1, g2, adminID, now)
	seedBarangKeluar(db, g2, adminID, now)
	seedStockOpname(db, g1, now)

	fmt.Println("[OK] sample data seeded")
}

func main() {
	cfg := config.Load()
	db := config.NewDatabase(cfg)
	if err := config.AutoMigrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	if err := config.SeedDefaultRoles(db); err != nil {
		log.Fatalf("seed roles: %v", err)
	}
	seedPermissions(db)
	grantAll(db, constant.RoleSuperAdmin)
	grantAll(db, constant.RoleAdmin)
	grantView(db, constant.RoleKaryawan)
	_ = upsertUser(db, "superadmin", "superadmin@stockrsd.local", "Super Admin", "Password123!", constant.RoleSuperAdmin)
	adminID := upsertUser(db, "admin", "admin@stockrsd.local", "Admin Gudang", "Password123!", constant.RoleAdmin)
	_ = upsertUser(db, "karyawan", "karyawan@stockrsd.local", "Karyawan Gudang", "Password123!", constant.RoleKaryawan)
	seedSampleData(db, adminID)
	fmt.Println("Done.")
}
