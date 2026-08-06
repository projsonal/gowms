package health

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"gorm.io/gorm"
)

const (
	statusUp   = "up"
	statusDown = "down"

	dbPingTimeout = 2 * time.Second
)

type ComponentStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type Report struct {
	Status     string            `json:"status"`
	UptimeSec  int64             `json:"uptime_seconds"`
	Components []ComponentStatus `json:"components"`
}

type Checker struct {
	db          *gorm.DB
	storagePath string
	startedAt   time.Time
}

func NewChecker(db *gorm.DB, storagePath string) *Checker {
	return &Checker{
		db:          db,
		storagePath: storagePath,
		startedAt:   time.Now(),
	}
}

func (c *Checker) CheckDatabase() ComponentStatus {
	comp := ComponentStatus{Name: "database"}

	if c.db == nil {
		comp.Status = statusDown
		comp.Message = "koneksi database belum diinisialisasi"
		return comp
	}

	sqlDB, err := c.db.DB()
	if err != nil {
		comp.Status = statusDown
		comp.Message = "gagal mengambil koneksi sql.DB: " + err.Error()
		return comp
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbPingTimeout)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		comp.Status = statusDown
		comp.Message = "ping database gagal: " + err.Error()
		return comp
	}

	comp.Status = statusUp
	return comp
}

func (c *Checker) CheckStorage() ComponentStatus {
	comp := ComponentStatus{Name: "storage"}

	if c.storagePath == "" {
		comp.Status = statusUp
		comp.Message = "storage tidak dikonfigurasi, pemeriksaan dilewati"
		return comp
	}

	info, err := os.Stat(c.storagePath)
	if err != nil {
		comp.Status = statusDown
		comp.Message = "direktori storage tidak dapat diakses: " + err.Error()
		return comp
	}
	if !info.IsDir() {
		comp.Status = statusDown
		comp.Message = "STORAGE_PATH bukan direktori"
		return comp
	}

	probe := filepath.Join(c.storagePath, ".health-check")
	if err := os.WriteFile(probe, []byte("ok"), 0o600); err != nil {
		comp.Status = statusDown
		comp.Message = "direktori storage tidak dapat ditulis: " + err.Error()
		return comp
	}
	_ = os.Remove(probe)

	comp.Status = statusUp
	return comp
}

func (c *Checker) Report() Report {
	components := []ComponentStatus{
		c.CheckDatabase(),
		c.CheckStorage(),
	}

	overall := statusUp
	for _, comp := range components {
		if comp.Status == statusDown {
			overall = statusDown
			break
		}
	}

	return Report{
		Status:     overall,
		UptimeSec:  int64(time.Since(c.startedAt).Seconds()),
		Components: components,
	}
}
