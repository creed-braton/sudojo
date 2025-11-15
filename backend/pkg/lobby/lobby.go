package lobby

import (
	"sudojo/pkg/event"
	"sudojo/pkg/game"
	"sudojo/pkg/player"
	"sudojo/pkg/sudoku"
	"sync/atomic"
	"time"
)

type Lobby interface {
	Close()
	Pump()
	Id() string
	LastEvent() int64
	Create(name string) (string, error)
	Join(token string, pub, sub event.EventChan) error
	Leave(token string, pub event.EventChan) error
	Insert(token, trace string, pub event.EventChan, row, col, val int)
	Ping(token, trace string, pub event.EventChan, row, col int)
	State(token string, pub event.EventChan)
}

type lobby struct {
	id        string
	lastEvent int64
	strict    bool
	game      game.Game
	bus       event.EventBus
	pool      player.Pool
}

var _ Lobby = &lobby{}

func New(
	id string,
	strict bool,
	game game.Game,
	bus event.EventBus,
	pool player.Pool,
) *lobby {
	return &lobby{
		id:        id,
		lastEvent: time.Now().UTC().Unix(),
		strict:    strict,
		game:      game,
		bus:       bus,
		pool:      pool,
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
	return atomic.LoadInt64(&l.lastEvent)
}

func (l *lobby) updateLastEvent() {
	atomic.StoreInt64(&l.lastEvent, time.Now().UTC().Unix())
}

func (l *lobby) Create(name string) (string, error) {
	return l.pool.Create(name)
}

func (l *lobby) Join(token string, pub, sub event.EventChan) error {
	players, err := l.pool.Join(token, pub, sub)
	if err != nil {
		return err
	}

	playerStatus := []*event.PlayerStatus{}
	for _, p := range players {
		playerStatus = append(playerStatus, &event.PlayerStatus{
			Name:   p.Name(),
			Active: p.Active(),
		})
	}

	e := event.New(
		event.JoinEvent, token, "", "", true,
		event.NewPlayerPayload(playerStatus),
	)
	pub.Send(e)
	l.updateLastEvent()

	return nil
}

func (l *lobby) Leave(token string, pub event.EventChan) error {
	players, err := l.pool.Leave(token)
	if err != nil {
		return err
	}

	playerStatus := []*event.PlayerStatus{}
	for _, p := range players {
		playerStatus = append(playerStatus, &event.PlayerStatus{
			Name:   p.Name(),
			Active: p.Active(),
		})
	}

	e := event.New(
		event.LeaveEvent, token, "", "", true,
		event.NewPlayerPayload(playerStatus),
	)
	pub.Send(e)
	l.updateLastEvent()

	return nil
}

func (l *lobby) strictInsert(token, trace string, pub event.EventChan, row, col, val int) {
	current, err := l.game.Strict(row, col, val)
	conflict := ""
	if err != nil {
		if err != game.ErrIncorrect {
			pub.Send(event.New(
				event.InsertEvent,
				token, trace, err.Error(),
				false, nil,
			))
		} else {
			conflict = err.Error()
		}
	}
	pub.Send(event.New(
		event.InsertEvent,
		token, trace, "",
		true, event.NewInsertPayload(row, col, val, current, conflict),
	))
}

func (l *lobby) laxInsert(token, trace string, pub event.EventChan, row, col, val int) {
	current, err := l.game.Lax(row, col, val)
	conflict := ""
	if err != nil {
		if err == game.ErrRowConflict ||
			err == game.ErrColConflict ||
			err == game.ErrBoxConflict {
			conflict = err.Error()
		} else {
			pub.Send(event.New(
				event.InsertEvent,
				token, trace, err.Error(),
				false, nil,
			))
		}
	}
	pub.Send(event.New(
		event.InsertEvent,
		token, trace, "",
		true, event.NewInsertPayload(row, col, val, current, conflict),
	))
}

func (l *lobby) Insert(token, trace string, pub event.EventChan, row, col, val int) {
	l.updateLastEvent()

	if l.strict {
		l.strictInsert(token, trace, pub, row, col, val)
	} else {
		l.laxInsert(token, trace, pub, row, col, val)
	}
}

func (l *lobby) Ping(token, trace string, pub event.EventChan, row, col int) {
	l.updateLastEvent()

	if !sudoku.ValidBounds(row, col) {
		pub.Send(event.New(
			event.PingEvent,
			token, trace,
			game.ErrOutOfBounds.Error(),
			false, nil,
		))
	}

	pub.Send(event.New(
		event.PingEvent,
		token, trace, "",
		true, event.NewPingPayload(row, col),
	))
}

func (l *lobby) State(token string, pub event.EventChan) {
	l.updateLastEvent()

	pub.Send(event.New(
		event.StateEvent,
		token, "", "",
		false, event.NewStatePayload(l.game.Current(), l.game.Initial()),
	))
}
