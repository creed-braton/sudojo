package manager

import (
	"sudojo/pkg/event"
	"sudojo/pkg/lobby"
)

// Creates a mock manager with a pre-initialized game from game.NewMock().
// Invalid parameters may produce undefined state. Caller must ensure
// parameters are valid as no error is returned for simplicity.
func NewMock(start, strict bool, maxSize int) *manager {
	return New(
		lobby.NewMock(start, strict, maxSize),
		event.NewHub(),
	)
}
