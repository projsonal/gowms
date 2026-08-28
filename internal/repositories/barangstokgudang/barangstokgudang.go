package barangstokgudang

import (
	"gorm.io/gorm"

	"github.com/projsonal/gowms/internal/model"
)

func AdjustStokGudangTx(tx *gorm.DB, barangID, gudangID uint, delta int) error {
	var row model.BarangStokGudang
	err := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("barang_id = ? AND gudang_id = ?", barangID, gudangID).
		First(&row).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}

		stok := delta
		if stok < 0 {
			stok = 0
		}
		return tx.Create(&model.BarangStokGudang{BarangID: barangID, GudangID: gudangID, Stok: stok}).Error
	}
	newStok := row.Stok + delta
	if newStok < 0 {
		newStok = 0
	}
	return tx.Model(&model.BarangStokGudang{}).Where("id = ?", row.ID).Update("stok", newStok).Error
}

func GetStokGudangTx(tx *gorm.DB, barangID, gudangID uint) (int, error) {
	var row model.BarangStokGudang
	err := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("barang_id = ? AND gudang_id = ?", barangID, gudangID).
		First(&row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return row.Stok, nil
}

func SetStokGudangTx(tx *gorm.DB, barangID, gudangID uint, stok int) error {
	if stok < 0 {
		stok = 0
	}
	res := tx.Model(&model.BarangStokGudang{}).
		Where("barang_id = ? AND gudang_id = ?", barangID, gudangID).
		Update("stok", stok)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return tx.Create(&model.BarangStokGudang{BarangID: barangID, GudangID: gudangID, Stok: stok}).Error
	}
	return nil
}

func SyncBarangStokTotalTx(tx *gorm.DB, barangID uint) error {
	var total int64
	if err := tx.Model(&model.BarangStokGudang{}).
		Where("barang_id = ?", barangID).
		Select("COALESCE(SUM(stok), 0)").Scan(&total).Error; err != nil {
		return err
	}
	return tx.Model(&model.Barang{}).Where("id = ?", barangID).Update("stok", total).Error
}

type StokRow struct {
	BarangID   uint
	KodeBarang string
	NamaBarang string
	GudangID   uint
	NamaGudang string
	Stok       int
}

func ListAll(db *gorm.DB) ([]StokRow, error) {
	var rows []StokRow
	err := db.Table("barang_stok_gudang bsg").
		Select("bsg.barang_id, b.kode_barang, b.nama AS nama_barang, bsg.gudang_id, g.nama AS nama_gudang, bsg.stok").
		Joins("JOIN barang b ON b.id = bsg.barang_id").
		Joins("JOIN gudangs g ON g.id = bsg.gudang_id").
		Where("bsg.stok > 0").
		Order("b.nama, g.nama").
		Scan(&rows).Error
	return rows, err
}

func ListByGudang(db *gorm.DB, gudangID uint) ([]StokRow, error) {
	var rows []StokRow
	err := db.Table("barang_stok_gudang bsg").
		Select("bsg.barang_id, b.kode_barang, b.nama AS nama_barang, bsg.gudang_id, bsg.stok").
		Joins("JOIN barang b ON b.id = bsg.barang_id").
		Where("bsg.gudang_id = ? AND bsg.stok > 0", gudangID).
		Order("b.nama").
		Scan(&rows).Error
	return rows, err
}

func ListByBarang(db *gorm.DB, barangID uint) ([]StokRow, error) {
	var rows []StokRow
	err := db.Table("barang_stok_gudang bsg").
		Select("bsg.barang_id, bsg.gudang_id, g.nama AS nama_gudang, bsg.stok").
		Joins("JOIN gudangs g ON g.id = bsg.gudang_id").
		Where("bsg.barang_id = ? AND bsg.stok > 0", barangID).
		Order("g.nama").
		Scan(&rows).Error
	return rows, err
}
