package dashboard

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/pkg/utils"
)

const dashboardDateFormat = "2006-01-02"

func (h *Controller) ReportPreview(c *fiber.Ctx) error {
	jenis := c.Params("jenis")
	fromStr := c.Query("from", "")
	toStr := c.Query("to", "")
	from, _ := parseDate(fromStr)
	to, _ := parseDate(toStr)
	if to.IsZero() {
		to = time.Now()
	}
	if from.IsZero() {
		from = to.AddDate(0, 0, -30)
	}
	to = time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 0, to.Location())

	db := h.db
	type entry struct {
		Nomor   string    `json:"nomor"`
		Tanggal time.Time `json:"tanggal"`
		Info    string    `json:"info"`
		Status  string    `json:"status"`
		Qty     int64     `json:"qty"`
	}
	var rows []entry

	switch jenis {
	case "barang-masuk":
		db.Raw(`SELECT bm.nomor_penerimaan AS nomor, bm.tanggal, COALESCE(g.nama, '-') AS info, bm.status,
		         COALESCE((SELECT COUNT(*) FROM barang_masuk_items WHERE barang_masuk_id = bm.id), 0) AS qty
		         FROM barang_masuk bm LEFT JOIN gudangs g ON g.id = bm.gudang_id
		         WHERE bm.tanggal BETWEEN ? AND ? ORDER BY bm.tanggal DESC`, from, to).Scan(&rows)
	case "barang-keluar":
		db.Raw(`SELECT bk.nomor_pengeluaran AS nomor, bk.tanggal, COALESCE(g.nama, '-') AS info, bk.status,
		         COALESCE((SELECT COUNT(*) FROM barang_keluar_items WHERE barang_keluar_id = bk.id), 0) AS qty
		         FROM barang_keluar bk LEFT JOIN gudangs g ON g.id = bk.gudang_id
		         WHERE bk.tanggal BETWEEN ? AND ? ORDER BY bk.tanggal DESC`, from, to).Scan(&rows)
	case "stock-opname":
		db.Raw(`SELECT so.nomor_opname AS nomor, so.tanggal, COALESCE(g.nama, '-') AS info, so.status, 0 AS qty
		         FROM stock_opnames so LEFT JOIN gudangs g ON g.id = so.gudang_id
		         WHERE so.tanggal BETWEEN ? AND ? ORDER BY so.tanggal DESC`, from, to).Scan(&rows)
	case "stok":
		db.Raw(`SELECT b.kode_barang AS nomor, b.updated_at AS tanggal,
		         COALESCE(k.nama, '-') || ' - ' || b.nama AS info,
		         CASE WHEN b.stok <= b.stok_minimum THEN 'Kritis' ELSE 'Aman' END AS status,
		         b.stok AS qty FROM barangs b LEFT JOIN kategoris k ON k.id = b.kategori_id ORDER BY b.stok ASC`).Scan(&rows)
	default:
		return utils.Fail(c, fiber.StatusBadRequest, "jenis laporan tidak dikenal", nil)
	}

	if rows == nil {
		rows = []entry{}
	}
	return utils.OK(c, "preview laporan", fiber.Map{
		"jenis": jenis,
		"from":  from.Format(dashboardDateFormat),
		"to":    to.Format(dashboardDateFormat),
		"count": len(rows),
		"rows":  rows,
	})
}

func parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, s)
}
