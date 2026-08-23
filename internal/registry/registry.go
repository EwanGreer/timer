// Package registry tracks running detached timers as one JSON file per
// timer in a running/ directory.
package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Entry describes a timer to record.
type Entry struct {
	Name      string
	Duration  time.Duration
	StartedAt time.Time
}

// record is the on-disk JSON shape.
type record struct {
	Name      string    `json:"name"`
	DurationS int64     `json:"duration_seconds"`
	StartedAt time.Time `json:"started_at"`
}

// Write records the timer as <os.Getpid()>.json in dir, atomically via a
// temp file and rename.
func Write(dir string, e Entry) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	rec := record{Name: e.Name, DurationS: int64(e.Duration / time.Second), StartedAt: e.StartedAt}
	b, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	tmp := filepath.Join(dir, fmt.Sprintf(".%d.json.tmp", os.Getpid()))
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(dir, fmt.Sprintf("%d.json", os.Getpid())))
}

// Remove deletes the registry file for pid. A missing file is not an
// error.
func Remove(dir string, pid int) error {
	err := os.Remove(filepath.Join(dir, fmt.Sprintf("%d.json", pid)))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
