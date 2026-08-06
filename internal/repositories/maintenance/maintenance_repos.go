package maintenance

import (
	"time"

	"gorm.io/gorm"

	"github.com/projsonal/gostock/internal/model"
)

const singletonID = 1

func (r *repository) Get() (*model.MaintenanceStatus, error) {
	var status model.MaintenanceStatus
	err := r.db.First(&status, singletonID).Error
	if err == nil {
		return &status, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	status = model.MaintenanceStatus{ID: singletonID, IsActive: false}
	if err := r.db.Create(&status).Error; err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *repository) Set(active bool, message string, estimatedUntil *time.Time, updatedBy uint) (*model.MaintenanceStatus, error) {
	if _, err := r.Get(); err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"is_active":       active,
		"message":         message,
		"estimated_until": estimatedUntil,
		"updated_by":      updatedBy,
	}
	if active {
		updates["started_at"] = time.Now()
	} else {
		updates["started_at"] = nil
		updates["estimated_until"] = nil
	}

	if err := r.db.Model(&model.MaintenanceStatus{}).Where("id = ?", singletonID).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.Get()
}
