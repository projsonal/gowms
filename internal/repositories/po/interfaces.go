package purchase_order

import (
	"gorm.io/gorm"

	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/utils"
)

type Filter struct {
	Status     string
	SupplierID uint
}

type Repository interface {
	List(p utils.PaginationParams, f Filter) ([]model.PurchaseOrder, int64, error)
	FindByID(id uint) (*model.PurchaseOrder, error)
	FindByNomor(nomor string) (*model.PurchaseOrder, error)
	Create(po *model.PurchaseOrder) error
	Update(po *model.PurchaseOrder, items []model.PurchaseOrderItem) error
	Delete(id uint) error

	Ajukan(id uint, userID uint) (*model.PurchaseOrder, error)
	SetujuiTolak(id uint, userID uint, setuju bool, catatan string) (*model.PurchaseOrder, error)
	Batalkan(id uint) (*model.PurchaseOrder, error)
	SetProtected(id uint, protected bool) (*model.PurchaseOrder, error)

	TambahPenerimaan(tx *gorm.DB, purchaseOrderID, barangID uint, qty int) error

	CountByStatus(status string) (int64, error)
}
