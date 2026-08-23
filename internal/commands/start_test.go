package commands

import (
	"testing"
	"time"
)

// stubNotify replaces the package notify function for the duration of a test.
func stubNotify(t *testing.T) *int {
	t.Helper()

	calls := 0
	orig := notify
	notify = func(title, message string, icon any) error {
		calls++
		return nil
	}
	t.Cleanup(func() { notify = orig })

	return &calls
}

func TestStartModelViewDoesNotFireNotification(t *testing.T) {
	calls := stubNotify(t)

	m := StartModel{done: true}
	_ = m.View()

	if *calls != 0 {
		t.Fatalf("View fired %d notifications, want 0", *calls)
	}
}

func TestStartModelUpdateFiresNotificationOnceOnDone(t *testing.T) {
	calls := stubNotify(t)

	m := StartModel{Remaining: 0}
	_, cmd := m.Update(tickMsg(time.Now()))
	if cmd == nil {
		t.Fatal("expected a command that fires the notification")
	}
	cmd()

	if *calls != 1 {
		t.Fatalf("notification fired %d times, want 1", *calls)
	}
}
