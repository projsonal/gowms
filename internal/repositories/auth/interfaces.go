package auth

import "github.com/projsonal/gowms/internal/model"

type Repository interface {
	SaveRefreshToken(t *model.RefreshToken) error
	FindActiveRefreshToken(userID uint, tokenHash string) (*model.RefreshToken, error)
	RevokeAllUserTokens(userID uint) error

	ListActiveSessions(userID uint) ([]model.RefreshToken, error)
	RevokeSession(userID, sessionID uint) error

	// OnlineUserIDs — dari daftar userID, kembalikan yang mana saja
	// SEDANG punya sesi aktif (refresh token belum dicabut & belum
	// kedaluwarsa). Dipakai kolom "Status" (Aktif/Nonaktif) di Manajemen
	// User supaya mencerminkan status login SAAT INI, bukan flag
	// aktif/nonaktif akun (yang itu field IsActive terpisah di model.User).
	OnlineUserIDs(userIDs []uint) (map[uint]bool, error)
}
