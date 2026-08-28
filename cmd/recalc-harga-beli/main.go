package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/config"
	"github.com/projsonal/gowms/pkg/constant"
)

func main() {
	apply := flag.Bool("apply", false, "eksekusi sungguhan (menimpa Barang.HargaBeli); tanpa flag ini cuma dry-run")
	flag.Parse()

	cfg := config.Load()
	db := config.NewDatabase(cfg)

	var barangList []model.Barang
	if err := db.Find(&barangList).Error; err != nil {
		log.Fatalf("gagal mengambil daftar barang: %v", err)
	}

	backupName := fmt.Sprintf("recalc-harga-beli-backup-%s.csv", time.Now().Format("20060102-150405"))
	backupFile, err := os.Create(backupName)
	if err != nil {
		log.Fatalf("gagal membuat file backup %s: %v — DIBATALKAN demi keamanan (tidak ada backup, tidak ada perubahan)", backupName, err)
	}
	defer backupFile.Close()
	csvWriter := csv.NewWriter(backupFile)
	defer csvWriter.Flush()
	_ = csvWriter.Write([]string{"kode_barang", "nama", "harga_lama", "harga_baru", "jumlah_baris_barang_masuk", "mode"})
	mode := "dry-run"
	if *apply {
		mode = "apply"
	}

	type row struct {
		BarangMasukID uint
		Tanggal       time.Time
		Qty           int
		HargaSatuan   int64
	}

	var totalDiubah int
	for _, b := range barangList {
		var items []row
		if err := db.Table("barang_masuk_items bmi").
			Select("bmi.barang_masuk_id, bm.tanggal, bmi.qty, bmi.harga_satuan").
			Joins("JOIN barang_masuk bm ON bm.id = bmi.barang_masuk_id").
			Where("bmi.barang_id = ? AND bm.status = ?", b.ID, constant.StatusBMSelesai).
			Scan(&items).Error; err != nil {
			log.Printf("[LEWATI] %s (%s): gagal ambil riwayat barang masuk: %v", b.KodeBarang, b.Nama, err)
			continue
		}
		if len(items) == 0 {
			continue
		}
		sort.Slice(items, func(i, j int) bool { return items[i].Tanggal.Before(items[j].Tanggal) })

		var stokRunning int
		var hargaRunning int64
		var adaBarisBerharga bool
		for _, it := range items {
			if it.HargaSatuan > 0 {
				totalNilai := int64(stokRunning)*hargaRunning + int64(it.Qty)*it.HargaSatuan
				stokRunning += it.Qty
				hargaRunning = totalNilai / int64(stokRunning)
				adaBarisBerharga = true
			} else {
				stokRunning += it.Qty
			}
		}
		if !adaBarisBerharga || hargaRunning == b.HargaBeli {
			continue
		}

		fmt.Printf("%s (%s): Rp%d -> Rp%d (dari %d baris Barang Masuk berharga)\n",
			b.KodeBarang, b.Nama, b.HargaBeli, hargaRunning, len(items))
		totalDiubah++
		_ = csvWriter.Write([]string{
			b.KodeBarang, b.Nama, fmt.Sprintf("%d", b.HargaBeli), fmt.Sprintf("%d", hargaRunning),
			fmt.Sprintf("%d", len(items)), mode,
		})

		if *apply {
			if err := db.Model(&model.Barang{}).Where("id = ?", b.ID).
				Update("harga_beli", hargaRunning).Error; err != nil {
				log.Printf("  [GAGAL SIMPAN] %s: %v", b.KodeBarang, err)
			}
		}
	}
	csvWriter.Flush()

	if *apply {
		fmt.Printf("\nSelesai — %d SKU diperbarui. Backup nilai lama/baru tersimpan di %s (dipakai untuk rollback manual kalau perlu).\n", totalDiubah, backupName)
	} else {
		fmt.Printf("\nDRY-RUN — %d SKU AKAN diperbarui kalau dijalankan dengan --apply. Belum ada perubahan ke database. Pratinjau nilai lama/baru tersimpan di %s.\n", totalDiubah, backupName)
	}
}
