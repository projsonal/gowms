package assetgudang

import (
	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

func (h *Controller) logHistory(c *fiber.Ctx, assetID uint, eventType, fieldLama, fieldBaru, catatan string) {
	userID, _ := c.Locals(constant.CtxUserID).(uint)
	var uid *uint
	userNama := ""
	if userID != 0 {
		uid = &userID
		if u, err := h.usersRepo.FindByID(userID); err == nil {
			userNama = u.FullName
		}
	}
	_ = h.historyRepo.Log(&model.AssetHistory{
		AssetID: assetID, EventType: eventType, FieldLama: fieldLama, FieldBaru: fieldBaru,
		Catatan: catatan, UserID: uid, UserNama: userNama,
	})
}

func (h *Controller) ListHistory(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id aset tidak valid", nil)
	}
	rows, err := h.historyRepo.ListByAsset(id, 100)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil riwayat aset", nil)
	}
	out := make([]AssetHistoryResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, AssetHistoryResponse{
			ID: r.ID, EventType: r.EventType, FieldLama: r.FieldLama, FieldBaru: r.FieldBaru,
			Catatan: r.Catatan, UserNama: r.UserNama, CreatedAt: r.CreatedAt,
		})
	}
	return utils.OK(c, "riwayat aset berhasil diambil", out)
}
