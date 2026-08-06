package constant

const (
	ErrUsersUserNotFound   = "user tidak ditemukan"
	ErrUsernameDuplikat    = "username sudah digunakan"
	ErrPasswordLamaSalah   = "password lama salah"
	ErrTanpaNomorHP        = "akun ini belum punya nomor HP terdaftar, hubungi administrator untuk melengkapinya"
	ErrOTPSalahKedaluwarsa = "kode OTP salah, sudah kedaluwarsa, atau sudah pernah dipakai"
	ErrGagalKirimOTPWa     = "gagal mengirim kode OTP lewat WhatsApp, coba lagi"
	ErrGagalBuatOTP        = "gagal membuat kode OTP"

	ErrGagalMengambilDaftarUser = "gagal mengambil daftar user"
	ErrGagalMembuatUser         = "gagal membuat user"
	ErrGagalMemperbaruiUser     = "gagal memperbarui user"
	ErrGagalMengubahPassword    = "gagal mengubah password"

	MsgDaftarUserBerhasil   = "daftar user berhasil diambil"
	MsgDetailUserBerhasil   = "detail user berhasil diambil"
	MsgUserBerhasilDibuat   = "user berhasil dibuat"
	MsgUserBerhasilDiubah   = "user berhasil diperbarui"
	MsgOTPPasswordTerkirim  = "kode OTP ganti password telah dikirim lewat WhatsApp"
	MsgPasswordBerhasilUbah = "password berhasil diubah"
)

const QueryIDEq = "id = ?"
