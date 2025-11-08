package lobby

import (
	"errors"
	"sort"
	"sudojo/pkg/event"
	"sudojo/pkg/game"
	"sudojo/pkg/player"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

var (
	ErrLobbyFull      = errors.New("lobby is already full")
	ErrPlayerNotFound = errors.New("player not found in lobby")
)

type Lobby interface {
	Close()
	Id() string
	Game() game.Game
	Strict() bool
	MaxPlayer() int
	LastEvent() int64
	Join(token, name string) (player.Player, error)
	Leave(p player.Player) error
	State(p player.Player) error
	Insert(row, col, val int, p player.Player, trace string) error
	Ping(row, col int, trace string) error
}

type lobby struct {
	id        string
	game      game.Game
	players   map[string]*status
	strict    bool
	maxPlayer int
	lastEvent int64
	bus       event.EventBus
	fanout    event.Fanout
	lock      sync.RWMutex
}

func New(
	id string,
	game game.Game,
	players map[string]string,
	maxPlayer int,
	strict bool,
) *lobby {
	bus := event.NewEventBus()
	fanout := event.NewFanout(bus)

	statusMap := make(map[string]*status)
	for token, name := range players {
		statusMap[token] = &status{
			Name:   name,
			Active: false,
		}
	}

	return &lobby{
		id:        id,
		game:      game,
		players:   statusMap,
		maxPlayer: maxPlayer,
		strict:    strict,
		lastEvent: time.Now().UTC().Unix(),
		bus:       bus,
		fanout:    fanout,
	}
}

func Open(maxPlayer int, strict bool) *lobby {
	seed := time.Now().UTC().UnixNano()
	game := game.Generate(seed)
	defer game.Start()

	bus := event.NewEventBus()
	fanout := event.NewFanout(bus)

	return &lobby{
		id:        uuid.NewString(),
		game:      game,
		players:   make(map[string]*status),
		strict:    strict,
		lastEvent: time.Now().UTC().Unix(),
		bus:       bus,
		fanout:    fanout,
	}
}

func (l *lobby) Close() {
	l.bus.Close()
}

func (l *lobby) Id() string {
	return l.id
}

func (l *lobby) Game() game.Game {
	return l.game
}

func (l *lobby) Strict() bool {
	return l.strict
}

func (l *lobby) MaxPlayer() int {
	return l.maxPlayer
}

func (l *lobby) LastEvent() int64 {
	return atomic.LoadInt64(&l.lastEvent)
}

func (l *lobby) updateLastEvent() {
	atomic.StoreInt64(&l.lastEvent, time.Now().UTC().Unix())
}

func (l *lobby) Create(name string) (string, error) {
	l.updateLastEvent()
	if err := player.ValidName(name); err != nil {
		return "", err
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	if len(l.players) >= l.maxPlayer {
		return "", ErrLobbyFull
	}

	token := player.NewToken()
	l.players[token] = &status{
		Name:   name,
		Active: false,
	}

	return token, nil
}

func (l *lobby) sortPlayers() *playerState {
	tokens := make([]string, 0, len(l.players))
	for token := range l.players {
		tokens = append(tokens, token)
	}
	sort.Strings(tokens)

	var players playerState
	for _, token := range tokens {
		players = append(players, l.players[token])
	}

	return &players
}

func (l *lobby) Join(token, name string) (player.Player, error) {
	l.updateLastEvent()
	l.lock.Lock()
	defer l.lock.Unlock()

	s := l.players[token]
	if s == nil {
		return nil, ErrPlayerNotFound
	}

	s.Active = true
	bus := event.NewEventBus()
	l.fanout.Register(token, bus)
	p := player.New(token, name, bus)

	e := event.New("join", p.Token(), "", "", true, l.sortPlayers())
	err := l.bus.Send(e)

	return p, err
}

func (l *lobby) Leave(p player.Player) error {
	l.updateLastEvent()
	l.fanout.Deregister(p.Token())
	p.Close()

	l.lock.Lock()
	defer l.lock.Unlock()

	s := l.players[p.Token()]
	if s == nil {
		return ErrPlayerNotFound
	}
	s.Active = false

	e := event.New("leave", p.Token(), "", "", true, l.sortPlayers())

	return l.bus.Send(e)
}

func (l *lobby) State(p player.Player) error {
	l.updateLastEvent()
	payload := &gameState{Current: l.game.Current(), Initial: l.game.Initial()}
	e := event.New("state", p.Token(), "", "", false, payload)
	return p.Send(e)
}

func (l *lobby) Insert(row, col, val int, p player.Player, trace string) error {
	l.updateLastEvent()
	return nil
}

func (l *lobby) Ping(row, col int, p player.Player, trace string) error {
	l.updateLastEvent()
	return nil
}
