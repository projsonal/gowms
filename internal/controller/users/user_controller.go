package users

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gostock/internal/middleware"
	"github.com/projsonal/gostock/internal/model"
	"github.com/projsonal/gostock/pkg/constant"
	cons "github.com/projsonal/gostock/pkg/constant"
	"github.com/projsonal/gostock/pkg/utils"
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

func (h *Controller) toResponse(u *model.User, roleName string) Response {
	return Response{
		ID: u.ID, Username: u.Username, Email: u.Email, FullName: u.FullName,
		RoleID: u.RoleID, RoleName: roleName, IsActive: u.IsActive, Is2FAEnabled: u.Is2FAEnabled,
	}
}

// toResponseSingle dipakai untuk endpoint yang mengembalikan 1 user saja
// (Detail, Create, Update) sehingga tetap boleh query role sekali per request.
func (h *Controller) toResponseSingle(u *model.User) Response {
	roleName := ""
	if r, err := h.roleRepo.FindByID(u.RoleID); err == nil {
		roleName = r.Name
	}
	return h.toResponse(u, roleName)
}

// List GET /api/v1/users — "Semua akun pengguna sistem".
func (h *Controller) List(c *fiber.Ctx) error {
	p := utils.PaginationFromContext(c)
	list, total, err := h.userRepo.List(p)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, cons.ErrGagalMengambilDaftarUser, nil)
	}

	roleNames := h.roleNameMap() // 1 query untuk semua role, bukan per-user
	responses := make([]Response, 0, len(list))
	for _, u := range list {
		responses = append(responses, h.toResponse(&u, roleNames[u.RoleID]))
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

func (h *Controller) currentUser(c *fiber.Ctx) (*model.User, error) {
	userID, _ := c.Locals(cons.CtxUserID).(uint)
	return h.userRepo.FindByID(userID)
}

func (h *Controller) RequestChangePasswordOTP(c *fiber.Ctx) error {
	var req RequestChangePasswordOTPRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}

	u, err := h.currentUser(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, constant.ErrUsersUserNotFound, nil)
	}
	if !utils.ComparePassword(u.PasswordHash, req.OldPassword) {
		return utils.Fail(c, fiber.StatusBadRequest, constant.ErrPasswordLamaSalah, nil)
	}
	if u.PhoneNumber == "" {
		return utils.Fail(c, fiber.StatusUnprocessableEntity, constant.ErrTanpaNomorHP, nil)
	}

	code, token, err := h.waOTPSvc.Generate()
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, cons.ErrGagalBuatOTP, nil)
	}
	if err := h.waSender.SendOTP(u.PhoneNumber, code); err != nil {
		log.Printf("users: gagal mengirim OTP ganti password lewat WhatsApp ke user %d: %v", u.ID, err)
		return utils.Fail(c, fiber.StatusBadGateway, cons.ErrGagalKirimOTPWa, nil)
	}

	return utils.OK(c, cons.MsgOTPPasswordTerkirim, RequestChangePasswordOTPResponse{
		OTPToken:  token,
		ExpiresIn: int(h.waOTPTTL.Seconds()),
	})
}

func (h *Controller) ConfirmChangePassword(c *fiber.Ctx) error {
	var req ConfirmChangePasswordRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}

	u, err := h.currentUser(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, cons.ErrUsersUserNotFound, nil)
	}
	if err := h.waOTPSvc.Verify(req.OTPToken, req.OTPCode); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, cons.ErrOTPSalahKedaluwarsa, nil)
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

func (h *Controller) RegisterRoutes(router fiber.Router) {
	authed := router.Group("/users", middleware.JWTAuth(h.jwtSvc))
	authed.Patch("/me/password/request-otp", h.RequestChangePasswordOTP)
	authed.Patch("/me/password/confirm", h.ConfirmChangePassword)

	g := authed.Group("", middleware.RequireRole(cons.RoleSuperAdmin, constant.RoleAdmin))
	g.Get("/", h.List)
	g.Get("/:id", h.Detail)
	g.Post("/", middleware.RequireRole(cons.RoleSuperAdmin), h.Create)
	g.Put("/:id", h.Update)
}
