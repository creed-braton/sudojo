package lobby

import (
	"sudojo/pkg/game"
	"sudojo/pkg/history"
	"sudojo/pkg/player"

	"github.com/google/uuid"
)

// Creates a mock lobby with a pre-initialized game from game.NewMock().
// Invalid parameters may produce undefined state. Caller must ensure
// parameters are valid as no error is returned for simplicity.
func NewMock(start, strict, ping bool, maxSize int) *lobby {
	config, _ := NewConfig(strict, ping, false, maxSize)

	return New(
		uuid.NewString(),
		config,
		game.NewMock(start),
		history.New(nil),
		make(map[string]player.Player),
	)
}
