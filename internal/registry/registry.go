package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Entry struct {
	Name      string
	Duration  time.Duration
	StartedAt time.Time
}

type record struct {
	Name      string    `json:"name"`
	DurationS int64     `json:"duration_seconds"`
	StartedAt time.Time `json:"started_at"`
}

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

func Remove(dir string, pid int) error {
	err := os.Remove(filepath.Join(dir, fmt.Sprintf("%d.json", pid)))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

type Timer struct {
	Pid       int
	Name      string
	Duration  time.Duration
	StartedAt time.Time
	Remaining time.Duration
}

// procAlive and procStartedAt are variables so tests can stub them.
var procAlive = defaultProcAlive
var procStartedAt = defaultProcStartedAt

// procChecksSupported is set in init by the platform files. When it stays
// false, Read cannot tell stale files from live timers and keeps every file.
var procChecksSupported = false

const startTimeTolerance = 2 * time.Second

func Read(dir string) ([]Timer, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	now := time.Now()
	var timers []Timer
	for _, de := range entries {
		name := de.Name()
		if de.IsDir() || !strings.HasSuffix(name, ".json") {
			continue
		}
		pid, err := strconv.Atoi(strings.TrimSuffix(name, ".json"))
		if err != nil {
			removeStale(filepath.Join(dir, name))
			continue
		}
		path := filepath.Join(dir, name)
		rec, err := readRecord(path)
		if err != nil {
			removeStale(path)
			continue
		}
		if procChecksSupported {
			if !procAlive(pid) {
				removeStale(path)
				continue
			}
			start, err := procStartedAt(pid)
			if err != nil || abs(start.Sub(rec.StartedAt)) > startTimeTolerance {
				removeStale(path)
				continue
			}
		}
		remaining := rec.StartedAt.Add(time.Duration(rec.DurationS) * time.Second).Sub(now)
		if remaining < 0 {
			remaining = 0
		}
		timers = append(timers, Timer{
			Pid:       pid,
			Name:      rec.Name,
			Duration:  time.Duration(rec.DurationS) * time.Second,
			StartedAt: rec.StartedAt,
			Remaining: remaining,
		})
	}
	sort.Slice(timers, func(i, j int) bool {
		return timers[i].StartedAt.Before(timers[j].StartedAt)
	})
	return timers, nil
}

func readRecord(path string) (record, error) {
	var rec record
	b, err := os.ReadFile(path)
	if err != nil {
		return rec, err
	}
	err = json.Unmarshal(b, &rec)
	return rec, err
}

// Failures are ignored; the next Read retries the removal.
func removeStale(path string) {
	os.Remove(path)
}

func abs(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}
