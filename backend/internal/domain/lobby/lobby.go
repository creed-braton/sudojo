package lobby

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrLobbyClosed = errors.New("lobby is getting closed")
)

type payload interface{}

type Event struct {
	Sender   string
	Receiver string
	Trace    string  `json:"trace_id"`
	Type     string  `json:"event_type"`
	Error    string  `json:"error,omitempty"`
	Payload  payload `json:"payload"`
}

type Lobby struct {
	Id            string
	GameManager   *GameManager
	PlayerManager *PlayerManager
	Events        chan *Event
	lock          sync.RWMutex
	done          chan struct{}
	once          sync.Once
}

func (l *Lobby) Init() {
	l.Events = make(chan *Event, 256)
	l.done = make(chan struct{})
	l.once = sync.Once{}
}

func New(strict bool, maxPlayer int) *Lobby {
	seed := time.Now().UTC().UnixNano()
	l := &Lobby{
		Id:            uuid.NewString(),
		GameManager:   newGameManager(seed, strict),
		PlayerManager: newPlayerManager(maxPlayer),
	}
	l.Init()
	return l
}

func (l *Lobby) close() {
	l.once.Do(func() {
		l.lock.Lock()
		close(l.done)
		l.lock.Unlock()
	})
}

func (l *Lobby) Create(name string) (string, error) {
	return l.PlayerManager.create(name)
}

func (l *Lobby) Join(trace, token string) error {
	p, err := l.PlayerManager.join(token)
	if err != nil {
		return err
	}
	l.Events <- &Event{
		Sender:   token,
		Receiver: "",
		Trace:    trace,
		Type:     "join",
		Payload:  p,
	}
	return nil
}

func (l *Lobby) Leave(trace, token string) error {
	p, err := l.PlayerManager.leave(token)
	if err != nil {
		return err
	}
	l.Events <- &Event{
		Sender:   token,
		Receiver: "",
		Trace:    trace,
		Type:     "leave",
		Payload:  p,
	}
	return nil
}

func (l *Lobby) Insert(trace, token string, row, col, val int) {
	p, err := l.GameManager.insert(row, col, val)
	if p == nil && err == nil {
		return
	}

	event := &Event{Sender: token, Trace: trace, Type: "insert"}
	if err != nil {
		event.Error = err.Error()
	}
	if p.Current == nil {
		event.Receiver = token
	}

	l.Events <- event
}

func (l *Lobby) State(trace, token string) {
	p := l.GameManager.state()
	if p == nil {
		return
	}

	l.Events <- &Event{
		Sender:   token,
		Receiver: token,
		Trace:    trace,
		Type:     "state",
		Payload:  p,
	}
}

func (l *Lobby) Ping(trace, token string, row, col int) {
	event := &Event{Sender: token, Trace: trace, Type: "ping"}

	p, err := l.GameManager.ping(row, col)
	if err != nil {
		event.Receiver = token
		event.Error = err.Error()
	} else if p != nil {
		event.Payload = p
	}

	l.Events <- event
}
