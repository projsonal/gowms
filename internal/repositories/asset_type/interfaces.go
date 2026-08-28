package assettype

import "github.com/projsonal/gowms/internal/model"

type Repository interface {
	List() ([]model.AssetType, error)
	FindByID(id uint) (*model.AssetType, error)
	FindByKode(kode string) (*model.AssetType, error)
	Create(t *model.AssetType) error
	Update(t *model.AssetType) error
	Delete(id uint) error

	CountAssetsUsing(kode string) (int64, error)
}
