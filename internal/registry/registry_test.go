package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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

func stubProc(t *testing.T, alive bool, start time.Time, startErr error) {
	t.Helper()
	oa, os_ := procAlive, procStartedAt
	procAlive = func(int) bool { return alive }
	procStartedAt = func(int) (time.Time, error) { return start, startErr }
	t.Cleanup(func() { procAlive, procStartedAt = oa, os_ })
}

func TestReadReturnsWrittenTimer(t *testing.T) {
	dir := t.TempDir()
	started := time.Now().Add(-10 * time.Second)
	if err := Write(dir, Entry{Name: "Tea", Duration: 5 * time.Minute, StartedAt: started}); err != nil {
		t.Fatal(err)
	}
	stubProc(t, true, started, nil)

	timers, err := Read(dir)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(timers) != 1 {
		t.Fatalf("len = %d, want 1", len(timers))
	}
	got := timers[0]
	if got.Pid != os.Getpid() || got.Name != "Tea" {
		t.Fatalf("timer = %+v, want pid %d name Tea", got, os.Getpid())
	}
	if got.Remaining > 290*time.Second || got.Remaining < 285*time.Second {
		t.Fatalf("remaining = %v, want about 290s", got.Remaining)
	}
}

func TestReadRemovesDeadPid(t *testing.T) {
	dir := t.TempDir()
	if err := Write(dir, Entry{Duration: time.Minute, StartedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	stubProc(t, false, time.Now(), nil)

	timers, err := Read(dir)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(timers) != 0 {
		t.Fatalf("len = %d, want 0", len(timers))
	}
	matches, _ := filepath.Glob(filepath.Join(dir, "*.json"))
	if len(matches) != 0 {
		t.Fatalf("stale files left behind: %v", matches)
	}
}

func TestReadRemovesStartTimeMismatch(t *testing.T) {
	dir := t.TempDir()
	if err := Write(dir, Entry{Duration: time.Minute, StartedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	stubProc(t, true, time.Now().Add(-time.Hour), nil)

	timers, err := Read(dir)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(timers) != 0 {
		t.Fatalf("len = %d, want 0", len(timers))
	}
}

func TestReadRemovesUnparseableFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "999.json"), []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	stubProc(t, true, time.Now(), nil)

	timers, err := Read(dir)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(timers) != 0 {
		t.Fatalf("len = %d, want 0", len(timers))
	}
	if _, err := os.Stat(filepath.Join(dir, "999.json")); !os.IsNotExist(err) {
		t.Fatalf("unparseable file not removed (stat err = %v)", err)
	}
}

func TestReadMissingDirReturnsEmpty(t *testing.T) {
	timers, err := Read(filepath.Join(t.TempDir(), "nope"))
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(timers) != 0 {
		t.Fatalf("len = %d, want 0", len(timers))
	}
}

func TestReadSortsOldestFirst(t *testing.T) {
	dir := t.TempDir()
	older := time.Date(2026, 8, 23, 14, 30, 0, 0, time.Local)
	newer := older.Add(5 * time.Minute)
	for pid, started := range map[int]time.Time{200: newer, 100: older} {
		b, err := json.Marshal(record{Name: fmt.Sprintf("pid%d", pid), DurationS: 300, StartedAt: started})
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("%d.json", pid)), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	oa, os_ := procAlive, procStartedAt
	procAlive = func(int) bool { return true }
	procStartedAt = func(pid int) (time.Time, error) {
		if pid == 100 {
			return older, nil
		}
		return newer, nil
	}
	t.Cleanup(func() { procAlive, procStartedAt = oa, os_ })

	timers, err := Read(dir)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(timers) != 2 {
		t.Fatalf("len = %d, want 2", len(timers))
	}
	if timers[0].Name != "pid100" || timers[1].Name != "pid200" {
		t.Fatalf("order = %q, %q, want pid100 then pid200", timers[0].Name, timers[1].Name)
	}
}

func TestReadKeepsTimersWithoutProcSupport(t *testing.T) {
	dir := t.TempDir()
	if err := Write(dir, Entry{Name: "Tea", Duration: 5 * time.Minute, StartedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	stubProc(t, false, time.Now(), nil)
	old := procChecksSupported
	procChecksSupported = false
	t.Cleanup(func() { procChecksSupported = old })

	timers, err := Read(dir)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(timers) != 1 {
		t.Fatalf("len = %d, want 1", len(timers))
	}
	matches, _ := filepath.Glob(filepath.Join(dir, "*.json"))
	if len(matches) != 1 {
		t.Fatalf("files after Read = %v, want 1", len(matches))
	}
}

func TestDefaultProcStartedAtOwnProcess(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("process start time not supported on this platform")
	}
	start, err := defaultProcStartedAt(os.Getpid())
	if err != nil {
		t.Fatalf("defaultProcStartedAt: %v", err)
	}
	if abs(time.Since(start)) > time.Minute {
		t.Fatalf("own process start = %v, want within 1 minute of now", start)
	}
}
