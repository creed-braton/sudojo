package lobby

import (
	"sudojo/pkg/event"
	"sudojo/pkg/game"
	"sudojo/pkg/player"
	"time"
)

type Lobby interface {
	Close()
	Pump()
	Id() string
	LastEvent() int64
}

type lobby struct {
	id        string
	lastEvent int64
	bus       event.EventBus
	strict    bool
	pool      player.Pool
}

var _ Lobby = &lobby{}

func New(
	id string,
	game game.Game,
	players map[string]string,
	maxPlayer int,
	strict bool,
) *lobby {
	bus := event.NewEventBus()

	return &lobby{
		id:        id,
		lastEvent: time.Now().UTC().Unix(),
		bus:       bus,
		strict:    strict,
		pool:      player.NewPlayerPool(players, maxPlayer, bus),
	}
}

func (l *lobby) Close() {
	l.bus.Close()
}

func (l *lobby) Pump() {
	l.bus.Pump()
}

func (l *lobby) Id() string {
	return l.id
}

func (l *lobby) LastEvent() int64 {
	return l.lastEvent
}
