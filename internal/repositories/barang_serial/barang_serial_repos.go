package barang_serial

import (
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

func applyFilter(q *gorm.DB, f Filter) *gorm.DB {
	if f.BarangID != 0 {
		q = q.Where("barang_id = ?", f.BarangID)
	}
	if f.GudangID != 0 {
		q = q.Where("gudang_id = ?", f.GudangID)
	}
	if f.Status != "" {
		q = q.Where(constant.QueryStatusEq, f.Status)
	}
	return q
}

func (r *repository) List(p utils.PaginationParams, f Filter) ([]model.BarangSerial, int64, error) {
	var list []model.BarangSerial
	var total int64

	q := applyFilter(r.db.Model(&model.BarangSerial{}), f)
	if p.Search != "" {
		q = q.Where("serial_number ILIKE ?", "%"+p.Search+"%")
	}
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := p.Apply(q.Session(&gorm.Session{}).
		Preload("Barang").Preload("Gudang").Preload("Rak").Order("id desc")).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *repository) FindByID(id uint) (*model.BarangSerial, error) {
	var s model.BarangSerial
	if err := r.db.Preload("Barang").Preload("Gudang").Preload("Rak").First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repository) FindBySerial(serial string) (*model.BarangSerial, error) {
	var s model.BarangSerial
	err := r.db.Preload("Barang").Preload("Gudang").Preload("Rak").
		Where("serial_number = ?", serial).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repository) CountByBarang(barangID uint) (int64, int64, int64, error) {
	var tersedia, terpasang, rusak int64
	base := r.db.Model(&model.BarangSerial{}).Where("barang_id = ?", barangID)
	if err := base.Session(&gorm.Session{}).Where("status = ?", constant.StatusSerialTersedia).Count(&tersedia).Error; err != nil {
		return 0, 0, 0, err
	}
	if err := base.Session(&gorm.Session{}).Where("status = ?", constant.StatusSerialTerpasang).Count(&terpasang).Error; err != nil {
		return 0, 0, 0, err
	}
	if err := base.Session(&gorm.Session{}).Where("status = ?", constant.StatusSerialRusak).Count(&rusak).Error; err != nil {
		return 0, 0, 0, err
	}
	return tersedia, terpasang, rusak, nil
}

func (r *repository) UpdateStatusManual(id uint, status string, catatan string) (*model.BarangSerial, error) {
	updates := map[string]interface{}{"status": status, "catatan": catatan}
	// Menandai "rusak" secara manual: unit dianggap tidak lagi ada di
	// lokasi gudang manapun secara siap-pakai (sama seperti barang yang
	// keluar), tapi tetap TIDAK dikaitkan ke dokumen Barang Keluar
	// manapun (BarangKeluarItemID tetap nil) — beda ceritanya dengan
	// "terpasang".
	if status == constant.StatusSerialRusak {
		updates["gudang_id"] = nil
		updates["rak_id"] = nil
	}
	res := r.db.Model(&model.BarangSerial{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, errors.New(constant.ErrSerialTidakDitemukan)
	}
	return r.FindByID(id)
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&model.BarangSerial{}, id).Error
}

// Create mendaftarkan satu unit baru secara MANUAL (bukan lewat Complete()
// dokumen Barang Masuk) — lihat dokumentasi lengkap di interfaces.go.
// Transaksional: validasi barang/gudang, cek keunikan SN, insert baris
// baru berstatus tersedia, DAN naikkan Barang.Stok +1 supaya agregat
// tetap sinkron dengan rincian per-unit — semuanya all-or-nothing.
func (r *repository) Create(barangID, gudangID uint, rakID *uint, serialNumber, catatan string) (*model.BarangSerial, error) {
	var created model.BarangSerial
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var b model.Barang
		if err := tx.First(&b, barangID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New(constant.ErrSerialBarangTidakAda)
			}
			return err
		}
		if !b.IsSerialized {
			return errors.New(constant.ErrSerialBarangBukanSerial)
		}

		var g model.Gudang
		if err := tx.First(&g, gudangID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New(constant.ErrSerialGudangTidakAda)
			}
			return err
		}

		var count int64
		if err := tx.Model(&model.BarangSerial{}).Where("serial_number = ?", serialNumber).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New(constant.ErrSerialSudahDipakai)
		}

		created = model.BarangSerial{
			BarangID:     barangID,
			SerialNumber: serialNumber,
			Status:       constant.StatusSerialTersedia,
			GudangID:     &gudangID,
			RakID:        rakID,
			Catatan:      catatan,
		}
		if err := tx.Create(&created).Error; err != nil {
			return err
		}

		// Naikkan stok agregat +1, konsisten dengan cara Complete() Barang
		// Masuk menambah stok saat unit-unit ber-SN dibuat lewat dokumen —
		// lihat internal/repositories/barang_masuk Complete().
		return tx.Model(&model.Barang{}).Where("id = ?", barangID).
			Update("stok", gorm.Expr("stok + ?", 1)).Error
	})
	if err != nil {
		return nil, err
	}
	return r.FindByID(created.ID)
}

// --- Helper transaksional dipakai LANGSUNG (tanpa lewat interface di
// atas) oleh repositories/barang_masuk & repositories/barang_keluar,
// supaya pembuatan/konsumsi unit SN ikut dalam transaksi DB yang sama
// dengan penyesuaian stok agregat & rak — all-or-nothing bersama proses
// Complete() dokumen induknya, persis pola adjustRak() di kedua paket
// itu. Diekspor (huruf besar) karena dipanggil lintas-paket.

