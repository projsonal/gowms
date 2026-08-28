package dashboard

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/utils"
)

type KategoriComposition struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
}

type BarangRanking struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

type AnalisaResponse struct {
	TotalSKU             int64                 `json:"total_sku"`
	TotalRestockBulanIni int64                 `json:"total_restock_bulan_ini"`
	StokMenipis          int64                 `json:"stok_menipis"`
	KategoriComposition  []KategoriComposition `json:"kategori_composition"`
	TopRestocked         []BarangRanking       `json:"top_restocked"`
	TopKeluar            []BarangRanking       `json:"top_keluar"`

	TotalAset     int64                 `json:"total_aset"`
	AsetRusak     int64                 `json:"aset_rusak"`
	AsetPerJenis  []KategoriComposition `json:"aset_per_jenis"`
	AsetPerStatus []KategoriComposition `json:"aset_per_status"`
	AsetPerGudang []KategoriComposition `json:"aset_per_gudang"`
}

func (h *Controller) Analisa(c *fiber.Ctx) error {
	db := h.db
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var totalSKU int64
	db.Model(&model.Barang{}).Count(&totalSKU)

	var stokMenipis int64
	db.Model(&model.Barang{}).Where("stok_minimum > 0 AND stok <= stok_minimum").Count(&stokMenipis)

	var totalRestock int64
	db.Model(&model.BarangMasuk{}).
		Where("status = 'selesai' AND completed_at >= ?", startOfMonth).
		Count(&totalRestock)

	var kategoriRows []KategoriComposition
	db.Model(&model.Barang{}).
		Select("kategori.nama AS label, COUNT(barang.id) AS value").
		Joins("JOIN kategori ON kategori.id = barang.kategori_id").
		Group("kategori.nama").
		Order("value DESC").
		Scan(&kategoriRows)

	var topRestocked []BarangRanking
	db.Table("barang_masuk_items").
		Select("barang.nama AS name, SUM(barang_masuk_items.qty) AS value").
		Joins("JOIN barang ON barang.id = barang_masuk_items.barang_id").
		Joins("JOIN barang_masuk ON barang_masuk.id = barang_masuk_items.barang_masuk_id AND barang_masuk.status = 'selesai'").
		Group("barang.nama").
		Order("value DESC").
		Limit(5).
		Scan(&topRestocked)

	var topKeluar []BarangRanking
	db.Table("barang_keluar_items").
		Select("barang.nama AS name, SUM(barang_keluar_items.qty) AS value").
		Joins("JOIN barang ON barang.id = barang_keluar_items.barang_id").
		Joins("JOIN barang_keluar ON barang_keluar.id = barang_keluar_items.barang_keluar_id AND barang_keluar.status = 'selesai'").
		Group("barang.nama").
		Order("value DESC").
		Limit(5).
		Scan(&topKeluar)

	var totalAset int64
	db.Model(&model.Asset{}).Count(&totalAset)

	var asetRusak int64
	db.Model(&model.Asset{}).Where("status = ?", "rusak").Count(&asetRusak)

	var asetPerJenis []KategoriComposition
	db.Model(&model.Asset{}).
		Select("jenis_aset AS label, COUNT(id) AS value").
		Group("jenis_aset").
		Order("value DESC").
		Scan(&asetPerJenis)

	var asetPerStatus []KategoriComposition
	db.Model(&model.Asset{}).
		Select("status AS label, COUNT(id) AS value").
		Group("status").
		Order("value DESC").
		Scan(&asetPerStatus)

	var asetPerGudang []KategoriComposition
	db.Table("assets").
		Select("gudangs.nama AS label, COUNT(assets.id) AS value").
		Joins("JOIN gudangs ON gudangs.id = assets.gudang_id").
		Group("gudangs.nama").
		Order("value DESC").
		Scan(&asetPerGudang)

	return utils.OK(c, "analisa data berhasil diambil", AnalisaResponse{
		TotalSKU:             totalSKU,
		TotalRestockBulanIni: totalRestock,
		StokMenipis:          stokMenipis,
		KategoriComposition:  kategoriRows,
		TopRestocked:         topRestocked,
		TopKeluar:            topKeluar,
		TotalAset:            totalAset,
		AsetRusak:            asetRusak,
		AsetPerJenis:         asetPerJenis,
		AsetPerStatus:        asetPerStatus,
		AsetPerGudang:        asetPerGudang,
	})
}
