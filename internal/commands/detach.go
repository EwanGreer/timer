package commands

import "time"

// RunDetached waits out the timer without a UI and fires the completion
// notification once. It is the body of a timer started with --detach.
func RunDetached(remaining time.Duration, name string) error {
	time.Sleep(remaining)
	return NotifyComplete(name)
}
