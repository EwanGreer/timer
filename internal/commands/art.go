package commands

import "github.com/EwanGreer/timer/internal/art"

func artFor(s *art.Set) *art.Set {
	if s == nil {
		return art.Default()
	}
	return s
}
