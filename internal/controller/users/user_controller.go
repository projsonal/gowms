package users

import (
	"errors"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/projsonal/gowms/internal/middleware"
	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/constant"
	cons "github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

func (h *Controller) roleNameMap() map[uint]string {
	roles, err := h.roleRepo.FindAll()
	m := make(map[uint]string, len(roles))
	if err != nil {
		return m
	}
	for _, r := range roles {
		m[r.ID] = r.Name
	}
	return m
}

func (h *Controller) toResponse(u *model.User, roleName string, isOnline bool) Response {
	return Response{
		ID: u.ID, Username: u.Username, Email: u.Email, FullName: u.FullName,
		PhoneNumber: u.PhoneNumber, AvatarURL: u.AvatarURL(),
		RoleID: u.RoleID, RoleName: roleName, IsActive: u.IsActive, IsOnline: isOnline,
		Is2FAEnabled: u.Is2FAEnabled,
		LastLoginAt:  u.LastLoginAt,
	}
}

func (h *Controller) toResponseSingle(u *model.User) Response {
	roleName := ""
	if r, err := h.roleRepo.FindByID(u.RoleID); err == nil {
		roleName = r.Name
	}
	online, _ := h.authRepo.OnlineUserIDs([]uint{u.ID})
	return h.toResponse(u, roleName, online[u.ID])
}

func (h *Controller) List(c *fiber.Ctx) error {
	p := utils.PaginationFromContext(c)
	list, total, err := h.userRepo.List(p)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, cons.ErrGagalMengambilDaftarUser, nil)
	}

	roleNames := h.roleNameMap()
	userIDs := make([]uint, 0, len(list))
	for _, u := range list {
		userIDs = append(userIDs, u.ID)
	}
	onlineMap, err := h.authRepo.OnlineUserIDs(userIDs)
	if err != nil {
		onlineMap = map[uint]bool{}
	}

	responses := make([]Response, 0, len(list))
	for _, u := range list {
		responses = append(responses, h.toResponse(&u, roleNames[u.RoleID], onlineMap[u.ID]))
	}
	return utils.OKWithMeta(c, cons.MsgDaftarUserBerhasil, responses, utils.BuildPaginationMeta(p, total))
}

func (h *Controller) Detail(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, cons.ErrIDInvalid, nil)
	}
	u, err := h.userRepo.FindByID(uint(id))
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, cons.ErrUsersUserNotFound, nil)
	}
	return utils.OK(c, cons.MsgDetailUserBerhasil, h.toResponseSingle(u))
}

func (h *Controller) Create(c *fiber.Ctx) error {
	var req CreateUserRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}

	if _, err := h.userRepo.FindByUsername(req.Username); err == nil {
		return utils.Fail(c, fiber.StatusConflict, cons.ErrUsernameDuplikat, nil)
	}

	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, cons.ErrHashPasswordFail, nil)
	}

	u := &model.User{
		Username: req.Username, Email: req.Email, PasswordHash: hashed,
		FullName: req.FullName, PhoneNumber: req.PhoneNumber, RoleID: req.RoleID, IsActive: true,
	}
	if err := h.userRepo.Create(u); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, cons.ErrGagalMembuatUser, nil)
	}
	return utils.Created(c, cons.MsgUserBerhasilDibuat, h.toResponseSingle(u))
}

func (h *Controller) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, cons.ErrIDInvalid, nil)
	}
	u, err := h.userRepo.FindByID(uint(id))
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, cons.ErrUsersUserNotFound, nil)
	}

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, cons.ErrPayloadInvalid, nil)
	}

	if req.Email != "" {
		u.Email = req.Email
	}
	if req.FullName != "" {
		u.FullName = req.FullName
	}
	if req.PhoneNumber != nil {
		u.PhoneNumber = *req.PhoneNumber
	}
	if req.RoleID != 0 {
		u.RoleID = req.RoleID
	}
	if req.IsActive != nil {
		u.IsActive = *req.IsActive
	}

	if err := h.userRepo.Update(u); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, cons.ErrGagalMemperbaruiUser, nil)
	}
	return utils.OK(c, cons.MsgUserBerhasilDiubah, h.toResponseSingle(u))
}

