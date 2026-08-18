package supplier

import (
	"strings"

	"gorm.io/gorm"

	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/utils"
)

func applyFilter(q *gorm.DB, f Filter) *gorm.DB {
	if f.OnlyActive {
		q = q.Where("is_active = ?", true)
	}
	return q
}

func (r *repository) List(p utils.PaginationParams, f Filter) ([]model.Supplier, int64, error) {
	var list []model.Supplier
	var total int64

	q := applyFilter(r.db.Model(&model.Supplier{}), f)
	if p.Search != "" {
		q = q.Where("nama ILIKE ? OR kode ILIKE ?", "%"+p.Search+"%", "%"+p.Search+"%")
	}
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := p.Apply(q.Session(&gorm.Session{}).Order("nama asc")).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *repository) FindByID(id uint) (*model.Supplier, error) {
	var s model.Supplier
	if err := r.db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repository) FindByKode(kode string) (*model.Supplier, error) {
	var s model.Supplier
	if err := r.db.Where("kode = ?", kode).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repository) Create(s *model.Supplier) error {
	return r.db.Create(s).Error
}

func (r *repository) Update(s *model.Supplier) error {
	return r.db.Save(s).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&model.Supplier{}, id).Error
}

// InUse — lihat dokumentasi di interfaces.go. Cukup satu baris ditemukan
// di salah satu tabel untuk dianggap "masih dipakai".
func (r *repository) InUse(id uint) (bool, error) {
	var poCount int64
	if err := r.db.Model(&model.PurchaseOrder{}).Where("supplier_id = ?", id).Count(&poCount).Error; err != nil {
		return false, err
	}
	if poCount > 0 {
		return true, nil
	}
	var bmCount int64
	if err := r.db.Model(&model.BarangMasuk{}).Where("supplier_id = ?", id).Count(&bmCount).Error; err != nil {
		return false, err
	}
	return bmCount > 0, nil
}

func (r *repository) CountAll() (int64, error) {
	var count int64
	err := r.db.Model(&model.Supplier{}).Count(&count).Error
	return count, err
}

func (r *repository) CountActive() (int64, error) {
	var count int64
	err := r.db.Model(&model.Supplier{}).Where("is_active = ?", true).Count(&count).Error
	return count, err
}

// KurirStats lihat dokumentasi di interfaces.go. Query langsung ke tabel
// pengiriman (model.Pengiriman) berdasarkan kecocokan NamaKurir — TANPA
// join/foreign key formal, karena NamaKurir memang disimpan sebagai teks
// bebas di Pengiriman (lihat catatan panjang soal ini di modules.ts
// frontend). Kalau kurirNames kosong (supplier belum mengisi Kerjasama
// Kurir), langsung kembalikan 0/0 tanpa query.
func (r *repository) KurirStats(kurirNames []string) (int64, int64, error) {
	if len(kurirNames) == 0 {
		return 0, 0, nil
	}
	// LOWER(...) di kedua sisi — nama_kurir disimpan sebagai teks bebas, jadi
	// variasi kapitalisasi ("GoSend" vs "gosend") sebelumnya bikin baris
	// yang sebenarnya cocok tidak ikut terhitung (Total Order/Rating jadi
	// 0 padahal datanya ada).
	lowerNames := make([]string, len(kurirNames))
	for i, n := range kurirNames {
		lowerNames[i] = strings.ToLower(strings.TrimSpace(n))
	}
	var totalOrder int64
	if err := r.db.Model(&model.Pengiriman{}).
		Where("LOWER(nama_kurir) IN ? AND status NOT IN ?", lowerNames, []string{"draft", "dibatalkan"}).
		Count(&totalOrder).Error; err != nil {
		return 0, 0, err
	}
	var terkirim int64
	if err := r.db.Model(&model.Pengiriman{}).
		Where("LOWER(nama_kurir) IN ? AND status = ?", lowerNames, "terkirim").
		Count(&terkirim).Error; err != nil {
		return 0, 0, err
	}
	return totalOrder, terkirim, nil
}
