package gudang

import (
	"gorm.io/gorm"

	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

const orderNamaAsc = "nama asc"

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

type unitCountRow struct {
	GudangID  uint
	UnitCount int64
	SkuCount  int64
}

func (r *repository) populateUnitCounts(list []model.Gudang) error {
	if len(list) == 0 {
		return nil
	}
	ids := make([]uint, len(list))
	for i, g := range list {
		ids[i] = g.ID
	}

	var rows []unitCountRow
	err := r.db.Table("barang_stok_gudang").
		Select("gudang_id, COALESCE(SUM(stok), 0) AS unit_count, COUNT(DISTINCT CASE WHEN stok > 0 THEN barang_id END) AS sku_count").
		Where("gudang_id IN ?", ids).
		Group("gudang_id").
		Scan(&rows).Error
	if err != nil {
		return err
	}

	counts := make(map[uint]unitCountRow, len(rows))
	for _, row := range rows {
		counts[row.GudangID] = row
	}
	for i := range list {
		if c, ok := counts[list[i].ID]; ok {
			list[i].UnitTersedia = c.UnitCount
			list[i].SkuTersedia = c.SkuCount
		}
	}
	return nil
}

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
	if err := r.populateUnitCounts(list); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *repository) FindGudangByID(id uint) (*model.Gudang, error) {
	var g model.Gudang
	if err := r.db.First(&g, id).Error; err != nil {
		return nil, err
	}

	list := []model.Gudang{g}
	if err := r.populateUnitCounts(list); err != nil {
		return nil, err
	}
	g = list[0]
	return &g, nil
}

func (r *repository) FindGudangByKode(kode string) (*model.Gudang, error) {
	if kode == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var g model.Gudang
	if err := r.db.Where("kode = ?", kode).First(&g).Error; err != nil {
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
