package config

import (
	"gorm.io/gorm"

	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/constant"
)

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
		return nil
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
