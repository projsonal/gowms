package selfupdate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type State string

const (
	StateIdle    State = "idle"
	StateRunning State = "running"
	StateSuccess State = "success"
	StateFailed  State = "failed"
)

// Status — dipersist ke file JSON (BUKAN disimpan di memori proses) —
// ini krusial: begitu update berhasil di-build, proses backend yang
// sedang berjalan SENGAJA mematikan dirinya sendiri supaya systemd
// menyalakannya ulang memakai binary baru (lihat catatan panjang di
// update_controller.go). Kalau status cuma disimpan di memori, begitu
// proses lama mati, statusnya ikut hilang, dan halaman Settings yang
// sedang polling jadi bingung — dengan file, proses BARU (hasil restart)
// tetap bisa membaca status yang sama persis.
type Status struct {
	State           State      `json:"state"`
	Message         string     `json:"message"`
	FromVersion     string     `json:"from_version,omitempty"`
	ToVersion       string     `json:"to_version,omitempty"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
	MaintenanceAuto bool       `json:"maintenance_auto"` // true kalau Mode Pemeliharaan diaktifkan OTOMATIS oleh proses update ini (bukan sudah aktif manual sebelumnya) — dipakai memutuskan apakah boleh dimatikan otomatis juga
	Acknowledged    bool       `json:"acknowledged"`     // true setelah proses backend yang hidup sempat menindaklanjuti (mis. matikan Mode Pemeliharaan) — mencegah tindak lanjut dobel
}

// ReadStatus membaca file status — file belum ada dianggap idle (belum
// pernah ada proses update sama sekali), BUKAN error.
func ReadStatus(path string) (Status, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Status{State: StateIdle}, nil
		}
		return Status{}, err
	}
	var s Status
	if err := json.Unmarshal(data, &s); err != nil {
		return Status{}, err
	}
	return s, nil
}

// WriteStatus menulis file status secara ATOMIK (tulis ke file sementara
// lalu rename) supaya proses lain yang sedang membaca (mis. skrip deploy
// & handler HTTP jalan bersamaan) tidak pernah melihat file JSON yang
// setengah tertulis/korup.
func WriteStatus(path string, s Status) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
