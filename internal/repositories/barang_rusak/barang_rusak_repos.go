package barang_rusak

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
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	return q
}

func (r *repository) List(p utils.PaginationParams, f Filter) ([]model.BarangRusak, int64, error) {
	var list []model.BarangRusak
	var total int64

	q := applyFilter(r.db.Model(&model.BarangRusak{}), f)
	if p.Search != "" {
		q = q.Where("label_barang ILIKE ? OR nama_barang ILIKE ?", "%"+p.Search+"%", "%"+p.Search+"%")
	}
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := p.Apply(q.Session(&gorm.Session{}).Preload("Barang").Preload("Pelapor").Preload("Pemeriksa").Order("created_at desc")).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *repository) FindByID(id uint) (*model.BarangRusak, error) {
	var b model.BarangRusak
	if err := r.db.Preload("Barang").Preload("Pelapor").Preload("Pemeriksa").First(&b, id).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *repository) Create(b *model.BarangRusak) error {
	return r.db.Create(b).Error
}

func (r *repository) Update(b *model.BarangRusak) error {
	return r.db.Save(b).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&model.BarangRusak{}, id).Error
}

func (r *repository) CountByStatus(status string) (int64, error) {
	var count int64
	err := r.db.Model(&model.BarangRusak{}).Where("status = ?", status).Count(&count).Error
	return count, err
}
