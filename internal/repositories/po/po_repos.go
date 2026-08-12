package purchase_order

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

const queryPurchaseOrderID = "purchase_order_id = ?"

func applyFilter(q *gorm.DB, f Filter) *gorm.DB {
	if f.Status != "" {
		q = q.Where(constant.QueryStatusEq, f.Status)
	}
	if f.SupplierID != 0 {
		q = q.Where("supplier_id = ?", f.SupplierID)
	}
	return q
}

func (r *repository) List(p utils.PaginationParams, f Filter) ([]model.PurchaseOrder, int64, error) {
	var list []model.PurchaseOrder
	var total int64

	q := applyFilter(r.db.Model(&model.PurchaseOrder{}), f)
	if p.Search != "" {
		q = q.Where("nomor_po ILIKE ?", "%"+p.Search+"%")
	}
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := p.Apply(q.Session(&gorm.Session{}).Preload("Supplier").Order("id desc")).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *repository) FindByID(id uint) (*model.PurchaseOrder, error) {
	var po model.PurchaseOrder
	if err := r.db.Preload("Supplier").Preload("Items").Preload("Items.Barang").First(&po, id).Error; err != nil {
		return nil, err
	}
	return &po, nil
}

func (r *repository) FindByNomor(nomor string) (*model.PurchaseOrder, error) {
	var po model.PurchaseOrder
	if err := r.db.Where("nomor_po = ?", nomor).First(&po).Error; err != nil {
		return nil, err
	}
	return &po, nil
}

func hitungTotal(items []model.PurchaseOrderItem) int64 {
	var total int64
	for i := range items {
		items[i].Subtotal = int64(items[i].QtyPesan) * items[i].HargaSatuan
		total += items[i].Subtotal
	}
	return total
}

func (r *repository) Create(po *model.PurchaseOrder) error {
	po.TotalEstimasi = hitungTotal(po.Items)
	return r.db.Create(po).Error
}

// Update mengganti header + seluruh item PO sekaligus (full replace),
// HANYA diizinkan selama status masih "draft" (dicek di layer controller
// sebelum memanggil ini). Item lama dihapus lalu diganti item baru supaya
// tidak perlu logika diff tambah/ubah/hapus per baris.
func (r *repository) Update(po *model.PurchaseOrder, items []model.PurchaseOrderItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where(queryPurchaseOrderID, po.ID).Delete(&model.PurchaseOrderItem{}).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].ID = 0
			items[i].PurchaseOrderID = po.ID
		}
		po.TotalEstimasi = hitungTotal(items)
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		return tx.Save(po).Error
	})
}

func (r *repository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where(queryPurchaseOrderID, id).Delete(&model.PurchaseOrderItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.PurchaseOrder{}, id).Error
	})
}

func (r *repository) Ajukan(id uint, userID uint) (*model.PurchaseOrder, error) {
	now := time.Now()
	updates := map[string]interface{}{
		"status":        constant.StatusPODiajukan,
		"diajukan_oleh": userID,
		"diajukan_at":   now,
	}
	if err := r.db.Model(&model.PurchaseOrder{}).Where("id = ? AND status = ?", id, constant.StatusPODraft).
		Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

func (r *repository) SetujuiTolak(id uint, userID uint, setuju bool, catatan string) (*model.PurchaseOrder, error) {
	status := constant.StatusPODitolak
	if setuju {
		status = constant.StatusPODisetujui
	}
	now := time.Now()
	updates := map[string]interface{}{
		"status":           status,
		"disetujui_oleh":   userID,
		"disetujui_at":     now,
		"catatan_approval": catatan,
	}
	res := r.db.Model(&model.PurchaseOrder{}).Where("id = ? AND status = ?", id, constant.StatusPODiajukan).Updates(updates)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, errors.New(constant.ErrPOTidakDiajukan)
	}
	return r.FindByID(id)
}

func (r *repository) Batalkan(id uint) (*model.PurchaseOrder, error) {
	res := r.db.Model(&model.PurchaseOrder{}).
		Where("id = ? AND status IN ?", id, []string{constant.StatusPODraft, constant.StatusPODiajukan}).
		Update("status", constant.StatusPODibatalkan)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, errors.New(constant.ErrPOBukanDraft)
	}
	return r.FindByID(id)
}

// SetProtected — aksi "Protect" di action bar tabel (khusus super_admin,
// lihat RegisterRoutes). Sama pola dengan Gudang/Barang/Supplier.
func (r *repository) SetProtected(id uint, protected bool) (*model.PurchaseOrder, error) {
	if err := r.db.Model(&model.PurchaseOrder{}).Where("id = ?", id).
		Update("is_protected", protected).Error; err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

// TambahPenerimaan lihat dokumentasi di interfaces.go — dipanggil dari
// dalam transaksi Barang Masuk.
func (r *repository) TambahPenerimaan(tx *gorm.DB, purchaseOrderID, barangID uint, qty int) error {
	if err := tx.Model(&model.PurchaseOrderItem{}).
		Where("purchase_order_id = ? AND barang_id = ?", purchaseOrderID, barangID).
		Update("qty_diterima", gorm.Expr("qty_diterima + ?", qty)).Error; err != nil {
		return err
	}

	var items []model.PurchaseOrderItem
	if err := tx.Where("purchase_order_id = ?", purchaseOrderID).Find(&items).Error; err != nil {
		return err
	}
	po := model.PurchaseOrder{Items: items}
	if po.IsFullyReceived() {
		if err := tx.Model(&model.PurchaseOrder{}).
			Where("id = ? AND status = ?", purchaseOrderID, constant.StatusPODisetujui).
			Update("status", constant.StatusPOSelesai).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *repository) CountByStatus(status string) (int64, error) {
	var count int64
	err := r.db.Model(&model.PurchaseOrder{}).Where(constant.QueryStatusEq, status).Count(&count).Error
	return count, err
}
