package task

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/internal/middleware"
	"github.com/projsonal/gowms/internal/model"
	taskRepo "github.com/projsonal/gowms/internal/repositories/task"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

func parseIDParam(c *fiber.Ctx) (uint, error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func parseTanggal(raw string) (time.Time, error) {
	return time.Parse("2006-01-02", raw)
}

func (h *Controller) List(c *fiber.Ctx) error {
	p := utils.PaginationFromContext(c)
	f := taskRepo.Filter{Status: c.Query("status", "")}

	roleName, _ := c.Locals(constant.CtxRoleName).(string)
	if roleName == constant.RoleKaryawan {
		userID, _ := c.Locals(constant.CtxUserID).(uint)
		f.AssignedTo = userID
	}

	list, total, err := h.repo.List(p, f)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar tugas", nil)
	}
	return utils.OKWithMeta(c, "daftar tugas berhasil diambil", list, utils.BuildPaginationMeta(p, total))
}

// Detail GET /tasks/:id
func (h *Controller) Detail(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id tugas tidak valid", nil)
	}
	t, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "tugas tidak ditemukan", nil)
	}

	roleName, _ := c.Locals(constant.CtxRoleName).(string)
	userID, _ := c.Locals(constant.CtxUserID).(uint)
	if roleName == constant.RoleKaryawan && t.AssignedTo != userID {
		return utils.Fail(c, fiber.StatusNotFound, "tugas tidak ditemukan", nil)
	}
	return utils.OK(c, "detail tugas berhasil diambil", t)
}

// Summary GET /tasks/summary — kartu ringkasan, dihitung dari sudut
// pandang requester (karyawan lihat ringkasan tugasnya sendiri saja).
func (h *Controller) Summary(c *fiber.Ctx) error {
	roleName, _ := c.Locals(constant.CtxRoleName).(string)
	var assignedTo uint
	if roleName == constant.RoleKaryawan {
		assignedTo, _ = c.Locals(constant.CtxUserID).(uint)
	}

	total, err := h.repo.CountByStatus(assignedTo, "")
	if err != nil {
		total, _ = h.repo.CountByStatus(assignedTo, "baru")
	}
	proses, _ := h.repo.CountByStatus(assignedTo, "proses")
	selesai, _ := h.repo.CountByStatus(assignedTo, "selesai")
	terlambat, _ := h.repo.CountOverdue(assignedTo)

	return utils.OK(c, "ringkasan tugas berhasil diambil", SummaryResponse{
		Total: total, Proses: proses, Terlambat: terlambat, Selesai: selesai,
	})
}

// Create POST /tasks — aksi "Add" di action bar tabel. HANYA super_admin
// & admin (lihat RegisterRoutes) yang boleh menugaskan tugas baru.
func (h *Controller) Create(c *fiber.Ctx) error {
	var req TaskRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	dueDate, err := parseTanggal(req.DueDate)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "format tanggal tidak valid (YYYY-MM-DD)", nil)
	}

	assignedBy, _ := c.Locals(constant.CtxUserID).(uint)
	t := &model.Task{
		Title:       req.Title,
		Description: req.Description,
		AssignedTo:  req.AssignedTo,
		AssignedBy:  assignedBy,
		DueDate:     dueDate,
		Priority:    req.Priority,
		Status:      "baru",
	}
	if err := h.repo.Create(t); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat tugas", nil)
	}
	created, _ := h.repo.FindByID(t.ID)
	return utils.Created(c, "tugas berhasil dibuat dan ditugaskan", created)
}

// Update PUT /tasks/:id
func (h *Controller) Update(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id tugas tidak valid", nil)
	}
	t, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "tugas tidak ditemukan", nil)
	}

	var req TaskRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	dueDate, err := parseTanggal(req.DueDate)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "format tanggal tidak valid (YYYY-MM-DD)", nil)
	}

	t.Title = req.Title
	t.Description = req.Description
	t.AssignedTo = req.AssignedTo
	t.DueDate = dueDate
	t.Priority = req.Priority
	if err := h.repo.Update(t); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui tugas", nil)
	}
	return utils.OK(c, "tugas berhasil diperbarui", t)
}

// UpdateStatus PATCH /tasks/:id/status — dipakai karyawan menandai tugas
// miliknya "proses"/"selesai", ATAU admin/super_admin mengubah status
// tugas siapa pun.
func (h *Controller) UpdateStatus(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id tugas tidak valid", nil)
	}
	t, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "tugas tidak ditemukan", nil)
	}

	roleName, _ := c.Locals(constant.CtxRoleName).(string)
	userID, _ := c.Locals(constant.CtxUserID).(uint)
	if roleName == constant.RoleKaryawan && t.AssignedTo != userID {
		return utils.Fail(c, fiber.StatusForbidden, "tugas ini bukan milik Anda", nil)
	}

	var req UpdateStatusRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	t.Status = req.Status
	if req.Status == "selesai" {
		now := time.Now()
		t.CompletedAt = &now
	} else {
		t.CompletedAt = nil
	}
	if err := h.repo.Update(t); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui status tugas", nil)
	}
	return utils.OK(c, "status tugas berhasil diperbarui", t)
}

// Delete DELETE /tasks/:id — HANYA super_admin & admin.
func (h *Controller) Delete(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id tugas tidak valid", nil)
	}
	if _, err := h.repo.FindByID(id); err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "tugas tidak ditemukan", nil)
	}
	if err := h.repo.Delete(id); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menghapus tugas", nil)
	}
	return utils.OK(c, "tugas berhasil dihapus", nil)
}

func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/tasks", middleware.JWTAuth(h.jwtSvc))

	view := middleware.RequirePermission(h.roleRepo, Module, constant.ActionView)
	tambah := middleware.RequirePermission(h.roleRepo, Module, constant.ActionTambah)
	edit := middleware.RequirePermission(h.roleRepo, Module, constant.ActionEdit)
	onlyStaff := middleware.RequireRole(constant.RoleSuperAdmin, constant.RoleAdmin)

	g.Get("/summary", view, h.Summary)
	g.Get("/", view, h.List)
	g.Get("/:id", view, h.Detail)
	g.Post("/", tambah, onlyStaff, h.Create)
	g.Put("/:id", edit, onlyStaff, h.Update)
	g.Patch("/:id/status", edit, h.UpdateStatus)
	g.Delete("/:id", onlyStaff, edit, h.Delete)
}
