package appinfo

import (
	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/internal/middleware"
	"github.com/projsonal/gowms/pkg/constant"
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

type VersionResponse struct {
	Version     string `json:"version"`
	AppName     string `json:"app_name"`
	Description string `json:"description"`
	Developer   string `json:"developer"`
}

func (h *Controller) Version(c *fiber.Ctx) error {
	return utils.OK(c, "versi aplikasi berhasil diambil", VersionResponse{
		Version:     CurrentVersion,
		AppName:     "WMS - RSD",
		Description: "WMS-RSD merupakan pelayanan gudang serta inventaris produk dalam perusahaan — mengelola stok, pengiriman, aset gudang (tiang/ODC/ONT/ODP/OLT/transportasi), hingga laporan operasional dalam satu sistem.",
		Developer:   "Tim Internal RSD",
	})
}

func (h *Controller) Changelog(c *fiber.Ctx) error {
	return utils.OK(c, "riwayat pembaruan berhasil diambil", changelogData)
}

func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/app")
	// Publik (tidak butuh login) — cuma menampilkan info aplikasi, tidak
	// ada data sensitif.
	g.Get("/version", h.Version)
	g.Get("/changelog", h.Changelog)

	// Cek/Update — WAJIB login (beda dari /version & /changelog di atas):
	// CheckUpdate & UpdateStatus murni baca, boleh role apa saja yang
	// sudah login; TriggerUpdate KHUSUS super_admin karena efeknya besar
	// (menyalakan Mode Pemeliharaan otomatis & me-restart proses backend
	// lewat skrip deploy — lihat update_controller.go).
	loggedIn := middleware.JWTAuth(h.jwtSvc)
	onlySuperAdmin := middleware.RequireRole(constant.RoleSuperAdmin)

	g.Get("/check-update", loggedIn, h.CheckUpdate)
	g.Get("/update-status", loggedIn, h.UpdateStatus)
	g.Post("/update", loggedIn, onlySuperAdmin, h.TriggerUpdate)
}
