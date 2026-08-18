package asset_port

import (
	"errors"

	"gorm.io/gorm"

	"github.com/projsonal/gowms/internal/model"
)

type repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) ListByAsset(assetID uint) ([]model.AssetPort, error) {
	var ports []model.AssetPort
	err := r.db.Preload("ChildAsset").
		Where("asset_id = ?", assetID).
		Order("port_number ASC").
		Find(&ports).Error
	return ports, err
}

func (r *repository) FindPort(assetID uint, portNumber int) (*model.AssetPort, error) {
	var p model.AssetPort
	err := r.db.Preload("ChildAsset").
		Where("asset_id = ? AND port_number = ?", assetID, portNumber).
		First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *repository) Upsert(p *model.AssetPort) error {
	existing, err := r.FindPort(p.AssetID, p.PortNumber)
	if err != nil {
		return err
	}
	if existing == nil {
		return r.db.Create(p).Error
	}
	p.ID = existing.ID
	return r.db.Model(&model.AssetPort{}).Where("id = ?", existing.ID).
		Updates(map[string]any{
			"status":         p.Status,
			"child_asset_id": p.ChildAssetID,
			"customer_name":  p.CustomerName,
			"customer_phone": p.CustomerPhone,
			"keterangan":     p.Keterangan,
		}).Error
}

func (r *repository) Clear(assetID uint, portNumber int) error {
	return r.db.Model(&model.AssetPort{}).
		Where("asset_id = ? AND port_number = ?", assetID, portNumber).
		Updates(map[string]any{
			"status":         "kosong",
			"child_asset_id": nil,
			"customer_name":  "",
			"customer_phone": "",
			"keterangan":     "",
		}).Error
}

func (r *repository) CountTerisi(assetID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.AssetPort{}).
		Where("asset_id = ? AND status = ?", assetID, "terisi").
		Count(&count).Error
	return count, err
}
