package barang_keluar

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

func applyFilter(q *gorm.DB, f Filter) *gorm.DB {
	if f.Status != "" {
		q = q.Where(constant.QueryStatusEq, f.Status)
	}
	if f.GudangID != 0 {
		q = q.Where(constant.QueryGudangIDEq, f.GudangID)
	}
	if f.KategoriID != 0 {
		q = q.Select("barang_keluar.*").Distinct().
			Joins("JOIN barang_keluar_items ON barang_keluar_items.barang_keluar_id = barang_keluar.id").
			Joins("JOIN barang ON barang.id = barang_keluar_items.barang_id").
			Where("barang.kategori_id = ?", f.KategoriID)
	}
	return q
}

func (r *repository) List(p utils.PaginationParams, f Filter) ([]model.BarangKeluar, int64, error) {
	var list []model.BarangKeluar
	var total int64

	q := applyFilter(r.db.Model(&model.BarangKeluar{}), f)
	if p.Search != "" {
		q = q.Where("nomor_pengeluaran ILIKE ?", "%"+p.Search+"%")
	}
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := p.Apply(q.Session(&gorm.Session{}).Preload("Gudang").Order("id desc")).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *repository) FindByID(id uint) (*model.BarangKeluar, error) {
	var bk model.BarangKeluar
	err := r.db.Preload("Gudang").Preload("Items").Preload("Items.Barang").Preload("Items.Rak").First(&bk, id).Error
	if err != nil {
		return nil, err
	}
	return &bk, nil
}

func (r *repository) FindByNomor(nomor string) (*model.BarangKeluar, error) {
	var bk model.BarangKeluar
	if err := r.db.Where("nomor_pengeluaran = ?", nomor).First(&bk).Error; err != nil {
		return nil, err
	}
	return &bk, nil
}

func (r *repository) Create(bk *model.BarangKeluar) error {
	return r.db.Create(bk).Error
}

func (r *repository) Update(bk *model.BarangKeluar, items []model.BarangKeluarItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("barang_keluar_id = ?", bk.ID).Delete(&model.BarangKeluarItem{}).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].ID = 0
			items[i].BarangKeluarID = bk.ID
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		return tx.Save(bk).Error
	})
}

func (r *repository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("barang_keluar_id = ?", id).Delete(&model.BarangKeluarItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.BarangKeluar{}, id).Error
	})
}

func (r *repository) Complete(id uint, userID uint) (*model.BarangKeluar, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var bk model.BarangKeluar
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Preload("Items").First(&bk, id).Error; err != nil {
			return err
		}
		if bk.Status != constant.StatusBKDraft {
			return errors.New(constant.ErrBKBukanDraft)
		}

		// Validasi ketersediaan SEMUA item terlebih dahulu, sebelum
		// mengubah apa pun — supaya dokumen dengan satu baris saja yang
		// kurang stok tidak menyebabkan sebagian barang lain terlanjur
		// terpotong (all-or-nothing).
		for _, item := range bk.Items {
			var b model.Barang
			if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&b, item.BarangID).Error; err != nil {
				return err
			}
			if b.Stok < item.Qty {
				return fmt.Errorf("%s (barang: %s, tersedia: %d, diminta: %d)",
					constant.ErrBKStokTidakCukup, b.Nama, b.Stok, item.Qty)
			}
			if item.RakID != nil {
				var rak model.Rak
				if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&rak, *item.RakID).Error; err != nil {
					return err
				}
				if rak.Terisi < item.Qty {
					return fmt.Errorf("%s (rak: %s, terisi: %d, diminta: %d)",
						constant.ErrBKRakTidakCukup, rak.KodeRak, rak.Terisi, item.Qty)
				}
			}
		}

		for _, item := range bk.Items {
			if err := tx.Model(&model.Barang{}).Where("id = ?", item.BarangID).
				Update("stok", gorm.Expr("stok - ?", item.Qty)).Error; err != nil {
				return err
			}
			if item.RakID != nil {
				if err := adjustRak(tx, *item.RakID, -item.Qty); err != nil {
					return err
				}
			}
		}

		now := time.Now()
		return tx.Model(&model.BarangKeluar{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":           constant.StatusBKSelesai,
			"dikeluarkan_oleh": userID,
			"completed_at":     now,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

// adjustRak lihat dokumentasi di package barang_masuk — logika identik,
// diduplikasi tipis di sini karena kedua package memang sengaja dibuat
// independen (tidak saling meng-import) supaya masing-masing modul tetap
// bisa berdiri sendiri.
func adjustRak(tx *gorm.DB, rakID uint, delta int) error {
	var rak model.Rak
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&rak, rakID).Error; err != nil {
		return err
	}
	rak.Terisi += delta
	if rak.Terisi < 0 {
		rak.Terisi = 0
	}
	rak.RecalculateStatus()
	return tx.Save(&rak).Error
}

func (r *repository) Batalkan(id uint) (*model.BarangKeluar, error) {
	res := r.db.Model(&model.BarangKeluar{}).Where("id = ? AND status = ?", id, constant.StatusBKDraft).
		Update("status", constant.StatusBKDibatalkan)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, errors.New(constant.ErrBKBukanDraft)
	}
	return r.FindByID(id)
}

func (r *repository) CountByStatus(status string) (int64, error) {
	var count int64
	err := r.db.Model(&model.BarangKeluar{}).Where(constant.QueryStatusEq, status).Count(&count).Error
	return count, err
}