func (h *Controller) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, cons.ErrIDInvalid, nil)
	}
	u, err := h.userRepo.FindByID(uint(id))
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, cons.ErrUsersUserNotFound, nil)
	}

	requesterID, _ := c.Locals(cons.CtxUserID).(uint)
	if requesterID == u.ID {
		return utils.Fail(c, fiber.StatusForbidden, "tidak bisa menghapus akun Anda sendiri", nil)
	}

	u.IsActive = false
	if err := h.userRepo.Update(u); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, cons.ErrGagalMemperbaruiUser, nil)
	}
	return utils.OK(c, "user berhasil dinonaktifkan", h.toResponseSingle(u))
}

func (h *Controller) currentUser(c *fiber.Ctx) (*model.User, error) {
	userID, _ := c.Locals(cons.CtxUserID).(uint)
	return h.userRepo.FindByID(userID)
}

func (h *Controller) ChangePassword(c *fiber.Ctx) error {
	var req ChangePasswordRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}

	if err := h.humanCheckSvc.Verify(req.HumanCheckToken); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "verifikasi gagal: "+err.Error(), nil)
	}

	u, err := h.currentUser(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, cons.ErrUsersUserNotFound, nil)
	}
	if !utils.ComparePassword(u.PasswordHash, req.OldPassword) {
		return utils.Fail(c, fiber.StatusBadRequest, cons.ErrPasswordLamaSalah, nil)
	}

	newHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, cons.ErrHashPasswordFail, nil)
	}
	u.PasswordHash = newHash

	if err := h.userRepo.Update(u); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, cons.ErrGagalMengubahPassword, nil)
	}
	return utils.OK(c, cons.MsgPasswordBerhasilUbah, nil)
}

type UpdateMeRequest struct {
	Username    string `json:"username" validate:"omitempty,min=4,max=50"`
	Email       string `json:"email" validate:"omitempty,email"`
	FullName    string `json:"full_name"`
	PhoneNumber string `json:"phone_number"`
}

func (h *Controller) UpdateMe(c *fiber.Ctx) error {
	u, err := h.currentUser(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, cons.ErrUsersUserNotFound, nil)
	}

	var req UpdateMeRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, cons.ErrPayloadInvalid, nil)
	}
	if errs := utils.Validate(req); errs != nil {
		return utils.Fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}

	if req.Username != "" && req.Username != u.Username {
		if existing, err := h.userRepo.FindByUsername(req.Username); err == nil && existing.ID != u.ID {
			return utils.Fail(c, fiber.StatusConflict, "username sudah dipakai user lain", nil)
		}
		u.Username = req.Username
	}
	if req.Email != "" {
		u.Email = req.Email
	}
	if req.FullName != "" {
		u.FullName = req.FullName
	}

	u.PhoneNumber = req.PhoneNumber

	if err := h.userRepo.Update(u); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, cons.ErrGagalMemperbaruiUser, nil)
	}
	return utils.OK(c, "profil berhasil diperbarui", h.toResponseSingle(u))
}

func (h *Controller) UploadAvatar(c *fiber.Ctx) error {
	u, err := h.currentUser(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, cons.ErrUsersUserNotFound, nil)
	}

	fh, err := c.FormFile("avatar")
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "file avatar tidak ditemukan (field: avatar)", nil)
	}
	const maxAvatarSize = 2 * 1024 * 1024
	if fh.Size > maxAvatarSize {
		return utils.Fail(c, fiber.StatusBadRequest, "ukuran file maksimal 2MB", nil)
	}
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	contentType := map[string]string{
		".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".png": "image/png",
	}[ext]
	if contentType == "" {
		return utils.Fail(c, fiber.StatusBadRequest, "format file harus jpg, jpeg, atau png", nil)
	}

	file, err := fh.Open()
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membaca file", nil)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membaca file", nil)
	}

	u.AvatarData = data
	u.AvatarContentType = contentType
	if err := h.userRepo.Update(u); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, cons.ErrGagalMemperbaruiUser, nil)
	}
	return utils.OK(c, "foto profil berhasil diperbarui", h.toResponseSingle(u))
}

