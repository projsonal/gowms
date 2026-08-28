package stock_opname

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/internal/repositories/barangstokgudang"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

func applyFilter(q *gorm.DB, f Filter) *gorm.DB {
	if f.Status != "" {
		q = q.Where(constant.QueryStatusEq, f.Status)
	}
	if f.GudangID != 0 {
		q = q.Where(constant.QueryGudangIDEq, f.GudangID)
	}
	return q
}

func (r *repository) List(p utils.PaginationParams, f Filter) ([]model.StockOpname, int64, error) {
	var list []model.StockOpname
	var total int64

	q := applyFilter(r.db.Model(&model.StockOpname{}), f)
	if p.Search != "" {
		q = q.Where("nomor_opname ILIKE ?", "%"+p.Search+"%")
	}
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := p.Apply(q.Session(&gorm.Session{}).
		Preload("Gudang").Preload("Items").
		Preload("Items.Barang", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Order("id desc")).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *repository) FindByID(id uint) (*model.StockOpname, error) {
	var so model.StockOpname

	err := r.db.Preload("Gudang").Preload("Items").
		Preload("Items.Barang", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		First(&so, id).Error
	if err != nil {
		return nil, err
	}
	return &so, nil
}

func (r *repository) FindByNomor(nomor string) (*model.StockOpname, error) {
	var so model.StockOpname
	if err := r.db.Where("nomor_opname = ?", nomor).First(&so).Error; err != nil {
		return nil, err
	}
	return &so, nil
}

func buildItems(tx *gorm.DB, soID uint, gudangID uint, inputs []ItemInput) ([]model.StockOpnameItem, error) {
	items := make([]model.StockOpnameItem, 0, len(inputs))
	for _, in := range inputs {
		if err := tx.First(&model.Barang{}, in.BarangID).Error; err != nil {
			return nil, err
		}
		stokSistem, err := barangstokgudang.GetStokGudangTx(tx, in.BarangID, gudangID)
		if err != nil {
			return nil, err
		}
		item := model.StockOpnameItem{
			StockOpnameID: soID,
			BarangID:      in.BarangID,
			StokSistem:    stokSistem,
			StokFisik:     in.StokFisik,
			Catatan:       in.Catatan,
		}
		item.HitungSelisih()
		items = append(items, item)
	}
	return items, nil
}

func (r *repository) Create(so *model.StockOpname, inputs []ItemInput) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(so).Error; err != nil {
			return err
		}
		items, err := buildItems(tx, so.ID, so.GudangID, inputs)
		if err != nil {
			return err
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		so.Items = items
		return nil
	})
}

func (r *repository) Update(so *model.StockOpname, inputs []ItemInput) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("stock_opname_id = ?", so.ID).Delete(&model.StockOpnameItem{}).Error; err != nil {
			return err
		}
		items, err := buildItems(tx, so.ID, so.GudangID, inputs)
		if err != nil {
			return err
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		return tx.Save(so).Error
	})
}

func (r *repository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("stock_opname_id = ?", id).Delete(&model.StockOpnameItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.StockOpname{}, id).Error
	})
}

func (r *repository) Complete(id uint, userID uint) (*model.StockOpname, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var so model.StockOpname
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Preload("Items").First(&so, id).Error; err != nil {
			return err
		}
		if so.Status != constant.StatusSODraft {
			return errors.New(constant.ErrSOBukanDraft)
		}

		for _, item := range so.Items {
			if item.Selisih == 0 {
				continue
			}
			if err := barangstokgudang.SetStokGudangTx(tx, item.BarangID, so.GudangID, item.StokFisik); err != nil {
				return err
			}
			if err := barangstokgudang.SyncBarangStokTotalTx(tx, item.BarangID); err != nil {
				return err
			}
		}

		now := time.Now()
		return tx.Model(&model.StockOpname{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":         constant.StatusSOSelesai,
			"dilakukan_oleh": userID,
			"completed_at":   now,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

func (r *repository) Batalkan(id uint) (*model.StockOpname, error) {
	res := r.db.Model(&model.StockOpname{}).Where("id = ? AND status = ?", id, constant.StatusSODraft).
		Update("status", constant.StatusSODibatalkan)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, errors.New(constant.ErrSOBukanDraft)
	}
	return r.FindByID(id)
}

func (r *repository) CountByStatus(status string) (int64, error) {
	var count int64
	err := r.db.Model(&model.StockOpname{}).Where(constant.QueryStatusEq, status).Count(&count).Error
	return count, err
}
