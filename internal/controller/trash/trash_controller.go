// Package trash mengimplementasikan fitur "Tempat Sampah" generik: barang
// yang dihapus dari UI (Manajemen Aset Gudang, Kelola Barang, Manajemen
// Gudang, Barang Rusak) TIDAK langsung hilang permanen — GORM menandainya
// `deleted_at` (soft-delete, lihat kolom DeletedAt di masing-masing model)
// sehingga otomatis tersembunyi dari query normal TAPI masih ada di
// database dan bisa dipulihkan (restore) atau baru benar-benar dihapus
// (hard delete) lewat menu ini.
//
// SENGAJA generik satu controller untuk semua tipe (bukan endpoint
// terpisah per modul) supaya menambah modul baru ke Tempat Sampah cukup
// menambah satu entri di `registry` di bawah — tidak perlu controller baru.
package trash

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/projsonal/gowms/internal/middleware"
	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

// entry — satu tipe data yang ikut fitur Tempat Sampah.
type entry struct {
	// label: dipakai di query ?type=... dan ditampilkan ke user.
	label string
	// nama, keterangan: field yang dibaca lewat reflection-free type switch
	// di summarize() untuk menampilkan ringkasan tiap baris sampah tanpa
	// perlu bentuk response terpisah per tipe.
}

var registry = map[string]entry{
	"aset":         {label: "Aset Gudang"},
	"barang":       {label: "Barang"},
	"gudang":       {label: "Gudang"},
	"barang_rusak": {label: "Barang Rusak"},
}

// Item — bentuk ringkas satu baris di Tempat Sampah, seragam untuk semua tipe.
type Item struct {
	Type      string    `json:"type"`
	ID        uint      `json:"id"`
	Judul     string    `json:"judul"`
	Subjudul  string    `json:"subjudul,omitempty"`
	DeletedAt time.Time `json:"deleted_at"`
}

type Controller struct {
	db     *gorm.DB
	jwtSvc *utils.JWTService
}

func New(db *gorm.DB, jwtSvc *utils.JWTService) *Controller {
	return &Controller{db: db, jwtSvc: jwtSvc}
}

func (h *Controller) modelForType(t string) (any, bool) {
	switch t {
	case "aset":
		return &model.Asset{}, true
	case "barang":
		return &model.Barang{}, true
	case "gudang":
		return &model.Gudang{}, true
	case "barang_rusak":
		return &model.BarangRusak{}, true
	default:
		return nil, false
	}
}

func summarize(t string) (judul func(any) string, subjudul func(any) string) {
	switch t {
	case "aset":
		return func(m any) string { return m.(*model.Asset).Nama },
			func(m any) string { return m.(*model.Asset).LabelRSD }
	case "barang":
		return func(m any) string { return m.(*model.Barang).Nama },
			func(m any) string { return m.(*model.Barang).KodeBarang }
	case "gudang":
		return func(m any) string { return m.(*model.Gudang).Nama },
			func(m any) string { return m.(*model.Gudang).Kode }
	case "barang_rusak":
		return func(m any) string { return m.(*model.BarangRusak).NamaBarang },
			func(m any) string { return m.(*model.BarangRusak).LabelBarang }
	default:
		return func(any) string { return "" }, func(any) string { return "" }
	}
}

// List GET /trash?type=aset|barang|gudang|barang_rusak — kalau `type`
// dikosongkan, kembalikan gabungan semua tipe (dipakai badge counter di
// ikon Tempat Sampah pada header).
func (h *Controller) List(c *fiber.Ctx) error {
	reqType := c.Query("type", "")
	types := []string{reqType}
	if reqType == "" {
		types = make([]string, 0, len(registry))
		for t := range registry {
			types = append(types, t)
		}
	} else if _, ok := registry[reqType]; !ok {
		return utils.Fail(c, fiber.StatusBadRequest, "tipe tidak dikenal", nil)
	}

	out := make([]Item, 0)
	for _, t := range types {
		modelPtr, _ := h.modelForType(t)
		judulFn, subjudulFn := summarize(t)

		// Unscoped(): WAJIB, supaya GORM tidak otomatis menambahkan
		// `WHERE deleted_at IS NULL` seperti query normal — di sini kita
		// justru mau baris yang SUDAH ter-soft-delete.
		rows, err := queryDeleted(h.db, t, modelPtr)
		if err != nil {
			return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data tempat sampah", nil)
		}
		for _, r := range rows {
			out = append(out, Item{
				Type: t, ID: r.id, Judul: judulFn(r.model), Subjudul: subjudulFn(r.model), DeletedAt: r.deletedAt,
			})
		}
	}
	return utils.OK(c, "data tempat sampah berhasil diambil", out)
}

