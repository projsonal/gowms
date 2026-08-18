package notification

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/internal/middleware"
	"github.com/projsonal/gowms/internal/model"
	notificationRepo "github.com/projsonal/gowms/internal/repositories/notification"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

type Controller struct {
	repo   notificationRepo.Repository
	jwtSvc *utils.JWTService
}

func New(repo notificationRepo.Repository, jwtSvc *utils.JWTService) *Controller {
	return &Controller{repo: repo, jwtSvc: jwtSvc}
}

func currentUser(c *fiber.Ctx) (uint, string) {
	userID, _ := c.Locals(constant.CtxUserID).(uint)
	role, _ := c.Locals(constant.CtxRoleName).(string)
	return userID, role
}

// List GET /notifications?page=&limit= — notifikasi milik user yang login
// (personal + broadcast relevan role-nya), terbaru dulu.
func (h *Controller) List(c *fiber.Ctx) error {
	userID, role := currentUser(c)
	p := utils.PaginationFromContext(c)
	rows, total, err := h.repo.List(userID, role, p)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil notifikasi", nil)
	}
	return utils.OKWithMeta(c, "notifikasi berhasil diambil", rows, utils.BuildPaginationMeta(p, total))
}

// UnreadCount GET /notifications/unread-count — dipoll berkala frontend
// untuk badge angka merah di ikon lonceng, TANPA perlu ambil seluruh daftar.
func (h *Controller) UnreadCount(c *fiber.Ctx) error {
	userID, role := currentUser(c)
	count, err := h.repo.UnreadCount(userID, role)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menghitung notifikasi belum dibaca", nil)
	}
	return utils.OK(c, "jumlah notifikasi belum dibaca berhasil diambil", fiber.Map{"unread_count": count})
}

// MarkRead PATCH /notifications/:id/read
func (h *Controller) MarkRead(c *fiber.Ctx) error {
	userID, _ := currentUser(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id tidak valid", nil)
	}
	if err := h.repo.MarkRead(uint(id), userID); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menandai notifikasi", nil)
	}
	return utils.OK(c, "notifikasi ditandai sudah dibaca", nil)
}

// MarkAllRead PATCH /notifications/read-all
func (h *Controller) MarkAllRead(c *fiber.Ctx) error {
	userID, role := currentUser(c)
	if err := h.repo.MarkAllRead(userID, role); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menandai semua notifikasi", nil)
	}
	return utils.OK(c, "semua notifikasi ditandai sudah dibaca", nil)
}

// Delete DELETE /notifications/:id — hapus notifikasi dari daftar user yang
// sedang login SAJA (lihat catatan panjang di model.NotificationDismissed:
// notifikasi broadcast dibagi ke banyak user lewat satu baris yang sama,
// jadi ini TIDAK menghapus baris Notification aslinya — pengguna lain yang
// jadi target broadcast yang sama tetap melihatnya seperti biasa).
func (h *Controller) Delete(c *fiber.Ctx) error {
	userID, _ := currentUser(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id tidak valid", nil)
	}
	if err := h.repo.Dismiss(uint(id), userID); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menghapus notifikasi", nil)
	}
	return utils.OK(c, "notifikasi dihapus dari daftar", nil)
}

func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/notifications", middleware.JWTAuth(h.jwtSvc))
	g.Get("/", h.List)
	g.Get("/unread-count", h.UnreadCount)
	g.Patch("/read-all", h.MarkAllRead)
	g.Patch("/:id/read", h.MarkRead)
	g.Delete("/:id", h.Delete)
}

// Helper diekspor supaya controller modul lain (barang_rusak, asset_gudang,
// maintenance, trash) bisa mengirim notifikasi tanpa saling import
// controller package satu sama lain (cuma butuh repo-nya).
func Notify(repo notificationRepo.Repository, notifType, title, message, linkHref string, userID *uint, roleTarget string) {
	_ = repo.Create(&model.Notification{
		UserID: userID, RoleTarget: roleTarget,
		Type: notifType, Title: title, Message: message, LinkHref: linkHref,
	})
}
