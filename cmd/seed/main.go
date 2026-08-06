package seed

import (
	"fmt"
	"log"
	"time"

	"github.com/projsonal/gostock/internal/model"
	"github.com/projsonal/gostock/pkg/config"
	"github.com/projsonal/gostock/pkg/constant"
	"github.com/projsonal/gostock/pkg/utils"
	"gorm.io/gorm"
)

var modules = []string{
	constant.ModuleDashboard, "kategori", "satuan", constant.ModuleManajemenGudang, "rak", "barang", constant.ModuleKelolaBarang,
	constant.ModuleSupplier, constant.ModulePurchaseOrder,
	constant.ModuleBarangMasuk, constant.ModuleBarangKeluar,
	constant.ModuleStockOpname, constant.ModulePengiriman,
	constant.ModuleLaporan, constant.ModuleManajemenUser, constant.ModuleSettings, "notifikasi",
}
var actions = []string{constant.ActionView, constant.ActionTambah, constant.ActionEdit, constant.ActionApprovalReject, constant.ActionPrint, constant.ActionAssignDelegasi}

func upsertUser(db *gorm.DB, username, email, fullname, password, roleName string) uint {
	var role model.Role
	if err := db.Where("name = ?", roleName).First(&role).Error; err != nil {
		log.Fatalf("role %s not found: %v", roleName, err)
	}
	hash, err := utils.HashPassword(password)
	if err != nil {
		log.Fatalf("hash err: %v", err)
	}
	var u model.User
	err = db.Where("username = ?", username).First(&u).Error
	if err == gorm.ErrRecordNotFound {
		u = model.User{Username: username, Email: email, FullName: fullname, PasswordHash: hash, RoleID: role.ID, IsActive: true}
		if err := db.Create(&u).Error; err != nil {
			log.Fatalf("create err: %v", err)
		}
		fmt.Printf("[CREATED USER] %s role=%s id=%d\n", username, roleName, u.ID)
	} else {
		u.PasswordHash = hash
		u.Email = email
		u.FullName = fullname
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
	db.Where("name = ?", roleName).First(&role)
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
	db.Where("name = ?", roleName).First(&role)
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

func seedSampleData(db *gorm.DB, adminID uint) {
	var c int64
	db.Model(&model.BarangMasuk{}).Count(&c)
	if c > 0 {
		fmt.Println("[SKIP] sample already present")
		return
	}
	// Kategori
	var kat model.Kategori
	if db.Where("nama = ?", "Alkes").First(&kat).Error == gorm.ErrRecordNotFound {
		kat = model.Kategori{Nama: "Alkes"}
		db.Create(&kat)
	}
	// Satuan
	var sat model.Satuan
	if db.Where("nama = ?", "Box").First(&sat).Error == gorm.ErrRecordNotFound {
		sat = model.Satuan{Nama: "Box", Singkatan: "bx"}
		db.Create(&sat)
	}
	// Gudang
	var g1, g2 model.Gudang
	if db.Where("nama = ?", "Gudang 1 Cimahi").First(&g1).Error == gorm.ErrRecordNotFound {
		g1 = model.Gudang{Nama: "Gudang 1 Cimahi", Alamat: "Cimahi -7.02090991881257, 107.64954118009624"}
		db.Create(&g1)
	}
	if db.Where("nama = ?", "Gudang 2 Bandung Lt.3").First(&g2).Error == gorm.ErrRecordNotFound {
		g2 = model.Gudang{Nama: "Gudang 2 Bandung Lt.3", Alamat: "Bandung Lt.3 -6.922156836137817, 107.61645745311088"}
		db.Create(&g2)
	}
	// Suppliers
	pic1, pic2, pic3 := "Bpk. Hadi", "Ibu Rina", "Bpk. Andi"
	suppliers := []model.Supplier{
		{Kode: "SUP-01", Nama: "PT Sumber Rejeki", PIC: &pic1, Email: "sales@sumberrejeki.co.id", Telepon: "021-5567890", Alamat: "Cikarang", IsActive: true},
		{Kode: "SUP-02", Nama: "CV Mitra Sejahtera", PIC: &pic2, Email: "info@mitra.id", Telepon: "022-2211445", Alamat: "Bandung", IsActive: true},
		{Kode: "SUP-03", Nama: "PT Logam Prima", PIC: &pic3, Email: "purchase@lp.com", Telepon: "021-8899221", Alamat: "Bekasi", IsActive: true},
	}
	for i := range suppliers {
		var e model.Supplier
		if db.Where("kode = ?", suppliers[i].Kode).First(&e).Error == gorm.ErrRecordNotFound {
			db.Create(&suppliers[i])
		} else {
			suppliers[i] = e
		}
	}
	// Barang
	barangList := []model.Barang{
		{KodeBarang: "BRG-001", Nama: "Sarung Tangan Steril M", KategoriID: kat.ID, SatuanID: sat.ID, StokMinimum: 50, HargaBeli: 45000, Stok: 245, IsActive: true},
		{KodeBarang: "BRG-002", Nama: "Masker Bedah 3 Ply", KategoriID: kat.ID, SatuanID: sat.ID, StokMinimum: 100, HargaBeli: 28000, Stok: 89, IsActive: true},
		{KodeBarang: "BRG-003", Nama: "Alcohol Swab", KategoriID: kat.ID, SatuanID: sat.ID, StokMinimum: 200, HargaBeli: 12000, Stok: 1240, IsActive: true},
		{KodeBarang: "BRG-004", Nama: "Handsanitizer 500ml", KategoriID: kat.ID, SatuanID: sat.ID, StokMinimum: 100, HargaBeli: 25000, Stok: 320, IsActive: true},
	}
	for i := range barangList {
		var e model.Barang
		if db.Where("kode_barang = ?", barangList[i].KodeBarang).First(&e).Error == gorm.ErrRecordNotFound {
			db.Create(&barangList[i])
		}
	}
	now := time.Now()
	// Purchase Orders spread across last 6 months
	statuses := []string{constant.StatusPODisetujui, constant.StatusPODiajukan, constant.StatusPODraft, constant.StatusPOSelesai}
	adminRef := adminID
	for i := 0; i < 8; i++ {
		created := now.AddDate(0, -(i / 2), -(i * 3))
		po := model.PurchaseOrder{
			NomorPO: fmt.Sprintf("PO-%04d", 890+i), SupplierID: suppliers[i%len(suppliers)].ID,
			Status: statuses[i%len(statuses)], TanggalPO: created, TotalEstimasi: int64((i + 1) * 2500000),
			DiajukanOleh: &adminRef, CreatedAt: created, UpdatedAt: created,
		}
		db.Create(&po)
	}
	// Barang Masuk
	for i := 0; i < 12; i++ {
		created := now.AddDate(0, -(i / 3), -(i * 5))
		gid := g1.ID
		if i%2 == 1 {
			gid = g2.ID
		}
		bm := model.BarangMasuk{
			NomorPenerimaan: fmt.Sprintf("IN-%04d", 2340+i), GudangID: gid,
			SupplierID: &suppliers[i%len(suppliers)].ID, Status: constant.StatusBMSelesai,
			Tanggal: created, Catatan: "Data seed", DiterimaOleh: &adminRef,
			CreatedAt: created, UpdatedAt: created,
		}
		db.Create(&bm)
	}
	// Barang Keluar
	for i := 0; i < 9; i++ {
		created := now.AddDate(0, -(i / 3), -(i * 7))
		bk := model.BarangKeluar{
			NomorPengeluaran: fmt.Sprintf("OUT-%04d", 1120+i), GudangID: g2.ID,
			Status: constant.StatusBKSelesai, Tanggal: created, Keperluan: "Distribusi ke RS",
			Penerima: "RS Hasan Sadikin", DikeluarkanOleh: &adminRef,
			CreatedAt: created, UpdatedAt: created,
		}
		db.Create(&bk)
	}
	// Pengiriman
	statuspg := []string{constant.StatusPGDalamPerjalanan, constant.StatusPGTerkirim, constant.StatusPGDijadwalkan, constant.StatusPGDraft}
	tujuan := []struct {
		nama string
		lat  float64
		lng  float64
	}{
		{"RS Hasan Sadikin, Bandung", -6.897472, 107.601444},
		{"Cikarang Selatan, Bekasi", -6.319167, 107.153889},
		{"Kota Bogor", -6.595038, 106.816635},
		{"Sumedang Kota", -6.859628, 107.921478},
	}
	for i := 0; i < 4; i++ {
		created := now.Add(-time.Duration(i) * time.Hour)
		originID := g1.ID
		if i%2 == 1 {
			originID = g2.ID
		}
		pg := model.Pengiriman{
			NomorPengiriman: fmt.Sprintf("JX-%d", 88213+i), GudangAsalID: originID,
			JenisPengambilan: "dropoff", NamaPenerima: tujuan[i].nama, AlamatTujuan: tujuan[i].nama,
			NamaKurir: []string{"Rudi Setiawan", "Ahmad Fauzi", "Wawan H.", "Sinta A."}[i], TeleponKurir: "08123456789",
			Status: statuspg[i], TanggalKirim: created,
			CreatedAt: created, UpdatedAt: created,
		}
		db.Create(&pg)
	}
	// Stock Opname
	for i := 0; i < 3; i++ {
		created := now.AddDate(0, 0, -i*10)
		so := model.StockOpname{
			NomorOpname: fmt.Sprintf("SO-%03d", 441+i), GudangID: g1.ID,
			Status: constant.StatusSOSelesai, Tanggal: created,
			CreatedAt: created, UpdatedAt: created,
		}
		db.Create(&so)
	}
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
