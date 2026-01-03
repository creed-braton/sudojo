package lobby

import "errors"

var (
	ErrInvalidSize = errors.New("invalid maximum lobby size")
)

// Config represents the configuration settings for a lobby.
type Config interface {
	// Returns whether the lobby uses strict mode validation.
	Strict() bool
	// Returns whether ping functionality is enabled.
	Ping() bool
	// Returns whether notes functionality is enabled.
	Notes() bool
	// Returns the maximum number of players allowed in the lobby.
	MaxSize() int
}

type config struct {
	strict  bool
	ping    bool
	notes   bool
	maxSize int
}

var _ Config = &config{}

// Creates a new config with the specified settings. Returns
// ErrInvalidSize if maxSize is not between 1 and 8.
func NewConfig(strict, ping, notes bool, maxSize int) (*config, error) {
	if maxSize < 1 || maxSize > 8 {
		return nil, ErrInvalidSize
	}

	return &config{
		strict:  strict,
		ping:    ping,
		notes:   notes,
		maxSize: maxSize,
	}, nil
}

func (c *config) Strict() bool {
	return c.strict
}

func (c *config) Ping() bool {
	return c.ping
}

func (c *config) Notes() bool {
	return c.notes
}

func (c *config) MaxSize() int {
	return c.maxSize
}
