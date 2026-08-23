package commands

import (
	"testing"
	"time"
)

type notifyRecord struct {
	calls   int
	title   string
	message string
}

func stubNotify(t *testing.T) *notifyRecord {
	t.Helper()

	var rec notifyRecord
	orig := notify
	notify = func(title, message string, icon any) error {
		rec.calls++
		rec.title = title
		rec.message = message
		return nil
	}
	t.Cleanup(func() { notify = orig })

	return &rec
}

func TestStartModelViewDoesNotFireNotification(t *testing.T) {
	rec := stubNotify(t)

	m := StartModel{done: true}
	_ = m.View()

	if rec.calls != 0 {
		t.Fatalf("View fired %d notifications, want 0", rec.calls)
	}
}

func TestStartModelUpdateFiresNotificationOnceOnDone(t *testing.T) {
	rec := stubNotify(t)

	m := StartModel{Remaining: 0}
	_, cmd := m.Update(tickMsg(time.Now()))
	if cmd == nil {
		t.Fatal("expected a command that fires the notification")
	}
	cmd()

	if rec.calls != 1 {
		t.Fatalf("notification fired %d times, want 1", rec.calls)
	}
}

func TestStartModelUpdateNotificationIncludesName(t *testing.T) {
	rec := stubNotify(t)

	m := StartModel{Remaining: 0, Name: "Tea"}
	_, cmd := m.Update(tickMsg(time.Now()))
	cmd()

	if rec.calls != 1 {
		t.Fatalf("notification fired %d times, want 1", rec.calls)
	}
	if rec.title != "Your Timer is Complete!" {
		t.Errorf("title = %q, want %q", rec.title, "Your Timer is Complete!")
	}
	if rec.message != `Your timer "Tea" is completed!` {
		t.Errorf("message = %q, want %q", rec.message, `Your timer "Tea" is completed!`)
	}
}

func TestStartModelUpdateNotificationWithoutNameKeepsDefaultMessage(t *testing.T) {
	rec := stubNotify(t)

	m := StartModel{Remaining: 0}
	_, cmd := m.Update(tickMsg(time.Now()))
	cmd()

	if rec.message != "Your timer is completed!" {
		t.Errorf("message = %q, want %q", rec.message, "Your timer is completed!")
	}
}
