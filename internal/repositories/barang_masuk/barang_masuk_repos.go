package barang_masuk

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/projsonal/gowms/internal/model"
	barangSerial "github.com/projsonal/gowms/internal/repositories/barang_serial"
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
		// Dokumen barang masuk bisa berisi beberapa item; JOIN + DISTINCT
		// (dengan SELECT eksplisit ke tabel utama supaya tidak "ambiguous
		// column" karena semua tabel yang di-JOIN sama-sama punya kolom id)
		// supaya satu dokumen tidak muncul dobel kalau punya >1 item
		// dengan kategori yang sama.
		q = q.Select("barang_masuk.*").Distinct().
			Joins("JOIN barang_masuk_items ON barang_masuk_items.barang_masuk_id = barang_masuk.id").
			Joins("JOIN barang ON barang.id = barang_masuk_items.barang_id").
			Where("barang.kategori_id = ?", f.KategoriID)
	}
	return q
}

func (r *repository) List(p utils.PaginationParams, f Filter) ([]model.BarangMasuk, int64, error) {
	var list []model.BarangMasuk
	var total int64

	q := applyFilter(r.db.Model(&model.BarangMasuk{}), f)
	if p.Search != "" {
		q = q.Where("nomor_penerimaan ILIKE ?", "%"+p.Search+"%")
	}
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := p.Apply(q.Session(&gorm.Session{}).
		Preload("Gudang").Preload("Items").Preload("Items.Barang").Order("id desc")).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *repository) FindByID(id uint) (*model.BarangMasuk, error) {
	var bm model.BarangMasuk
	err := r.db.Preload("Gudang").
		Preload("Items").Preload("Items.Barang").Preload("Items.Rak").
		First(&bm, id).Error
	if err != nil {
		return nil, err
	}
	return &bm, nil
}

func (r *repository) FindByNomor(nomor string) (*model.BarangMasuk, error) {
	var bm model.BarangMasuk
	if err := r.db.Where("nomor_penerimaan = ?", nomor).First(&bm).Error; err != nil {
		return nil, err
	}
	return &bm, nil
}

func (r *repository) Create(bm *model.BarangMasuk) error {
	return r.db.Create(bm).Error
}

func (r *repository) Update(bm *model.BarangMasuk, items []model.BarangMasukItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("barang_masuk_id = ?", bm.ID).Delete(&model.BarangMasukItem{}).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].ID = 0
			items[i].BarangMasukID = bm.ID
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		return tx.Save(bm).Error
	})
}

func (r *repository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("barang_masuk_id = ?", id).Delete(&model.BarangMasukItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.BarangMasuk{}, id).Error
	})
}

func (r *repository) Complete(id uint, userID uint, serials map[uint][]string) (*model.BarangMasuk, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var bm model.BarangMasuk
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Preload("Items").First(&bm, id).Error; err != nil {
			return err
		}
		if bm.Status != constant.StatusBMDraft {
			return errors.New(constant.ErrBMBukanDraft)
		}

		for _, item := range bm.Items {
			var b model.Barang
			if err := tx.First(&b, item.BarangID).Error; err != nil {
				return err
			}
			// Barang bertanda IsSerialized WAJIB disertai SN sejumlah
			// persis Qty saat diselesaikan — inilah titik masuk data unit
			// fisik (lihat barang_serial.CreateUnitsTx & model.BarangSerial).
			if b.IsSerialized {
				sn := serials[item.ID]
				if len(sn) != item.Qty {
					return fmt.Errorf("%s (barang: %s, qty: %d, sn diisi: %d)",
						constant.ErrSerialJumlahTidakSesuai, b.Nama, item.Qty, len(sn))
				}
				if err := barangSerial.CreateUnitsTx(tx, item.BarangID, bm.GudangID, item.RakID, item.ID, sn); err != nil {
					return err
				}
			}

			if err := tx.Model(&model.Barang{}).Where("id = ?", item.BarangID).
				Update("stok", gorm.Expr("stok + ?", item.Qty)).Error; err != nil {
				return err
			}
			if item.RakID != nil {
				if err := adjustRak(tx, *item.RakID, item.Qty); err != nil {
					return err
				}
			}
		}

		now := time.Now()
		return tx.Model(&model.BarangMasuk{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":        constant.StatusBMSelesai,
			"diterima_oleh": userID,
			"completed_at":  now,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

// adjustRak menambah/mengurangi Terisi pada rak, meng-clamp ke [0, tak
// terbatas] lalu menghitung ulang status kosong/terisi_sebagian/penuh —
// sama persis dengan logika Rak.RecalculateStatus di modul Manajemen
// Gudang (murni dari angka tercatat, tidak perlu sensor IoT).
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

func (r *repository) Batalkan(id uint) (*model.BarangMasuk, error) {
	res := r.db.Model(&model.BarangMasuk{}).Where("id = ? AND status = ?", id, constant.StatusBMDraft).
		Update("status", constant.StatusBMDibatalkan)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, errors.New(constant.ErrBMBukanDraft)
	}
	return r.FindByID(id)
}

func (r *repository) CountByStatus(status string) (int64, error) {
	var count int64
	err := r.db.Model(&model.BarangMasuk{}).Where(constant.QueryStatusEq, status).Count(&count).Error
	return count, err
}
