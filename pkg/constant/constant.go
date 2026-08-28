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

	ModuleAsetGudang = "aset_gudang"

	ModuleBarangRusak = "barang_rusak"
)

const (
	JenisAsetTiang        = "tiang"
	JenisAsetODC          = "odc"
	JenisAsetOLT          = "olt"
	JenisAsetONT          = "ont"
	JenisAsetODP          = "odp"
	JenisAsetModem        = "modem"
	JenisAsetTransportasi = "transportasi"
)

const (
	TipeGudangPusat  = "pusat"
	TipeGudangCabang = "cabang"
)

const (
	StatusPengecekan = "pengecekan"
	StatusRetur      = "retur"
	StatusRusak      = "rusak"
)

const (
	CtxUserID   = "ctx_user_id"
	CtxRoleID   = "ctx_role_id"
	CtxRoleName = "ctx_role_name"

	CtxSessionID = "ctx_session_id"
)

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
