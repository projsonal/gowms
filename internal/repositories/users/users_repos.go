package users

import (
	"time"

	"gorm.io/gorm"

	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/utils"
)

func (r *repository) FindByUsername(username string) (*model.User, error) {
	var u model.User
	if err := r.db.Where("username = ?", username).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repository) FindByID(id uint) (*model.User, error) {
	var u model.User
	if err := r.db.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repository) FindByEmail(email string) (*model.User, error) {
	var u model.User
	if err := r.db.Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repository) Create(u *model.User) error {
	return r.db.Create(u).Error
}

func (r *repository) Update(u *model.User) error {
	return r.db.Save(u).Error
}

func (r *repository) UpdateTOTPSecret(userID uint, secret string, enabled bool) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{"totp_secret": secret, "is_2fa_enabled": enabled}).Error
}

func (r *repository) UpdateLastLogin(userID uint) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).
		Update("last_login_at", gorm.Expr("now()")).Error
}

// List — dipakai layar "Manajemen User: Semua akun pengguna sistem".
func (r *repository) List(p utils.PaginationParams) ([]model.User, int64, error) {
	var list []model.User
	var total int64

	q := r.db.Model(&model.User{})
	if p.Search != "" {
		q = q.Where("username ILIKE ? OR full_name ILIKE ? OR email ILIKE ?",
			"%"+p.Search+"%", "%"+p.Search+"%", "%"+p.Search+"%")
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := p.Apply(q.Order("id desc")).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// RegisterFailedLogin menaikkan counter gagal login secara atomik. Kalau
// counter mencapai maxAttempts, akun dikunci selama lockDuration —
// mitigasi brute-force yang lolos dari captcha (mis. captcha di-bypass
// tapi password tetap ditebak berkali-kali).
func (r *repository) RegisterFailedLogin(userID uint, maxAttempts int, lockDuration time.Duration) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var u model.User
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&u, userID).Error; err != nil {
			return err
		}
		u.FailedLoginAttempts++
		updates := map[string]interface{}{"failed_login_attempts": u.FailedLoginAttempts}
		if u.FailedLoginAttempts >= maxAttempts {
			lockedUntil := time.Now().Add(lockDuration)
			updates["locked_until"] = lockedUntil
		}
		return tx.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
	})
}

// ResetFailedLogin dipanggil setelah login berhasil (reset counter & unlock).
func (r *repository) ResetFailedLogin(userID uint) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{"failed_login_attempts": 0, "locked_until": nil}).Error
}
