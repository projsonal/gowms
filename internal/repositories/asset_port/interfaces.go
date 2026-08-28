package asset_port

import "github.com/projsonal/gowms/internal/model"

type Repository interface {
	ListByAsset(assetID uint) ([]model.AssetPort, error)

	FindPort(assetID uint, portNumber int) (*model.AssetPort, error)

	Upsert(p *model.AssetPort) error

	Clear(assetID uint, portNumber int) error

	CountTerisi(assetID uint) (int64, error)
}
