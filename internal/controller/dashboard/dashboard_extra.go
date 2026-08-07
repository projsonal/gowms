package dashboard

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/utils"
	"gorm.io/gorm"
)

type TrendPoint struct {
	Bulan  string `json:"bulan"`
	Masuk  int64  `json:"masuk"`
	Keluar int64  `json:"keluar"`
}

func (h *Controller) Trend(c *fiber.Ctx) error {
	db := h.db
	now := time.Now()
	// Build 6 months back including current
	months := make([]time.Time, 6)
	for i := 0; i < 6; i++ {
		m := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -(5 - i), 0)
		months[i] = m
	}
	monthNames := []string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}
	out := make([]TrendPoint, 0, 6)
	for _, m := range months {
		start := m
		end := m.AddDate(0, 1, 0)
		var masuk int64
		var keluar int64
		db.Model(&model.BarangMasuk{}).Where("tanggal >= ? AND tanggal < ?", start, end).Count(&masuk)
		db.Model(&model.BarangKeluar{}).Where("tanggal >= ? AND tanggal < ?", start, end).Count(&keluar)
		out = append(out, TrendPoint{
			Bulan:  monthNames[m.Month()-1],
			Masuk:  masuk,
			Keluar: keluar,
		})
	}
	return utils.OK(c, "tren 6 bulan berhasil diambil", out)
}

type ActivityItem struct {
	ID   string    `json:"id"`
	User string    `json:"user"`
	Act  string    `json:"act"`
	Time string    `json:"time"`
	Type string    `json:"type"`
	At   time.Time `json:"at"`
}

func (h *Controller) Activity(c *fiber.Ctx) error {
	db := h.db
	type row struct {
		ID        string
		UserName  string
		Nomor     string
		Type      string
		CreatedAt time.Time
	}
	var rows []row

	q := `
	  SELECT 'bm-' || bm.id::text AS id, COALESCE(u.full_name, u.username, 'System') AS user_name, bm.nomor_penerimaan AS nomor, 'in' AS type, bm.created_at
	  FROM barang_masuk bm LEFT JOIN users u ON u.id = bm.diterima_oleh
	  UNION ALL
	  SELECT 'bk-' || bk.id::text, COALESCE(u.full_name, u.username, 'System'), bk.nomor_pengeluaran, 'out', bk.created_at
	  FROM barang_keluar bk LEFT JOIN users u ON u.id = bk.dikeluarkan_oleh
	  UNION ALL
	  SELECT 'po-' || po.id::text, COALESCE(u.full_name, u.username, 'System'), po.nomor_po, 'po', po.created_at
	  FROM purchase_orders po LEFT JOIN users u ON u.id = po.diajukan_oleh
	  UNION ALL
	  SELECT 'pg-' || pg.id::text, 'System', pg.nomor_pengiriman, 'ship', pg.created_at
	  FROM pengiriman pg
	  UNION ALL
	  SELECT 'so-' || so.id::text, 'System', so.nomor_opname, 'opname', so.created_at
	  FROM stock_opname so
	  ORDER BY created_at DESC LIMIT 15
	`
	if err := db.Raw(q).Scan(&rows).Error; err != nil {
		return utils.OK(c, "aktivitas terbaru berhasil diambil", []ActivityItem{})
	}
	out := make([]ActivityItem, 0, len(rows))
	for _, r := range rows {
		verb := map[string]string{
			"in":     "menambahkan Barang Masuk",
			"out":    "mencatat Barang Keluar",
			"po":     "membuat Purchase Order",
			"ship":   "membuat Pengiriman",
			"opname": "melakukan Stock Opname",
		}[r.Type]
		out = append(out, ActivityItem{
			ID:   r.ID,
			User: r.UserName,
			Act:  verb + " " + r.Nomor,
			Time: humanTime(r.CreatedAt),
			Type: r.Type,
			At:   r.CreatedAt,
		})
	}
	return utils.OK(c, "aktivitas terbaru berhasil diambil", out)
}

