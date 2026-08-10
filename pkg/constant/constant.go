package constant

const (
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
	RoleKaryawan   = "karyawan"
)

const (
	ActionView           = "view"
	ActionTambah         = "tambah"
	ActionEdit           = "edit"
	ActionApprovalReject = "approval_reject"
	ActionPrint          = "print"
	ActionAssignDelegasi = "assign_delegasi"
)

// Nama modul untuk RBAC, sesuai menu sidebar pada mockup. Modul selain
// user & role akan dipakai saat modul terkait (barang, PO, dst) dikerjakan.
const (
	ModuleDashboard       = "dashboard"
	ModuleManajemenUser   = "manajemen_user"
	ModuleSettings        = "settings"
	ModuleKelolaBarang    = "kelola_barang"
	ModulePurchaseOrder   = "purchase_order"
	ModuleSupplier        = "supplier"
	ModuleBarangMasuk     = "barang_masuk"
	ModuleBarangKeluar    = "barang_keluar"
	ModuleManajemenGudang = "manajemen_gudang"
	ModuleStockOpname     = "stock_opname"
	ModulePengiriman      = "pengiriman"
	ModuleLaporan         = "laporan"
	ModuleCOD             = "cod"
	ModuleTaskManagement  = "tasks"
)

// Fiber Locals context key, di-set oleh middleware auth setelah JWT valid.
const (
	CtxUserID   = "ctx_user_id"
	CtxRoleID   = "ctx_role_id"
	CtxRoleName = "ctx_role_name"
	// CtxSessionID: ID baris refresh_tokens yang menerbitkan access token
	// yang sedang dipakai — lihat pkg/utils/jwt.go JWTClaims.SessionID &
	// internal/controller/auth ListSessions/RevokeSession.
	CtxSessionID = "ctx_session_id"
)

// Status akun & sesi.
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
)

const (
	ErrPayloadInvalid   = "payload tidak valid"
	ErrValidationFailed = " validasi gagal"
	ErrIDInvalid        = "id tidak valid"
	ErrHashPasswordFail = "gagal mengenskripsi password"
	ErrInternalGagal    = "terjadi kesalahan pada server, coba lagi nanti"
)

const (
	QueryPage   = "page"
	QueryLimit  = "limit"
	QuerySearch = "search"
	QuerySort   = "sort"
)
