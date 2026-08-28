package barang

import (
	"strconv"
	"strings"

	"gorm.io/gorm"

	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/internal/repositories/barangstokgudang"
	"github.com/projsonal/gowms/pkg/utils"
)

func applyFilter(q *gorm.DB, f Filter) *gorm.DB {
	if f.KategoriID != 0 {
		q = q.Where("kategori_id = ?", f.KategoriID)
	}
	if f.SatuanID != 0 {
		q = q.Where("satuan_id = ?", f.SatuanID)
	}
	if f.StokMenipis {
		q = q.Where("stok_minimum > 0 AND stok <= stok_minimum")
	}
	if f.OnlyActive {
		q = q.Where("is_active = ?", true)
	}
	if len(f.ApprovalStatuses) > 0 {
		if f.OrSubmittedBy != 0 {
			q = q.Where("approval_status IN ? OR diajukan_oleh = ?", f.ApprovalStatuses, f.OrSubmittedBy)
		} else {
			q = q.Where("approval_status IN ?", f.ApprovalStatuses)
		}
	}
	return q
}

func (r *repository) List(p utils.PaginationParams, f Filter) ([]model.Barang, int64, error) {
	var list []model.Barang
	var total int64

	q := applyFilter(r.db.Model(&model.Barang{}), f)
	if p.Search != "" {
		q = q.Where("nama ILIKE ? OR kode_barang ILIKE ?", "%"+p.Search+"%", "%"+p.Search+"%")
	}
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := p.Apply(q.Session(&gorm.Session{}).Preload("Kategori").Preload("Satuan").Preload("Didelegasikan").Order("nama asc")).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *repository) FindByID(id uint) (*model.Barang, error) {
	var b model.Barang
	if err := r.db.Preload("Kategori").Preload("Satuan").Preload("Didelegasikan").First(&b, id).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *repository) FindByKode(kode string) (*model.Barang, error) {
	var b model.Barang
	if err := r.db.Where("kode_barang = ?", kode).First(&b).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *repository) Create(b *model.Barang) error {
	return r.db.Create(b).Error
}

func (r *repository) Update(b *model.Barang) error {
	return r.db.Save(b).Error
}

func (r *repository) SetStokGudangAwal(barangID, gudangID uint, stok int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return barangstokgudang.SetStokGudangTx(tx, barangID, gudangID, stok)
	})
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&model.Barang{}, id).Error
}

func (r *repository) AdjustStok(id uint, delta int) (*model.Barang, error) {
	var b model.Barang
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&b, id).Error; err != nil {
			return err
		}
		b.Stok += delta
		if b.Stok < 0 {
			b.Stok = 0
		}
		return tx.Save(&b).Error
	})
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *repository) NextSKUNumber(prefix string) (int, error) {
	var codes []string
	err := r.db.Unscoped().Model(&model.Barang{}).
		Where("kode_barang LIKE ?", prefix+"-%").
		Pluck("kode_barang", &codes).Error
	if err != nil {
		return 0, err
	}
	return maxSuffixNumber(codes, prefix+"-") + 1, nil
}

func maxSuffixNumber(values []string, sep string) int {
	max := 0
	for _, v := range values {
		idx := strings.LastIndex(v, sep)
		if idx == -1 {
			continue
		}
		suffix := v[idx+len(sep):]
		n, err := strconv.Atoi(suffix)
		if err != nil {
			continue
		}
		if n > max {
			max = n
		}
	}
	return max
}

func (r *repository) CountAll() (int64, error) {
	var count int64
	err := r.db.Model(&model.Barang{}).Count(&count).Error
	return count, err
}

func (r *repository) CountStokMenipis() (int64, error) {
	var count int64
	err := r.db.Model(&model.Barang{}).
		Where("stok_minimum > 0 AND stok <= stok_minimum").
		Count(&count).Error
	return count, err
}

func (r *repository) SumNilaiInventaris() (int64, error) {
	var total int64
	err := r.db.Model(&model.Barang{}).
		Select("COALESCE(SUM(stok * harga_beli), 0)").
		Scan(&total).Error
	return total, err
}
