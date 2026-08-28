package assetgudang

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/utils"
)

func toPortResponse(a model.Asset, ports []model.AssetPort) []AssetPortResponse {
	byNumber := make(map[int]model.AssetPort, len(ports))
	for _, p := range ports {
		byNumber[p.PortNumber] = p
	}
	out := make([]AssetPortResponse, a.JumlahPort)
	for i := 0; i < a.JumlahPort; i++ {
		num := i + 1
		res := AssetPortResponse{PortNumber: num, Status: "kosong"}
		if p, ok := byNumber[num]; ok {
			res.Status = p.Status
			res.CustomerName = p.CustomerName
			res.CustomerPhone = p.CustomerPhone
			res.Keterangan = p.Keterangan
			if p.ChildAssetID != nil {
				res.ChildAssetID = p.ChildAssetID
				if p.ChildAsset != nil {
					res.ChildAssetNama = p.ChildAsset.Nama
					res.ChildAssetLabel = p.ChildAsset.LabelRSD
				}
			}
		}
		out[i] = res
	}
	return out
}

func (h *Controller) ListPorts(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id aset tidak valid", nil)
	}
	a, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "aset tidak ditemukan", nil)
	}
	if a.JumlahPort <= 0 {
		return utils.OK(c, "aset ini tidak punya port", []AssetPortResponse{})
	}
	ports, err := h.portRepo.ListByAsset(a.ID)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data port", nil)
	}
	return utils.OK(c, "data port berhasil diambil", toPortResponse(*a, ports))
}

func (h *Controller) SetPort(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id aset tidak valid", nil)
	}
	portNumber, err := strconv.Atoi(c.Params("nomor"))
	if err != nil || portNumber <= 0 {
		return utils.Fail(c, fiber.StatusBadRequest, "nomor port tidak valid", nil)
	}
	a, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "aset tidak ditemukan", nil)
	}
	if portNumber > a.JumlahPort {
		return utils.Fail(c, fiber.StatusBadRequest, "nomor port melebihi jumlah port aset ini", nil)
	}

	var req AssetPortRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	if req.ChildAssetID != nil && req.CustomerName != "" {
		return utils.Fail(c, fiber.StatusBadRequest,
			"port cuma bisa tersambung ke SALAH SATU: aset lain (child_asset_id) atau pelanggan (customer_name), tidak dua-duanya", nil)
	}
	var child *model.Asset
	if req.ChildAssetID != nil {
		if *req.ChildAssetID == a.ID {
			return utils.Fail(c, fiber.StatusUnprocessableEntity, "aset tidak bisa tersambung ke dirinya sendiri", nil)
		}
		child, err = h.repo.FindByID(*req.ChildAssetID)
		if err != nil {
			return utils.Fail(c, fiber.StatusBadRequest, "aset tujuan tidak ditemukan", nil)
		}
		if !model.JenisIndukValid(child.JenisAset, a.JenisAset) {
			return utils.Fail(c, fiber.StatusUnprocessableEntity,
				fmt.Sprintf("%s tidak bisa berinduk ke %s — cek urutan hierarki jaringan (OLT -> ODC -> ODP -> ONT)", child.JenisAset, a.JenisAset), nil)
		}
	}

	status := "kosong"
	if req.ChildAssetID != nil || req.CustomerName != "" {
		status = "terisi"
	}

	port := &model.AssetPort{
		AssetID: a.ID, PortNumber: portNumber, Status: status,
		ChildAssetID: req.ChildAssetID, CustomerName: req.CustomerName,
		CustomerPhone: req.CustomerPhone, Keterangan: req.Keterangan,
	}
	if err := h.portRepo.Upsert(port); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menyimpan data port", nil)
	}

	if child != nil {
		child.ParentAssetID = &a.ID
		_ = h.repo.Update(child)
	}

	ports, err := h.portRepo.ListByAsset(a.ID)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data port", nil)
	}
	return utils.OK(c, "port berhasil disimpan", toPortResponse(*a, ports))
}

func (h *Controller) ClearPort(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id aset tidak valid", nil)
	}
	portNumber, err := strconv.Atoi(c.Params("nomor"))
	if err != nil || portNumber <= 0 {
		return utils.Fail(c, fiber.StatusBadRequest, "nomor port tidak valid", nil)
	}
	if err := h.portRepo.Clear(id, portNumber); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengosongkan port", nil)
	}
	return utils.OK(c, "port berhasil dikosongkan", nil)
}
