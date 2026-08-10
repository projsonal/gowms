package users

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/internal/middleware"
	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/constant"
	cons "github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

// roleNameMap mengambil SEMUA role dalam satu query lalu memetakannya
// {role_id: role_name}, dipakai List() untuk menghindari N+1 query
// (sebelumnya toResponse memanggil roleRepo.FindByID per user di dalam loop).
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
		PhoneNumber: u.PhoneNumber, AvatarURL: u.AvatarURL,
		RoleID: u.RoleID, RoleName: roleName, IsActive: u.IsActive, IsOnline: isOnline,
		Is2FAEnabled: u.Is2FAEnabled,
		LastLoginAt:  u.LastLoginAt,
	}
}

// toResponseSingle dipakai untuk endpoint yang mengembalikan 1 user saja
// (Detail, Create, Update) sehingga tetap boleh query role sekali per request.
func (h *Controller) toResponseSingle(u *model.User) Response {
	roleName := ""
	if r, err := h.roleRepo.FindByID(u.RoleID); err == nil {
		roleName = r.Name
	}
	online, _ := h.authRepo.OnlineUserIDs([]uint{u.ID})
	return h.toResponse(u, roleName, online[u.ID])
}

// List GET /api/v1/users — "Semua akun pengguna sistem".
func (h *Controller) List(c *fiber.Ctx) error {
	p := utils.PaginationFromContext(c)
	list, total, err := h.userRepo.List(p)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, cons.ErrGagalMengambilDaftarUser, nil)
	}

	roleNames := h.roleNameMap() // 1 query untuk semua role, bukan per-user
	userIDs := make([]uint, 0, len(list))
	for _, u := range list {
		userIDs = append(userIDs, u.ID)
	}
	onlineMap, err := h.authRepo.OnlineUserIDs(userIDs) // 1 query untuk status online semua user sekaligus
	if err != nil {
		onlineMap = map[uint]bool{} // gagal cek status online BUKAN alasan gagal load seluruh daftar user
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
		FullName: req.FullName, RoleID: req.RoleID, IsActive: true,
	}
	if err := h.userRepo.Create(u); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, cons.ErrGagalMembuatUser, nil)
	}
	return utils.Created(c, cons.MsgUserBerhasilDibuat, h.toResponseSingle(u))
}

// Update PUT /api/v1/users/:id
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

// Delete DELETE /api/v1/users/:id — HANYA super_admin (lihat RegisterRoutes).
//
// Ini SENGAJA menonaktifkan akun (is_active=false), BUKAN menghapus baris
// user dari database: banyak tabel lain (barang_masuk.diterima_oleh,
// barang_keluar.dikeluarkan_oleh, purchase_orders.diajukan_oleh, dst)
// mereferensikan user_id untuk jejak histori transaksi — hard delete akan
// merusak riwayat itu atau gagal karena foreign key. Akun nonaktif otomatis
// tidak bisa login lagi (lihat auth_controller.go: cek IsActive saat Login).
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

	if err := h.captchaSvc.Verify(req.CaptchaToken, req.CaptchaAnswer); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "verifikasi captcha gagal: "+err.Error(), nil)
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

// UpdateMeRequest — form Settings -> Profil "Simpan Perubahan". SENGAJA
// tidak punya field RoleID/IsActive: siapa pun (termasuk karyawan) boleh
// ubah profilnya sendiri, tapi role & status aktif akun HANYA boleh
// diubah lewat Manajemen User (Update/Delete di atas, khusus staff) —
// supaya user tidak bisa menaikkan role-nya sendiri.
type UpdateMeRequest struct {
	Email       string `json:"email" validate:"omitempty,email"`
	FullName    string `json:"full_name"`
	PhoneNumber string `json:"phone_number"`
	AvatarURL   string `json:"avatar_url"`
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

	if req.Email != "" {
		u.Email = req.Email
	}
	if req.FullName != "" {
		u.FullName = req.FullName
	}
	// PhoneNumber & AvatarURL boleh dikosongkan lagi oleh user (mis. hapus
	// foto profil), jadi TIDAK dicek `!= ""` seperti field lain di atas.
	u.PhoneNumber = req.PhoneNumber
	u.AvatarURL = req.AvatarURL

	if err := h.userRepo.Update(u); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, cons.ErrGagalMemperbaruiUser, nil)
	}
	return utils.OK(c, "profil berhasil diperbarui", h.toResponseSingle(u))
}

// UploadAvatar POST /users/me/avatar (multipart/form-data, field "avatar")
// — simpan foto profil di disk lokal (StorageConfig.Path/avatars/) dan
// langsung update AvatarURL milik user yang sedang login. Dibatasi 2MB &
// hanya jpg/jpeg/png supaya tidak disalahgunakan untuk upload file besar
// sembarangan.
func (h *Controller) UploadAvatar(c *fiber.Ctx) error {
	u, err := h.currentUser(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, cons.ErrUsersUserNotFound, nil)
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "file avatar tidak ditemukan (field: avatar)", nil)
	}
	const maxAvatarSize = 2 * 1024 * 1024 // 2MB
	if file.Size > maxAvatarSize {
		return utils.Fail(c, fiber.StatusBadRequest, "ukuran file maksimal 2MB", nil)
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return utils.Fail(c, fiber.StatusBadRequest, "format file harus jpg, jpeg, atau png", nil)
	}

	avatarDir := filepath.Join(h.storagePath, "avatars")
	if err := os.MkdirAll(avatarDir, 0o755); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menyiapkan folder upload", nil)
	}
	filename := fmt.Sprintf("user-%d-%d%s", u.ID, time.Now().UnixNano(), ext)
	if err := c.SaveFile(file, filepath.Join(avatarDir, filename)); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menyimpan file", nil)
	}

	u.AvatarURL = "/uploads/avatars/" + filename
	if err := h.userRepo.Update(u); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, cons.ErrGagalMemperbaruiUser, nil)
	}
	return utils.OK(c, "foto profil berhasil diperbarui", h.toResponseSingle(u))
}

func (h *Controller) RegisterRoutes(router fiber.Router) {
	authed := router.Group("/users", middleware.JWTAuth(h.jwtSvc))
	authed.Patch("/me/password", h.ChangePassword)
	authed.Patch("/me", h.UpdateMe)
	authed.Post("/me/avatar", h.UploadAvatar)

	g := authed.Group("", middleware.RequireRole(cons.RoleSuperAdmin, constant.RoleAdmin))
	g.Get("/", h.List)
	g.Get("/:id", h.Detail)
	g.Post("/", middleware.RequireRole(cons.RoleSuperAdmin), h.Create)
	g.Put("/:id", h.Update)
	// Delete (nonaktifkan akun) HANYA super_admin, sama seperti Create.
	g.Delete("/:id", middleware.RequireRole(cons.RoleSuperAdmin), h.Delete)
}
