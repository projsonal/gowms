package appinfo

import (
	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/pkg/utils"
)


const CurrentVersion = "v1.3.0"

type VersionChanges struct {
	New []string `json:"new,omitempty"`
	Fix []string `json:"fix,omitempty"`
}

type VersionEntry struct {
	Version string         `json:"version"`
	Date    string         `json:"date"`
	Changes VersionChanges `json:"changes"`
}

var changelogData = []VersionEntry{
	{
		Version: "v1.3.0",
		Date:    "11 Agustus 2026",
		Changes: VersionChanges{
			New: []string{
				"Titik koordinat gudang di peta Pickup & Dropoff",
				"Kerjasama Kurir di data Supplier (menggantikan Email) — Total Order & Rating dihitung otomatis dari data pengiriman",
				"Resi pengiriman didesain ulang: barcode besar, SKU berawalan WRSD-, Order ID, berat barang, kontak pengirim lengkap",
				"Cetak (Print) tersedia di seluruh tabel & laporan untuk Super Admin",
				"Widget Traffic Pengiriman & Rekap Data di dashboard Super Admin",
			},
			Fix: []string{
				"Matrix perizinan sekarang benar-benar mengendalikan tombol Tambah/Ubah/Cetak di setiap modul (sebelumnya sebagian tombol staff-only walau izin sudah diaktifkan)",
				"Kapasitas Terpakai & Rak Penuh di Manajemen Gudang dihitung otomatis dari data rak",
				"Peta pelacakan pengiriman tidak lagi error saat lokasi belum tersedia",
			},
		},
	},
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
			New: []string{"tambah Menu", "integrasi ke backend"},
			Fix: []string{"Perbaikan menu dashboard"},
		},
	},
	{
		Version: "v1.1.0",
		Date:    "1 Agustus 2026",
		Changes: VersionChanges{
			New: []string{"Auth register,login", "lupa password", "Dashboard"},
		},
	},
}

type Controller struct{}

func New() *Controller { return &Controller{} }

type VersionResponse struct {
	Version string `json:"version"`
	AppName string `json:"app_name"`
}

func (h *Controller) Version(c *fiber.Ctx) error {
	return utils.OK(c, "versi aplikasi berhasil diambil", VersionResponse{
		Version: CurrentVersion,
		AppName: "WMS - RSD",
	})
}

func (h *Controller) Changelog(c *fiber.Ctx) error {
	return utils.OK(c, "riwayat pembaruan berhasil diambil", changelogData)
}

func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/app")
	g.Get("/version", h.Version)
	g.Get("/changelog", h.Changelog)
}