// CreateUnitsTx mendaftarkan unit baru (hasil Barang Masuk) sejumlah
// serials. Menolak kalau ada SN yang sudah pernah terdaftar di sistem
// (unique secara global, lihat model.BarangSerial.SerialNumber).
func CreateUnitsTx(tx *gorm.DB, barangID, gudangID uint, rakID *uint, bmItemID uint, serials []string) error {
	if len(serials) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(serials))
	rows := make([]model.BarangSerial, 0, len(serials))
	for _, sn := range serials {
		if seen[sn] {
			return fmt.Errorf("%s: %s", constant.ErrSerialDuplikatInput, sn)
		}
		seen[sn] = true

		var count int64
		if err := tx.Model(&model.BarangSerial{}).Where("serial_number = ?", sn).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("%s: %s", constant.ErrSerialSudahDipakai, sn)
		}
		rows = append(rows, model.BarangSerial{
			BarangID:          barangID,
			SerialNumber:      sn,
			Status:            constant.StatusSerialTersedia,
			GudangID:          &gudangID,
			RakID:             rakID,
			BarangMasukItemID: &bmItemID,
		})
	}
	return tx.Create(&rows).Error
}

// ConsumeUnitsTx menandai unit-unit dengan SN yang dipilih sebagai
// "terpasang" (keluar dari gudang) saat dokumen Barang Keluar
// diselesaikan. Memvalidasi tiap SN: harus sudah terdaftar untuk
// BarangID yang sama, berstatus "tersedia", dan sedang berada di
// GudangID asal dokumen — supaya tidak bisa "mengeluarkan" unit yang
// sebenarnya berada/tercatat di gudang lain atau sudah lebih dulu
// keluar/rusak.
func ConsumeUnitsTx(tx *gorm.DB, barangID, gudangID uint, bkItemID uint, serials []string) error {
	if len(serials) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(serials))
	for _, sn := range serials {
		if seen[sn] {
			return fmt.Errorf("%s: %s", constant.ErrSerialDuplikatInput, sn)
		}
		seen[sn] = true

		var unit model.BarangSerial
		err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("serial_number = ?", sn).First(&unit).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%s: %s", constant.ErrSerialTidakDitemukan, sn)
			}
			return err
		}
		if unit.BarangID != barangID {
			return fmt.Errorf("%s: %s", constant.ErrSerialBarangTidakSesuai, sn)
		}
		if unit.Status != constant.StatusSerialTersedia || unit.GudangID == nil || *unit.GudangID != gudangID {
			return fmt.Errorf("%s: %s", constant.ErrSerialTidakTersedia, sn)
		}

		if err := tx.Model(&model.BarangSerial{}).Where("id = ?", unit.ID).Updates(map[string]interface{}{
			"status":                constant.StatusSerialTerpasang,
			"gudang_id":             nil,
			"rak_id":                nil,
			"barang_keluar_item_id": bkItemID,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

// CountAvailableTx menghitung berapa unit BERSTATUS TERSEDIA milik
// barangID yang sedang berada di gudangID — dipakai memvalidasi qty
// permintaan Barang Keluar sebelum minta operator memilih SN satu-satu.
func CountAvailableTx(tx *gorm.DB, barangID, gudangID uint) (int64, error) {
	var count int64
	err := tx.Model(&model.BarangSerial{}).
		Where("barang_id = ? AND gudang_id = ? AND status = ?", barangID, gudangID, constant.StatusSerialTersedia).
		Count(&count).Error
	return count, err
}

// RiwayatDokumen mengikuti jejak BarangMasukItemID/BarangKeluarItemID
// milik satu unit untuk mengambil NOMOR dokumen induknya (bukan cuma ID
// baris item) — lewat JOIN manual, BUKAN preload GORM, supaya tidak perlu
// menambah field relasi baru ke model.BarangMasukItem/BarangKeluarItem
// (yang dipakai luas di modul lain; menambah field di sana berisiko
// mengubah bentuk JSON respons endpoint lain yang tidak terkait).
//
// sql.ErrNoRows SENGAJA diabaikan (bukan dikembalikan sebagai error) —
// kalau baris item-nya kebetulan sudah tidak ada (mis. dihapus saat masih
// draft, jarang terjadi untuk unit yang sudah "tersedia"/"terpasang"),
// hasilnya cukup nomor kosong, bukan gagal total menampilkan unit.
func (r *repository) RiwayatDokumen(s *model.BarangSerial) (string, string, error) {
	var nomorMasuk, nomorKeluar string
	if s.BarangMasukItemID != nil {
		row := r.db.Table("barang_masuk_items").
			Select("barang_masuk.nomor_penerimaan").
			Joins("JOIN barang_masuk ON barang_masuk.id = barang_masuk_items.barang_masuk_id").
			Where("barang_masuk_items.id = ?", *s.BarangMasukItemID).Row()
		if err := row.Scan(&nomorMasuk); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", "", err
		}
	}
	if s.BarangKeluarItemID != nil {
		row := r.db.Table("barang_keluar_items").
			Select("barang_keluar.nomor_pengeluaran").
			Joins("JOIN barang_keluar ON barang_keluar.id = barang_keluar_items.barang_keluar_id").
			Where("barang_keluar_items.id = ?", *s.BarangKeluarItemID).Row()
		if err := row.Scan(&nomorKeluar); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", "", err
		}
	}
	return nomorMasuk, nomorKeluar, nil
}
