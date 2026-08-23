package commands

import "time"

func RunDetached(remaining time.Duration, name string) error {
	time.Sleep(remaining)
	return NotifyComplete(name)
}
