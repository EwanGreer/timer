package commands

import "github.com/EwanGreer/timer/internal/art"

// artFor returns the model's art set, falling back to the built-in art when
// none was provided.
func artFor(s *art.Set) *art.Set {
	if s == nil {
		return art.Default()
	}
	return s
}
