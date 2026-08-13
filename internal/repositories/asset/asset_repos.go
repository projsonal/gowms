package asset

import (
	"gorm.io/gorm"

	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/utils"
)

type repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func applyFilter(q *gorm.DB, f Filter) *gorm.DB {
	if f.JenisAset != "" {
		q = q.Where("jenis_aset = ?", f.JenisAset)
	}
	if f.GudangID != 0 {
		q = q.Where("gudang_id = ?", f.GudangID)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	return q
}

func (r *repository) List(p utils.PaginationParams, f Filter) ([]model.Asset, int64, error) {
	var list []model.Asset
	var total int64

	q := applyFilter(r.db.Model(&model.Asset{}), f)
	if p.Search != "" {
		q = q.Where("nama ILIKE ? OR label_rsd ILIKE ? OR kode_ba ILIKE ?",
			"%"+p.Search+"%", "%"+p.Search+"%", "%"+p.Search+"%")
	}
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := p.Apply(q.Session(&gorm.Session{}).Preload("Gudang").Order("created_at desc")).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *repository) FindByID(id uint) (*model.Asset, error) {
	var a model.Asset
	if err := r.db.Preload("Gudang").First(&a, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *repository) Create(a *model.Asset) error {
	return r.db.Create(a).Error
}

func (r *repository) Update(a *model.Asset) error {
	return r.db.Save(a).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&model.Asset{}, id).Error
}

// NextRSDNumber menghitung nomor urut berikutnya untuk label RSD di
// gudang tertentu, dari jumlah aset berkoordinat (bukan transportasi)
// yang sudah ada di gudang itu. Reset otomatis per gudang karena hanya
// menghitung baris milik gudang_id tsb.
func (r *repository) NextRSDNumber(gudangID uint) (int, error) {
	var count int64
	err := r.db.Model(&model.Asset{}).
		Where("gudang_id = ? AND label_rsd <> ''", gudangID).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count) + 1, nil
}

// NextBANumber menghitung nomor urut berikutnya untuk kode BA, global
// lintas gudang (khusus aset transportasi).
func (r *repository) NextBANumber() (int, error) {
	var count int64
	err := r.db.Model(&model.Asset{}).
		Where("kode_ba <> ''").
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count) + 1, nil
}

func (r *repository) CountByJenis(jenisAset string) (int64, error) {
	var count int64
	err := r.db.Model(&model.Asset{}).Where("jenis_aset = ?", jenisAset).Count(&count).Error
	return count, err
}
