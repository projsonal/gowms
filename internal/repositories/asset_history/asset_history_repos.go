package asset_history

import (
	"gorm.io/gorm"

	"github.com/projsonal/gowms/internal/model"
)

type Repository interface {
	Log(h *model.AssetHistory) error

	ListByAsset(assetID uint, limit int) ([]model.AssetHistory, error)
}

type repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Log(h *model.AssetHistory) error {
	return r.db.Create(h).Error
}

func (r *repository) ListByAsset(assetID uint, limit int) ([]model.AssetHistory, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var rows []model.AssetHistory
	err := r.db.Where("asset_id = ?", assetID).
		Order("created_at DESC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}