type deletedRow struct {
	id        uint
	deletedAt time.Time
	model     any
}

func queryDeleted(db *gorm.DB, t string, modelPtr any) ([]deletedRow, error) {
	out := make([]deletedRow, 0)
	switch t {
	case "aset":
		var rows []model.Asset
		if err := db.Unscoped().Where("deleted_at IS NOT NULL").Order("deleted_at DESC").Find(&rows).Error; err != nil {
			return nil, err
		}
		for i := range rows {
			out = append(out, deletedRow{id: rows[i].ID, deletedAt: rows[i].DeletedAt.Time, model: &rows[i]})
		}
	case "barang":
		var rows []model.Barang
		if err := db.Unscoped().Where("deleted_at IS NOT NULL").Order("deleted_at DESC").Find(&rows).Error; err != nil {
			return nil, err
		}
		for i := range rows {
			out = append(out, deletedRow{id: rows[i].ID, deletedAt: rows[i].DeletedAt.Time, model: &rows[i]})
		}
	case "gudang":
		var rows []model.Gudang
		if err := db.Unscoped().Where("deleted_at IS NOT NULL").Order("deleted_at DESC").Find(&rows).Error; err != nil {
			return nil, err
		}
		for i := range rows {
			out = append(out, deletedRow{id: rows[i].ID, deletedAt: rows[i].DeletedAt.Time, model: &rows[i]})
		}
	case "barang_rusak":
		var rows []model.BarangRusak
		if err := db.Unscoped().Where("deleted_at IS NOT NULL").Order("deleted_at DESC").Find(&rows).Error; err != nil {
			return nil, err
		}
		for i := range rows {
			out = append(out, deletedRow{id: rows[i].ID, deletedAt: rows[i].DeletedAt.Time, model: &rows[i]})
		}
	}
	_ = modelPtr
	return out, nil
}

// Restore POST /trash/:type/:id/restore — batalkan soft-delete (set
// deleted_at kembali NULL), baris langsung muncul lagi di modul aslinya.
func (h *Controller) Restore(c *fiber.Ctx) error {
	t := c.Params("type")
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return utils.Fail(c, fiber.StatusBadRequest, "id tidak valid", nil)
	}
	modelPtr, ok := h.modelForType(t)
	if !ok {
		return utils.Fail(c, fiber.StatusBadRequest, "tipe tidak dikenal", nil)
	}
	res := h.db.Unscoped().Model(modelPtr).Where("id = ? AND deleted_at IS NOT NULL", id).Update("deleted_at", nil)
	if res.Error != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memulihkan data", nil)
	}
	if res.RowsAffected == 0 {
		return utils.Fail(c, fiber.StatusNotFound, "data tidak ditemukan di tempat sampah", nil)
	}
	return utils.OK(c, "data berhasil dipulihkan", nil)
}

// Purge DELETE /trash/:type/:id — hapus PERMANEN, tidak bisa dibatalkan.
// Dibatasi Super Admin & Admin lewat middleware di RegisterRoutes.
func (h *Controller) Purge(c *fiber.Ctx) error {
	t := c.Params("type")
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return utils.Fail(c, fiber.StatusBadRequest, "id tidak valid", nil)
	}
	modelPtr, ok := h.modelForType(t)
	if !ok {
		return utils.Fail(c, fiber.StatusBadRequest, "tipe tidak dikenal", nil)
	}
	res := h.db.Unscoped().Where("id = ? AND deleted_at IS NOT NULL", id).Delete(modelPtr)
	if res.Error != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menghapus permanen", nil)
	}
	if res.RowsAffected == 0 {
		return utils.Fail(c, fiber.StatusNotFound, "data tidak ditemukan di tempat sampah", nil)
	}
	return utils.OK(c, "data berhasil dihapus permanen", nil)
}

// Dibatasi Super Admin & Admin — melihat, memulihkan, ATAU menghapus
// permanen data orang lain adalah aksi administratif, karyawan tidak
// diberi akses ke Tempat Sampah sama sekali.
func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/trash", middleware.JWTAuth(h.jwtSvc), middleware.RequireRole(constant.RoleSuperAdmin, constant.RoleAdmin))
	g.Get("/", h.List)
	g.Post("/:type/:id/restore", h.Restore)
	g.Delete("/:type/:id", h.Purge)
}
