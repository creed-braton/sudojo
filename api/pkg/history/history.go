package history

import (
	"fmt"
	"sync"
)

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
// Holds an initial and current map of artifacts. The initial artifacts were
// provided to the constructor of the instance and are expected to be persisted.
// The new artifact are kept seperate to remember which artifacts needs to be
// persisted when tearing down the history instance.
type History interface {
	// Appends a new artifact to the history, drops (player, row, col, val)
	// duplicates.
	Append(a Artifact)
	// Returns a list of all artifacts (initial and current combined) in the
	// history.
	Artifacts() []Artifact
	// Writes all current artifacts to the initial map and returns which
	// artifacts were merged so that they can be persisted.
	Flush() []Artifact
}

type history struct {
	initial   map[string]Artifact
	artifacts map[string]Artifact
	lock      sync.RWMutex
}

var _ History = &history{}

// Creates a new history instance with the provided initial artifacts.
func New(artifacts []Artifact) *history {
	initial := make(map[string]Artifact)
	if artifacts != nil {
		for _, a := range artifacts {
			key := fmt.Sprintf(
				"%s-%d-%d-%d", a.Player(),
				a.Row(), a.Column(), a.Value(),
			)

			initial[key] = a
		}
	}
	return &history{
		initial:   initial,
		artifacts: make(map[string]Artifact),
	}
}

func (h *history) Append(a Artifact) {
	h.lock.Lock()
	defer h.lock.Unlock()

	key := fmt.Sprintf(
		"%s-%d-%d-%d", a.Player(),
		a.Row(), a.Column(), a.Value(),
	)

	if _, exist := h.initial[key]; exist {
		return
	}
	if _, exist := h.artifacts[key]; exist {
		return
	}

	h.artifacts[key] = a
}

func (h *history) Artifacts() []Artifact {
	h.lock.RLock()
	defer h.lock.RUnlock()

	c := make([]Artifact, 0, len(h.initial)+len(h.artifacts))
	for _, a := range h.initial {
		c = append(c, a)
	}
	for _, a := range h.artifacts {
		c = append(c, a)
	}
	return c
}

func (h *history) Flush() []Artifact {
	h.lock.Lock()
	defer h.lock.Unlock()

	flushed := make([]Artifact, 0, len(h.artifacts))
	for key, a := range h.artifacts {
		h.initial[key] = a
		flushed = append(flushed, a)
	}
	h.artifacts = make(map[string]Artifact)
	return flushed
}
