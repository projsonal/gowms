// Package users mengakses tabel users — dipakai modul Auth (login, cek
// kredensial) dan modul Manajemen User (CRUD akun).
package users

import (
	"time"

	"github.com/projsonal/gostock/internal/model"
	"github.com/projsonal/gostock/pkg/utils"
)

// Repository mendefinisikan seluruh operasi data untuk User.
type Repository interface {
	FindByUsername(username string) (*model.User, error)
	FindByID(id uint) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	Create(u *model.User) error
	Update(u *model.User) error
	UpdateTOTPSecret(userID uint, secret string, enabled bool) error
	UpdateLastLogin(userID uint) error
	List(p utils.PaginationParams) ([]model.User, int64, error)

	// RegisterFailedLogin menaikkan counter gagal login; kalau sudah
	// melewati ambang batas, kunci akun sementara (LockedUntil).
	RegisterFailedLogin(userID uint, maxAttempts int, lockDuration time.Duration) error
	// ResetFailedLogin dipanggil setelah login berhasil (reset counter & unlock).
	ResetFailedLogin(userID uint) error
}
