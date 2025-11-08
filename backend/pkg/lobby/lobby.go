package lobby

import (
	"sudojo/pkg/event"
	"sudojo/pkg/game"
	"sudojo/pkg/lobby/manager"
	"sudojo/pkg/player"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type Lobby interface {
	Close()
	Poll() error
	Id() string
	LastEvent() int64
	manager.Player
	manager.Game
}

type lobby struct {
	id            string
	bus           event.EventBus
	poll          func() error
	lastEvent     int64
	playerManager manager.Player
	gameManager   manager.Game
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
	fanout := event.NewFanout(bus)

	return &lobby{
		id:            id,
		bus:           bus,
		poll:          fanout.Poll,
		lastEvent:     time.Now().UTC().Unix(),
		playerManager: manager.NewPlayerManager(players, maxPlayer, bus, fanout),
		gameManager:   manager.NewGameManager(game, bus, strict),
	}
}

func Open(maxPlayer int, strict bool) *lobby {
	seed := time.Now().UTC().UnixNano()
	game := game.Generate(seed)
	players := make(map[string]string)

	bus := event.NewEventBus()
	fanout := event.NewFanout(bus)

	return &lobby{
		id:            uuid.NewString(),
		bus:           bus,
		poll:          fanout.Poll,
		lastEvent:     time.Now().UTC().Unix(),
		playerManager: manager.NewPlayerManager(players, maxPlayer, bus, fanout),
		gameManager:   manager.NewGameManager(game, bus, strict),
	}
}

func (l *lobby) Close() {
	l.bus.Close()
}

func (l *lobby) Poll() error {
	return l.poll()
}

func (l *lobby) Id() string {
	return l.id
}

func (l *lobby) LastEvent() int64 {
	return atomic.LoadInt64(&l.lastEvent)
}

func (l *lobby) updateLastEvent() {
	atomic.StoreInt64(&l.lastEvent, time.Now().UTC().Unix())
}

func (l *lobby) Game() game.Game {
	return l.gameManager.Game()
}

func (l *lobby) Strict() bool {
	return l.gameManager.Strict()
}

func (l *lobby) MaxPlayer() int {
	return l.playerManager.MaxPlayer()
}

func (l *lobby) Create(name string) (string, error) {
	l.updateLastEvent()
	return l.playerManager.Create(name)
}

func (l *lobby) Join(token, name string) (player.Player, error) {
	l.updateLastEvent()
	return l.playerManager.Join(token, name)
}

func (l *lobby) Leave(p player.Player) error {
	l.updateLastEvent()
	return l.playerManager.Leave(p)
}

func (l *lobby) State(p player.Player) error {
	l.updateLastEvent()
	return l.gameManager.State(p)
}

func (l *lobby) Insert(row, col, val int, p player.Player, trace string) error {
	l.updateLastEvent()
	return l.gameManager.Insert(row, col, val, p, trace)
}

func (l *lobby) Ping(row, col int, p player.Player, trace string) error {
	l.updateLastEvent()
	return l.gameManager.Ping(row, col, p, trace)
}
