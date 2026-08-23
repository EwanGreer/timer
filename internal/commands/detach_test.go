package commands

import (
	"testing"
	"time"
)

func TestRunDetachedNotifiesWithName(t *testing.T) {
	rec := stubNotify(t)

	RunDetached(time.Millisecond, "Tea")

	if rec.calls != 1 {
		t.Fatalf("notification fired %d times, want 1", rec.calls)
	}
	if rec.message != `Your timer "Tea" is completed!` {
		t.Errorf("message = %q, want %q", rec.message, `Your timer "Tea" is completed!`)
	}
}

func TestRunDetachedNotifiesWithoutName(t *testing.T) {
	rec := stubNotify(t)

	RunDetached(time.Millisecond, "")

	if rec.calls != 1 {
		t.Fatalf("notification fired %d times, want 1", rec.calls)
	}
	if rec.message != "Your timer is completed!" {
		t.Errorf("message = %q, want %q", rec.message, "Your timer is completed!")
	}
}
