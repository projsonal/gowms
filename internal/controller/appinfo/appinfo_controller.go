package appinfo

import (
	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/pkg/utils"
)

// CurrentVersion — SATU sumber kebenaran versi aplikasi yang sedang
// berjalan. Naikkan angka ini SETIAP kali rilis, dan tambahkan entri baru
// yang sesuai di changelogData di bawah (dan di changelog.yml di root
// repo, biar dua-duanya tetap sinkron — lihat catatan di file itu).
//
// CATATAN DIAGNOSTIK: kalau setelah rebuild+restart backend, GET
// /stockrsd/app/version MASIH menunjukkan versi LAMA (bukan yang ini),
// berarti binary yang sedang jalan BUKAN hasil build dari source code
// terbaru — cek lagi proses build/deploy-nya (mis. masih menjalankan
// binary lama yang di-cache oleh NSSM/PM2, atau build gagal diam-diam).
// Ini cara paling pasti membedakan "kode belum ke-update" vs "ada bug
// baru" saat suatu fitur backend sudah diperbaiki di source tapi masih
// terlihat error yang sama di aplikasi.
const CurrentVersion = "v1.2.1"

type VersionChanges struct {
	New []string `json:"new,omitempty"`
	Fix []string `json:"fix,omitempty"`
}

type VersionEntry struct {
	Version string         `json:"version"`
	Date    string         `json:"date"` // format bebas-baca, mis. "10 Agustus 2026"
	Changes VersionChanges `json:"changes"`
}

// changelogData — SAMA persis isinya dengan changelog.yml di root repo
// (lihat catatan panjang di file itu soal kenapa datanya dobel, bukan
// baca-langsung dari YAML). Urutan: rilis terbaru DULU.
var changelogData = []VersionEntry{
	{
		Version: "v1.2.1",
		Date:    "10 Agustus 2026",
		Changes: VersionChanges{
			Fix: []string{
				"Perbaikan format tanggal di Barang Masuk/Keluar, Pengiriman, Purchase Order, Stock Opname",
				"Bypass izin penuh untuk Super Admin di seluruh modul",
			},
		},
	},
	{
		Version: "v1.2.0",
		Date:    "10 Agustus 2026",
		Changes: VersionChanges{
			New: []string{"Purchase Order", "Approval Workflow"},
			Fix: []string{"Perbaikan stok minus"},
		},
	},
	{
		Version: "v1.1.0",
		Date:    "1 Agustus 2026",
		Changes: VersionChanges{
			New: []string{"OTP Login", "WhatsApp Notification"},
		},
	},
}

type Controller struct{}

func New() *Controller { return &Controller{} }

type VersionResponse struct {
	Version string `json:"version"`
	AppName string `json:"app_name"`
}

// Version GET /app/version — dipoll berkala oleh frontend (lihat
// component/system/VersionWatcher.tsx) untuk mendeteksi kalau server
// sudah di-deploy ulang dengan versi baru selagi user masih membuka
// halaman lama di browser-nya, lalu menawarkan muat ulang halaman.
// SENGAJA tanpa auth — endpoint ini perlu bisa dicek bahkan sebelum login.
func (h *Controller) Version(c *fiber.Ctx) error {
	return utils.OK(c, "versi aplikasi berhasil diambil", VersionResponse{
		Version: CurrentVersion,
		AppName: "StokRSD WMS",
	})
}

// Changelog GET /app/changelog — dipakai halaman /changelog di frontend.
func (h *Controller) Changelog(c *fiber.Ctx) error {
	return utils.OK(c, "riwayat pembaruan berhasil diambil", changelogData)
}

func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/app")
	g.Get("/version", h.Version)
	g.Get("/changelog", h.Changelog)
}
