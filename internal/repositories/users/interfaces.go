package users

import (
	"time"

	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/utils"
)

type Repository interface {
	FindByUsername(username string) (*model.User, error)
	FindByID(id uint) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	Create(u *model.User) error
	Update(u *model.User) error
	Delete(id uint) error
	UpdateTOTPSecret(userID uint, secret string, enabled bool) error
	UpdateLastLogin(userID uint) error
	List(p utils.PaginationParams) ([]model.User, int64, error)

	RegisterFailedLogin(userID uint, maxAttempts int, lockDuration time.Duration) error

	ResetFailedLogin(userID uint) error
}
