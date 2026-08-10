package auth

import (
	"time"

	"github.com/projsonal/gowms/internal/model"
	"gorm.io/gorm"
)

func (r *repository) SaveRefreshToken(t *model.RefreshToken) error {
	return r.db.Create(t).Error
}

func (r *repository) FindActiveRefreshToken(userID uint, tokenHash string) (*model.RefreshToken, error) {
	var t model.RefreshToken
	err := r.db.
		Where("user_id = ? AND token_hash = ? AND revoked = false", userID, tokenHash).
		First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repository) RevokeAllUserTokens(userID uint) error {
	return r.db.Model(&model.RefreshToken{}).
		Where("user_id = ?", userID).
		Update("revoked", true).Error
}

func (r *repository) ListActiveSessions(userID uint) ([]model.RefreshToken, error) {
	var sessions []model.RefreshToken
	err := r.db.
		Where("user_id = ? AND revoked = false", userID).
		Order("created_at DESC").
		Find(&sessions).Error
	return sessions, err
}

func (r *repository) OnlineUserIDs(userIDs []uint) (map[uint]bool, error) {
	result := make(map[uint]bool, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}
	var onlineIDs []uint
	err := r.db.Model(&model.RefreshToken{}).
		Distinct("user_id").
		Where("user_id IN ? AND revoked = false AND expires_at > ?", userIDs, time.Now()).
		Pluck("user_id", &onlineIDs).Error
	if err != nil {
		return nil, err
	}
	for _, id := range onlineIDs {
		result[id] = true
	}
	return result, nil
}

func (r *repository) RevokeSession(userID, sessionID uint) error {
	result := r.db.Model(&model.RefreshToken{}).
		Where("id = ? AND user_id = ?", sessionID, userID).
		Update("revoked", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
