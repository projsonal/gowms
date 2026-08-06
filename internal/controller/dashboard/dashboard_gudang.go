package dashboard

import (
	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/utils"
)

type RakLoad struct {
	ID        uint   `json:"id"`
	KodeRak   string `json:"kode_rak"`
	GudangID  uint   `json:"gudang_id"`
	Kapasitas int    `json:"kapasitas"`
	Terisi    int    `json:"terisi"`
	Kosong    int    `json:"kosong"`
	Status    string `json:"status"`
	PctTerisi int    `json:"pct_terisi"`
}

type GudangLoad struct {
	GudangID       uint      `json:"gudang_id"`
	Nama           string    `json:"nama"`
	Alamat         string    `json:"alamat"`
	TotalKapasitas int       `json:"total_kapasitas"`
	TotalTerisi    int       `json:"total_terisi"`
	TotalKosong    int       `json:"total_kosong"`
	PctTerisi      int       `json:"pct_terisi"`
	RakPenuh       int       `json:"rak_penuh"`
	RakKosong      int       `json:"rak_kosong"`
	RakSebagian    int       `json:"rak_sebagian"`
	Raks           []RakLoad `json:"raks"`
}

func (h *Controller) GudangBeban(c *fiber.Ctx) error {
	db := h.db
	var gudangs []model.Gudang
	if err := db.Preload("Raks").Order("id ASC").Find(&gudangs).Error; err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	out := make([]GudangLoad, 0, len(gudangs))
	for _, g := range gudangs {
		gl := GudangLoad{GudangID: g.ID, Nama: g.Nama, Alamat: g.Alamat, Raks: []RakLoad{}}
		for _, r := range g.Raks {
			kosong := r.Kapasitas - r.Terisi
			if kosong < 0 {
				kosong = 0
			}
			pct := 0
			if r.Kapasitas > 0 {
				pct = (r.Terisi * 100) / r.Kapasitas
				if pct > 100 {
					pct = 100
				}
			}
			gl.Raks = append(gl.Raks, RakLoad{
				ID: r.ID, KodeRak: r.KodeRak, GudangID: r.GudangID,
				Kapasitas: r.Kapasitas, Terisi: r.Terisi, Kosong: kosong,
				Status: r.Status, PctTerisi: pct,
			})
			gl.TotalKapasitas += r.Kapasitas
			gl.TotalTerisi += r.Terisi
			gl.TotalKosong += kosong
			switch r.Status {
			case "penuh":
				gl.RakPenuh++
			case "kosong":
				gl.RakKosong++
			case "terisi_sebagian":
				gl.RakSebagian++
			}
		}
		if gl.TotalKapasitas > 0 {
			gl.PctTerisi = (gl.TotalTerisi * 100) / gl.TotalKapasitas
			if gl.PctTerisi > 100 {
				gl.PctTerisi = 100
			}
		}
		out = append(out, gl)
	}

	return utils.OK(c, "beban gudang berhasil diambil", fiber.Map{
		"gudangs": out,
	})
}
