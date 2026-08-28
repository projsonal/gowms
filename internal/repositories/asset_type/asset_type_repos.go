package assettype

import (
	"errors"

	"gorm.io/gorm"

	"github.com/projsonal/gowms/internal/model"
)

type repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) List() ([]model.AssetType, error) {
	var list []model.AssetType
	err := r.db.Order("urutan asc, id asc").Find(&list).Error
	return list, err
}

func (r *repository) FindByID(id uint) (*model.AssetType, error) {
	var t model.AssetType
	if err := r.db.First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repository) FindByKode(kode string) (*model.AssetType, error) {
	var t model.AssetType
	if err := r.db.Where("kode = ?", kode).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repository) Create(t *model.AssetType) error {
	return r.db.Create(t).Error
}

func (r *repository) Update(t *model.AssetType) error {
	return r.db.Save(t).Error
}

func (r *repository) Delete(id uint) error {
	t, err := r.FindByID(id)
	if err != nil {
		return err
	}
	if t.IsSystem {
		return errors.New("jenis aset bawaan sistem tidak bisa dihapus")
	}
	count, err := r.CountAssetsUsing(t.Kode)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("jenis aset ini masih dipakai oleh data aset yang ada, tidak bisa dihapus")
	}
	return r.db.Delete(&model.AssetType{}, id).Error
}

func (r *repository) CountAssetsUsing(kode string) (int64, error) {
	var count int64
	err := r.db.Model(&model.Asset{}).Where("jenis_aset = ?", kode).Count(&count).Error
	return count, err
}
