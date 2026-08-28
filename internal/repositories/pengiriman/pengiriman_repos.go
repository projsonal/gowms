package pengiriman

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

const idStatusQuery = "id = ? AND status = ?"

func applyFilter(q *gorm.DB, f Filter) *gorm.DB {
	if f.Status != "" {
		q = q.Where(constant.QueryStatusEq, f.Status)
	}
	if f.GudangAsalID != 0 {
		q = q.Where("gudang_asal_id = ?", f.GudangAsalID)
	}
	if f.Jenis != "" {
		q = q.Where("jenis_pengambilan = ?", f.Jenis)
	}
	return q
}

func (r *repository) List(p utils.PaginationParams, f Filter) ([]model.Pengiriman, int64, error) {
	var list []model.Pengiriman
	var total int64

	q := applyFilter(r.db.Model(&model.Pengiriman{}), f)
	if p.Search != "" {
		q = q.Where("nomor_pengiriman ILIKE ? OR nama_penerima ILIKE ?", "%"+p.Search+"%", "%"+p.Search+"%")
	}
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := p.Apply(q.Session(&gorm.Session{}).Preload("GudangAsal").Order("id desc")).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *repository) FindByID(id uint) (*model.Pengiriman, error) {
	var pg model.Pengiriman
	err := r.db.Preload("GudangAsal").
		Preload("BarangKeluar").
		Preload("BarangKeluar.Items").
		Preload("BarangKeluar.Items.Barang").
		Preload("BarangKeluar.Items.Barang.Satuan").
		First(&pg, id).Error
	if err != nil {
		return nil, err
	}
	return &pg, nil
}

func (r *repository) FindByNomor(nomor string) (*model.Pengiriman, error) {
	var pg model.Pengiriman
	if err := r.db.Where("nomor_pengiriman = ?", nomor).First(&pg).Error; err != nil {
		return nil, err
	}
	return &pg, nil
}

func (r *repository) Create(pg *model.Pengiriman) error {
	return r.db.Create(pg).Error
}

func (r *repository) Update(pg *model.Pengiriman) error {
	return r.db.Save(pg).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&model.Pengiriman{}, id).Error
}

func (r *repository) Jadwalkan(id uint, namaKurir, teleponKurir string, estimasiTiba *time.Time) (*model.Pengiriman, error) {
	res := r.db.Model(&model.Pengiriman{}).Where(idStatusQuery, id, constant.StatusPGDraft).
		Updates(map[string]interface{}{
			"status":        constant.StatusPGDijadwalkan,
			"nama_kurir":    namaKurir,
			"telepon_kurir": teleponKurir,
			"estimasi_tiba": estimasiTiba,
		})
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, errors.New(constant.ErrPGBukanDraft)
	}
	return r.FindByID(id)
}

func (r *repository) Mulai(id uint) (*model.Pengiriman, error) {
	res := r.db.Model(&model.Pengiriman{}).Where(idStatusQuery, id, constant.StatusPGDijadwalkan).
		Update("status", constant.StatusPGDalamPerjalanan)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, errors.New(constant.ErrPGBukanDijadwalkan)
	}
	return r.FindByID(id)
}

func (r *repository) RecordLocation(id uint, lat, lng float64, kecepatan *float64, recordedAt time.Time) (*model.Pengiriman, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var pg model.Pengiriman
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&pg, id).Error; err != nil {
			return err
		}
		if pg.Status != constant.StatusPGDalamPerjalanan {
			return errors.New(constant.ErrBMBukanDraft)
		}

		point := model.PengirimanTrackingPoint{
			PengirimanID: id,
			Lat:          lat,
			Lng:          lng,
			KecepatanKmh: kecepatan,
			RecordedAt:   recordedAt,
		}
		if err := tx.Create(&point).Error; err != nil {
			return err
		}

		return tx.Model(&model.Pengiriman{}).Where("id = ?", id).Updates(map[string]interface{}{
			"last_lat":         lat,
			"last_lng":         lng,
			"last_location_at": recordedAt,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

func (r *repository) ListTrackingPoints(pengirimanID uint, limit int) ([]model.PengirimanTrackingPoint, error) {
	if limit <= 0 || limit > 1000 {
		limit = 200
	}
	var points []model.PengirimanTrackingPoint
	err := r.db.Where("pengiriman_id = ?", pengirimanID).Order("recorded_at desc").Limit(limit).Find(&points).Error
	return points, err
}

func (r *repository) Selesaikan(id uint, catatan string) (*model.Pengiriman, error) {
	now := time.Now()
	res := r.db.Model(&model.Pengiriman{}).Where(idStatusQuery, id, constant.StatusPGDalamPerjalanan).
		Updates(map[string]interface{}{
			"status":  constant.StatusPGTerkirim,
			"catatan": catatan,

			"waktu_terkirim": now,
		})
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, errors.New(constant.ErrPGBukanPerjalan)
	}
	return r.FindByID(id)
}

func (r *repository) Batalkan(id uint) (*model.Pengiriman, error) {
	res := r.db.Model(&model.Pengiriman{}).
		Where("id = ? AND status IN ?", id, []string{constant.StatusPGDraft, constant.StatusPGDijadwalkan}).
		Update("status", constant.StatusPGDibatalkan)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, errors.New(constant.ErrPGBukanDraft)
	}
	return r.FindByID(id)
}

func (r *repository) SetProtected(id uint, protected bool) (*model.Pengiriman, error) {
	if err := r.db.Model(&model.Pengiriman{}).Where("id = ?", id).
		Update("is_protected", protected).Error; err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

func (r *repository) CountByStatus(status string) (int64, error) {
	var count int64
	err := r.db.Model(&model.Pengiriman{}).Where(constant.QueryStatusEq, status).Count(&count).Error
	return count, err
}
