package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteCreatesFileWithRecord(t *testing.T) {
	dir := t.TempDir()
	started := time.Date(2026, 8, 23, 14, 32, 5, 0, time.Local)

	if err := Write(dir, Entry{Name: "Tea", Duration: 5 * time.Minute, StartedAt: started}); err != nil {
		t.Fatalf("Write: %v", err)
	}

	path := filepath.Join(dir, fmt.Sprintf("%d.json", os.Getpid()))
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var rec record
	if err := json.Unmarshal(b, &rec); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if rec.Name != "Tea" || rec.DurationS != 300 || !rec.StartedAt.Equal(started) {
		t.Fatalf("record = %+v, want Tea 300s %v", rec, started)
	}
}

func TestWriteCreatesMissingDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "running")

	if err := Write(dir, Entry{Duration: time.Minute}); err != nil {
		t.Fatalf("Write: %v", err)
	}
}

func TestRemoveDeletesFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "4242.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Remove(dir, 4242); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "4242.json")); !os.IsNotExist(err) {
		t.Fatalf("file still exists after Remove (stat err = %v)", err)
	}
}

func TestRemoveMissingFileIsNotAnError(t *testing.T) {
	if err := Remove(t.TempDir(), 9999); err != nil {
		t.Fatalf("Remove: %v", err)
	}
}
