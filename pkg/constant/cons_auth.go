package constant

// digunakan untuk pesan auth
const (
	ErrSessionExpired = "sesi login sudah kedaluwarsa, silakan login ulang"
	ErrUserNotFound   = "user tidak ditemukan"
	ErrInvalidToken   = "token tidak valid"

	MsgLoginSuccess      = "login berhasil"
	MsgOTPVerified       = "Verifikasi OTP Berhasil"
	MsgSessionCreated    = "berhasil membaca sesi aktif"
	MsgSessionRevoked    = "sesi berhasil dicabut"
	MsgSessionListLoaded = "berhasil memuat daftar sesi aktif"
)
