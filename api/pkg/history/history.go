package history

import "sync"

// Represents an insert artifact which holds metadata about an insert event.
type Artifact interface {
	// Token of the player who initiated the event.
	Player() string
	// Timestamp when the event occurred.
	Timestamp() int64
	// Row of the insertion.
	Row() int
	// Column of the insertion.
	Column() int
	// Value of the insertion.
	Value() int
}

type artifact struct {
	player    string
	timestamp int64
	row       int
	column    int
	value     int
}

var _ Artifact = &artifact{}

// Creates a new artifact with provided values.
func NewArtifact(player string, ts int64, row, col, val int) *artifact {
	return &artifact{
		player:    player,
		timestamp: ts,
		row:       row,
		column:    col,
		value:     val,
	}
}

func (a *artifact) Player() string {
	return a.player
}

func (a *artifact) Timestamp() int64 {
	return a.timestamp
}

func (a *artifact) Row() int {
	return a.row
}

func (a *artifact) Column() int {
	return a.column
}

func (a *artifact) Value() int {
	return a.value
}

// Thread-safe lobby history to track insertion events of the game session.
// Holds an initial and current list of artifacts. The initial artifacts were
// provided to the constructor of the instance and are expected to be persisted
// The new artifact are kept seperate to remember which artifacts needs to be
// persisted when tearing down the history instance.
type History interface {
	// Appends a new artifact to the history.
	Append(a Artifact)
	// Returns a list of all artifacts (initial and current combined) in the
	// history.
	Artifacts() []Artifact
	// Writes all current artifacts to the initial list and returns which
	// artifacts were merged so that they can be persisted.
	Flush() []Artifact
}

type history struct {
	initial   []Artifact
	artifacts []Artifact
	lock      sync.RWMutex
}

var _ History = &history{}

// Creates a new history instance with the provided initial artifacts.
func New(initial []Artifact) *history {
	if initial == nil {
		initial = make([]Artifact, 0)
	}
	return &history{initial: initial, artifacts: make([]Artifact, 0)}
}

func (h *history) Append(a Artifact) {
	h.lock.Lock()
	h.artifacts = append(h.artifacts, a)
	h.lock.Unlock()
}

func (h *history) Artifacts() []Artifact {
	h.lock.RLock()
	defer h.lock.RUnlock()

	c := make([]Artifact, 0, len(h.initial)+len(h.artifacts))
	c = append(c, h.initial...)
	c = append(c, h.artifacts...)

	return c
}

func (h *history) Flush() []Artifact {
	h.lock.Lock()
	defer h.lock.Unlock()

	flushed := h.artifacts
	h.initial = append(h.initial, h.artifacts...)
	h.artifacts = make([]Artifact, 0)

	return flushed
}
