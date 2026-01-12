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
	Pings() bool
	// Returns whether notes functionality is enabled.
	Notes() bool
	// Returns the maximum number of players allowed in the lobby.
	MaxSize() int
}

type config struct {
	strict  bool
	pings   bool
	notes   bool
	maxSize int
}

var _ Config = &config{}

func validSize(maxSize int) bool {
	return maxSize >= 1 && maxSize <= 8
}

// Creates a new config with the specified settings. Returns
// ErrInvalidSize if maxSize is not between 1 and 8.
func NewConfig(strict, pings, notes bool, maxSize int) (*config, error) {
	if !validSize(maxSize) {
		return nil, ErrInvalidSize
	}

	return &config{
		strict:  strict,
		pings:   pings,
		notes:   notes,
		maxSize: maxSize,
	}, nil
}

func (c *config) Strict() bool {
	return c.strict
}

func (c *config) Pings() bool {
	return c.pings
}

func (c *config) Notes() bool {
	return c.notes
}

func (c *config) MaxSize() int {
	return c.maxSize
}