func (h *Controller) DeleteAvatar(c *fiber.Ctx) error {
	u, err := h.currentUser(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, cons.ErrUsersUserNotFound, nil)
	}
	u.AvatarData = nil
	u.AvatarContentType = ""
	if err := h.userRepo.Update(u); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, cons.ErrGagalMemperbaruiUser, nil)
	}
	return utils.OK(c, "foto profil berhasil dihapus", h.toResponseSingle(u))
}

func (h *Controller) ServeAvatar(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, cons.ErrIDInvalid, nil)
	}
	u, err := h.userRepo.FindByID(uint(id))
	if err != nil || len(u.AvatarData) == 0 {
		return utils.Fail(c, fiber.StatusNotFound, "foto profil tidak ditemukan", nil)
	}
	c.Set("Content-Type", u.AvatarContentType)
	c.Set("Cache-Control", "private, max-age=86400")
	return c.Send(u.AvatarData)
}

func (h *Controller) UserSessions(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, cons.ErrIDInvalid, nil)
	}
	if _, err := h.userRepo.FindByID(uint(id)); err != nil {
		return utils.Fail(c, fiber.StatusNotFound, cons.ErrUsersUserNotFound, nil)
	}

	sessions, err := h.authRepo.ListActiveSessions(uint(id))
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memuat daftar sesi", nil)
	}

	result := make([]SessionResponse, 0, len(sessions))
	for _, s := range sessions {
		result = append(result, SessionResponse{
			ID: s.ID, Browser: s.Browser, BrowserVersion: s.BrowserVersion,
			OS: s.OS, OSVersion: s.OSVersion, DeviceType: s.DeviceType,
			IPAddress: s.IPAddress, Location: s.Location,
			CreatedAt: s.CreatedAt.Format(time.RFC3339),
		})
	}
	return utils.OK(c, "berhasil memuat daftar sesi aktif", SessionListResponse{Sessions: result})
}

func (h *Controller) RevokeUserSession(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, cons.ErrIDInvalid, nil)
	}
	sessionID, err := strconv.ParseUint(c.Params("sessionId"), 10, 64)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id sesi tidak valid", nil)
	}
	if _, err := h.userRepo.FindByID(uint(id)); err != nil {
		return utils.Fail(c, fiber.StatusNotFound, cons.ErrUsersUserNotFound, nil)
	}

	if err := h.authRepo.RevokeSession(uint(id), uint(sessionID)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.Fail(c, fiber.StatusNotFound, "sesi tidak ditemukan", nil)
		}
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mencabut sesi", nil)
	}
	return utils.OK(c, "sesi berhasil dicabut, perangkat tsb otomatis ter-logout", nil)
}

func (h *Controller) RegisterRoutes(router fiber.Router) {
	authed := router.Group("/users", middleware.JWTAuth(h.jwtSvc))
	authed.Patch("/me/password", h.ChangePassword)
	authed.Patch("/me", h.UpdateMe)
	authed.Post("/me/avatar", h.UploadAvatar)
	authed.Delete("/me/avatar", h.DeleteAvatar)

	authed.Get("/:id/avatar", h.ServeAvatar)

	g := authed.Group("", middleware.RequireRole(cons.RoleSuperAdmin, constant.RoleAdmin))
	g.Get("/", h.List)
	g.Get("/:id", h.Detail)
	g.Post("/", middleware.RequireRole(cons.RoleSuperAdmin), h.Create)
	g.Put("/:id", h.Update)

	g.Delete("/:id", middleware.RequireRole(cons.RoleSuperAdmin), h.Delete)

	g.Get("/:id/sessions", middleware.RequireRole(cons.RoleSuperAdmin), h.UserSessions)
	g.Delete("/:id/sessions/:sessionId", middleware.RequireRole(cons.RoleSuperAdmin), h.RevokeUserSession)
}