func (h *Controller) Notifications(c *fiber.Ctx) error {
	db := h.db
	sinceStr := c.Query("since", "")
	var since time.Time
	if sinceStr != "" {
		if t, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			since = t
		}
	}
	type row struct {
		ID        string
		Title     string
		Body      string
		Kind      string
		CreatedAt time.Time
	}
	var rows []row
	sinceFilter := ""
	args := []any{}
	if !since.IsZero() {
		sinceFilter = "WHERE created_at > ?"
		args = append(args, since, since, since, since, since)
	}
	_ = sinceFilter
	// Untuk barang_masuk & barang_keluar, karyawan HANYA perlu diberi tahu
	// saat statusnya sudah "selesai" (disetujui admin/super_admin lewat
	// endpoint PATCH .../selesai — lihat middleware `edit` di RegisterRoutes,
	// karyawan biasanya tidak punya izin `edit` sehingga transisi status ini
	// pasti dilakukan admin/super_admin). Dipakai `completed_at`, BUKAN
	// `created_at`, supaya notifikasi muncul saat disetujui, bukan saat
	// pertama kali diinput oleh karyawan.
	q := `
	  SELECT 'bm-' || bm.id::text AS id, 'Barang Masuk Disetujui' AS title, bm.nomor_penerimaan AS body, 'in_approved' AS kind, bm.completed_at AS created_at
	  FROM barang_masuk bm
	  WHERE bm.status = 'selesai' AND bm.completed_at IS NOT NULL ` + optAnd(!since.IsZero(), "bm.completed_at") + `
	  UNION ALL
	  SELECT 'bk-' || bk.id::text, 'Barang Keluar Disetujui', bk.nomor_pengeluaran, 'out_approved', bk.completed_at
	  FROM barang_keluar bk
	  WHERE bk.status = 'selesai' AND bk.completed_at IS NOT NULL ` + optAnd(!since.IsZero(), "bk.completed_at") + `
	  UNION ALL
	  SELECT 'pg-' || pg.id::text, CASE WHEN pg.status='Terkirim' THEN 'Pengiriman Selesai' ELSE 'Pengiriman' END, pg.nomor_pengiriman, 'ship', pg.created_at
	  FROM pengiriman pg ` + optWhere(!since.IsZero()) + `
	  UNION ALL
	  SELECT 'po-' || po.id::text, 'Purchase Order', po.nomor_po, 'po', po.created_at
	  FROM purchase_orders po ` + optWhere(!since.IsZero()) + `
	  UNION ALL
	  SELECT 'so-' || so.id::text, 'Stock Opname', so.nomor_opname, 'opname', so.created_at
	  FROM stock_opname so ` + optWhere(!since.IsZero()) + `
	  ORDER BY created_at DESC LIMIT 20
	`
	tx := db.Raw(q, args...).Scan(&rows)
	if tx.Error != nil {
		return utils.OK(c, "notifikasi kosong", []map[string]any{})
	}
	out := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		out = append(out, map[string]any{
			"id":         r.ID,
			"title":      r.Title,
			"body":       r.Body,
			"kind":       r.Kind,
			"created_at": r.CreatedAt.Format(time.RFC3339),
			"time":       humanTime(r.CreatedAt),
		})
	}
	return utils.OK(c, "notifikasi berhasil diambil", out)
}

func optWhere(with bool) string {
	if with {
		return "WHERE created_at > ?"
	}
	return ""
}

// optAnd sama seperti optWhere, tapi untuk ditambahkan setelah klausa WHERE
// yang sudah ada (bm.status = 'selesai' dst.) dan bisa memakai kolom lain
// selain created_at (di sini: completed_at).
func optAnd(with bool, column string) string {
	if with {
		return "AND " + column + " > ?"
	}
	return ""
}

func humanTime(t time.Time) string {
	d := time.Since(t)
	if d < time.Minute {
		return "baru saja"
	}
	if d < time.Hour {
		mins := int(d.Minutes())
		return itoa(mins) + " menit lalu"
	}
	if d < 24*time.Hour {
		hrs := int(d.Hours())
		return itoa(hrs) + " jam lalu"
	}
	days := int(d.Hours() / 24)
	return itoa(days) + " hari lalu"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := make([]byte, 0, 8)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		return "-" + string(buf)
	}
	return string(buf)
}

var _ *gorm.DB = nil
