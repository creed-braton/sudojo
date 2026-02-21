package history

import (
	"fmt"
	"sort"
	"sync"
)

// Represents an artifact which holds metadata about an insert event.
type Artifact interface {
	// Token of the player who initiated the event.
	Player() string
	// Nanosecond timestamp when the event occurred.
	Timestamp() int64
	// Row of the insertion.
	Row() int
	// Column of the insertion.
	Column() int
	// Value of the insertion.
	Value() int
}

type artifact struct {
	row       int
	column    int
	value     int
	player    string
	timestamp int64
}

var _ Artifact = &artifact{}

// Creates a new artifact with the token of the player which made the insertion,
// row and column of the cell position, the inserted value, and the current time
// in nanoseconds.
func NewArtifact(row, col, val int, player string, now int64) *artifact {
	return &artifact{
		row:       row,
		column:    col,
		value:     val,
		player:    player,
		timestamp: now,
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

// Concurrency-safe lobby history to track insertion events of the game session.
// Holds an initial and current map of artifacts. The initial artifacts were
// provided to the constructor of the instance and are expected to already be
// persisted. The new artifact are kept seperate to remember which artifacts needs
// to be persisted when tearing down the history instance.
type History interface {
	// Appends a new artifact to the history, drops (player, row, col, val)
	// duplicates.
	Append(a Artifact)
	// Returns a map of all artifacts grouped by player, sorted by timestamp.
	Artifacts() map[string][]Artifact
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

func (h *history) Artifacts() map[string][]Artifact {
	h.lock.RLock()
	defer h.lock.RUnlock()

	artifacts := make(
		map[string][]Artifact,
		len(h.initial)+len(h.artifacts),
	)

	for _, a := range h.initial {
		player := a.Player()
		artifacts[player] = append(artifacts[player], a)
	}
	for _, a := range h.artifacts {
		player := a.Player()
		artifacts[player] = append(artifacts[player], a)
	}

	for player := range artifacts {
		subset := artifacts[player]
		sort.Slice(subset, func(i, j int) bool {
			return subset[i].Timestamp() < subset[j].Timestamp()
		})
		artifacts[player] = subset
	}

	return artifacts
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
