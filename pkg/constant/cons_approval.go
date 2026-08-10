package constant

// Status alur persetujuan data yang dibuat role admin (lihat model.Barang.
// ApprovalStatus & internal/controller/barang Approve()/Reject()).
// super_admin membuat data -> langsung ApprovalDisetujui.
// admin membuat data       -> ApprovalMenunggu sampai di-review super_admin.
const (
	ApprovalDisetujui = "disetujui"
	ApprovalMenunggu  = "menunggu"
	ApprovalDitolak   = "ditolak"
)
