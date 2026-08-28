package role

import (
	"gorm.io/gorm"

	"github.com/projsonal/gowms/internal/model"
)

func (r *repository) FindAll() ([]model.Role, error) {
	var roles []model.Role
	err := r.db.Order("id asc").Find(&roles).Error
	return roles, err
}

func (r *repository) FindByID(id uint) (*model.Role, error) {
	var role model.Role
	if err := r.db.First(&role, id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *repository) FindByName(name string) (*model.Role, error) {
	var role model.Role
	if err := r.db.Where("name = ?", name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *repository) Create(role *model.Role) error {
	return r.db.Create(role).Error
}

func (r *repository) Update(role *model.Role) error {
	return r.db.Save(role).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&model.Role{}, id).Error
}

func (r *repository) FindOrCreatePermission(module, action string) (*model.Permission, error) {
	var p model.Permission
	err := r.db.
		Where("module = ? AND action = ?", module, action).
		FirstOrCreate(&p, model.Permission{Module: module, Action: action}).Error
	return &p, err
}

func (r *repository) ReplaceRolePermissions(roleID uint, permissionIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error; err != nil {
			return err
		}
		for _, pid := range permissionIDs {
			if err := tx.Create(&model.RolePermission{RoleID: roleID, PermissionID: pid}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *repository) HasPermission(roleID uint, module, action string) (bool, error) {
	var count int64
	err := r.db.
		Table("role_permissions").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ? AND permissions.module = ? AND permissions.action = ?", roleID, module, action).
		Count(&count).Error
	return count > 0, err
}

func (r *repository) GetMatrix(roleID uint) ([]ModulePermission, error) {
	type row struct {
		Module string
		Action string
	}
	var rows []row
	err := r.db.
		Table("role_permissions").
		Select("permissions.module, permissions.action").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ?", roleID).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	itemsByModule := map[string]*ModulePermission{}
	for _, row := range rows {
		item, ok := itemsByModule[row.Module]
		if !ok {
			item = &ModulePermission{Module: row.Module}
			itemsByModule[row.Module] = item
		}
		switch row.Action {
		case "view":
			item.View = true
		case "tambah":
			item.Tambah = true
		case "edit":
			item.Edit = true
		case "approval_reject":
			item.ApprovalReject = true
		case "print":
			item.Print = true
		case "assign_delegasi":
			item.AssignDelegasi = true
		}
	}

	result := make([]ModulePermission, 0, len(itemsByModule))
	for _, v := range itemsByModule {
		result = append(result, *v)
	}
	return result, nil
}
