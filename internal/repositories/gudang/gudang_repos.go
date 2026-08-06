package gudang

import (
	"gorm.io/gorm"

	"github.com/projsonal/gostock/internal/model"
	"github.com/projsonal/gostock/pkg/constant"
	"github.com/projsonal/gostock/pkg/utils"
)

const orderNamaAsc = "nama asc"

// ---- Kategori ----


func (r *repository) ListKategori(p utils.PaginationParams) ([]model.Kategori, int64, error) {
	var list []model.Kategori
	var total int64

	q := r.db.Model(&model.Kategori{})
	if p.Search != "" {
		q = q.Where(constant.QueryNamaILike, "%"+p.Search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := p.Apply(q.Order(orderNamaAsc)).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *repository) FindKategoriByID(id uint) (*model.Kategori, error) {
	var k model.Kategori
	if err := r.db.First(&k, id).Error; err != nil {
		return nil, err
	}
	return &k, nil
}

func (r *repository) FindKategoriByNama(nama string) (*model.Kategori, error) {
	var k model.Kategori
	if err := r.db.Where("nama = ?", nama).First(&k).Error; err != nil {
		return nil, err
	}
	return &k, nil
}

func (r *repository) CreateKategori(k *model.Kategori) error {
	return r.db.Create(k).Error
}

func (r *repository) UpdateKategori(k *model.Kategori) error {
	return r.db.Save(k).Error
}

func (r *repository) DeleteKategori(id uint) error {
	return r.db.Delete(&model.Kategori{}, id).Error
}

func (r *repository) CountKategori() (int64, error) {
	var count int64
	err := r.db.Model(&model.Kategori{}).Count(&count).Error
	return count, err
}

// ---- Satuan ----

func (r *repository) ListSatuan(p utils.PaginationParams) ([]model.Satuan, int64, error) {
	var list []model.Satuan
	var total int64

	q := r.db.Model(&model.Satuan{})
	if p.Search != "" {
		q = q.Where(constant.QueryNamaILike, "%"+p.Search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := p.Apply(q.Order(orderNamaAsc)).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *repository) FindSatuanByID(id uint) (*model.Satuan, error) {
	var s model.Satuan
	if err := r.db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repository) FindSatuanByNama(nama string) (*model.Satuan, error) {
	var s model.Satuan
	if err := r.db.Where("nama = ?", nama).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repository) CreateSatuan(s *model.Satuan) error {
	return r.db.Create(s).Error
}

func (r *repository) UpdateSatuan(s *model.Satuan) error {
	return r.db.Save(s).Error
}

func (r *repository) DeleteSatuan(id uint) error {
	return r.db.Delete(&model.Satuan{}, id).Error
}

// ---- Gudang ----

func (r *repository) ListGudang(p utils.PaginationParams) ([]model.Gudang, int64, error) {
	var list []model.Gudang
	var total int64

	q := r.db.Model(&model.Gudang{})
	if p.Search != "" {
		q = q.Where(constant.QueryNamaILike, "%"+p.Search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := p.Apply(q.Order(orderNamaAsc)).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *repository) FindGudangByID(id uint) (*model.Gudang, error) {
	var g model.Gudang
	if err := r.db.Preload("Raks").First(&g, id).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *repository) CreateGudang(g *model.Gudang) error {
	return r.db.Create(g).Error
}

func (r *repository) UpdateGudang(g *model.Gudang) error {
	return r.db.Save(g).Error
}

func (r *repository) DeleteGudang(id uint) error {
	return r.db.Delete(&model.Gudang{}, id).Error
}

func (r *repository) CountGudang() (int64, error) {
	var count int64
	err := r.db.Model(&model.Gudang{}).Count(&count).Error
	return count, err
}

// ---- Rak ----

func (r *repository) ListRak(p utils.PaginationParams, gudangID uint) ([]model.Rak, int64, error) {
	var list []model.Rak
	var total int64

	base := r.db.Model(&model.Rak{})
	if gudangID != 0 {
		base = base.Where(constant.QueryGudangIDEq, gudangID)
	}
	if p.Search != "" {
		base = base.Where(constant.QueryKodeRakILIKE, "%"+p.Search+"%")
	}

	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := p.Apply(base.Session(&gorm.Session{}).Preload("Gudang").Order("kode_rak asc")).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *repository) FindRakByID(id uint) (*model.Rak, error) {
	var rak model.Rak
	if err := r.db.Preload("Gudang").First(&rak, id).Error; err != nil {
		return nil, err
	}
	return &rak, nil
}

func (r *repository) FindRakByKode(kode string) (*model.Rak, error) {
	var rak model.Rak
	if err := r.db.Where("kode_rak = ?", kode).First(&rak).Error; err != nil {
		return nil, err
	}
	return &rak, nil
}

func (r *repository) CreateRak(rak *model.Rak) error {
	return r.db.Create(rak).Error
}

func (r *repository) UpdateRak(rak *model.Rak) error {
	return r.db.Save(rak).Error
}

func (r *repository) DeleteRak(id uint) error {
	return r.db.Delete(&model.Rak{}, id).Error
}

func (r *repository) CountRakAll() (int64, error) {
	var count int64
	err := r.db.Model(&model.Rak{}).Count(&count).Error
	return count, err
}

func (r *repository) CountRakByStatus(status string) (int64, error) {
	var count int64
	err := r.db.Model(&model.Rak{}).Where(constant.QueryStatusEq, status).Count(&count).Error
	return count, err
}

func (r *repository) CountRakByGudang(gudangID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Rak{}).Where(constant.QueryGudangIDEq, gudangID).Count(&count).Error
	return count, err
}

func (r *repository) AdjustRakTerisi(rakID uint, delta int) (*model.Rak, error) {
	var rak model.Rak
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&rak, rakID).Error; err != nil {
			return err
		}
		rak.Terisi += delta
		if rak.Terisi < 0 {
			rak.Terisi = 0
		}
		rak.RecalculateStatus()
		return tx.Save(&rak).Error
	})
	if err != nil {
		return nil, err
	}
	return &rak, nil
}
