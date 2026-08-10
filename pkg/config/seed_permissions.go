package config

import (
	"gorm.io/gorm"

	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/constant"
)

// operationalModules — modul kerja sehari-hari (BUKAN manajemen_user/
// settings, yang sengaja tidak diberi izin default ke admin/karyawan demi
// prinsip least-privilege — super_admin tetap bisa akses semua modul
// tanpa bergantung matrix ini sama sekali, lihat RequirePermission()).
var operationalModules = []string{
	constant.ModuleKelolaBarang,
	constant.ModuleSupplier,
	constant.ModuleManajemenGudang,
	constant.ModuleBarangMasuk,
	constant.ModuleBarangKeluar,
	constant.ModulePurchaseOrder,
	constant.ModuleStockOpname,
	constant.ModulePengiriman,
	constant.ModuleCOD,
	constant.ModuleLaporan,
	constant.ModuleDashboard,
}

// grant memasangkan satu module dengan beberapa action sekaligus — helper
// kecil supaya daftar izin default di bawah gampang dibaca.
type grant struct {
	modules []string
	actions []string
}

func expandGrants(grants []grant) []struct{ module, action string } {
	out := make([]struct{ module, action string }, 0)
	for _, g := range grants {
		for _, mod := range g.modules {
			for _, action := range g.actions {
				out = append(out, struct{ module, action string }{mod, action})
			}
		}
	}
	return out
}

// SeedDefaultPermissions memberi izin awal yang masuk akal untuk role
// "admin" dan "karyawan" SAAT PERTAMA KALI dibuat (role_permissions-nya
// masih kosong sama sekali) — supaya akun baru langsung bisa dipakai,
// bukan terkunci total sampai super_admin sempat membuka Perizinan Hak
// Akses dan menyalakan puluhan toggle satu per satu. Kalau sebuah role
// SUDAH punya minimal satu baris role_permissions (mis. sudah pernah
// diatur manual lewat UI Perizinan Hak Akses), fungsi ini TIDAK
// menimpanya — jadi aman dipanggil berulang setiap kali server start.
//
// super_admin SENGAJA tidak diseed di sini sama sekali: role itu selalu
// lolos RequirePermission() tanpa bergantung baris di tabel ini (lihat
// internal/middleware/rbac_middleware.go), jadi tidak butuh baris apa pun.
func SeedDefaultPermissions(db *gorm.DB) error {
	var adminRole, karyawanRole model.Role
	if err := db.Where("name = ?", constant.RoleAdmin).First(&adminRole).Error; err != nil {
		return err
	}
	if err := db.Where("name = ?", constant.RoleKaryawan).First(&karyawanRole).Error; err != nil {
		return err
	}

	adminGrants := expandGrants([]grant{
		{modules: operationalModules, actions: []string{constant.ActionView, constant.ActionTambah, constant.ActionEdit, constant.ActionPrint}},
	})
	if err := seedRoleIfEmpty(db, adminRole.ID, adminGrants); err != nil {
		return err
	}

	karyawanGrants := expandGrants([]grant{
		{modules: operationalModules, actions: []string{constant.ActionView, constant.ActionTambah}},
		// Karyawan (kurir) WAJIB izin "edit" khusus di modul Pengiriman —
		// endpoint kirim ping GPS (POST /pengiriman/:id/lokasi) dijaga
		// permission "edit" (lihat RegisterRoutes di
		// pengiriman_controller.go), bukan "tambah". Tanpa ini, fitur
		// "Bagikan Lokasi" kurir akan selalu gagal dengan "role anda
		// tidak memiliki izin untuk aksi ini" walau role karyawan-nya
		// sendiri sudah benar. Ditambahkan ke daftar grant YANG SAMA
		// (bukan panggilan seedRoleIfEmpty terpisah) supaya guard
		// "sudah pernah diseed" di bawah tidak membuat baris ini
		// terlewat begitu saja pada panggilan kedua.
		{modules: []string{constant.ModulePengiriman}, actions: []string{constant.ActionEdit}},
	})
	if err := seedRoleIfEmpty(db, karyawanRole.ID, karyawanGrants); err != nil {
		return err
	}
	return nil
}

func seedRoleIfEmpty(db *gorm.DB, roleID uint, grants []struct{ module, action string }) error {
	var count int64
	if err := db.Model(&model.RolePermission{}).Where("role_id = ?", roleID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil // sudah pernah dikonfigurasi (manual/seeding sebelumnya) -- jangan ditimpa
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, g := range grants {
			var perm model.Permission
			if err := tx.Where("module = ? AND action = ?", g.module, g.action).
				FirstOrCreate(&perm, model.Permission{Module: g.module, Action: g.action}).Error; err != nil {
				return err
			}
			rp := model.RolePermission{RoleID: roleID, PermissionID: perm.ID}
			if err := tx.Where("role_id = ? AND permission_id = ?", roleID, perm.ID).
				FirstOrCreate(&rp).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
