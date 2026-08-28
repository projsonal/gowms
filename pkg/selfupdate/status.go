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

type Status struct {
	State           State      `json:"state"`
	Message         string     `json:"message"`
	FromVersion     string     `json:"from_version,omitempty"`
	ToVersion       string     `json:"to_version,omitempty"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
	MaintenanceAuto bool       `json:"maintenance_auto"`
	Acknowledged    bool       `json:"acknowledged"`
}

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
