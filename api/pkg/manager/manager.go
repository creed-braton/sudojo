package manager

import (
	"sudojo/pkg/event"
	"sudojo/pkg/lobby"
	"sync"
	"sync/atomic"
	"time"
)

// Holds lobby state and handles event flow logic.
type Manager interface {
	// Closes the central event buffer, causing the fanout to close
	// out all player buffers if Pump() is being called. Puts a
	// close event with provided reason code to notify players.
	Close(reason int)
	// Pumps event from the central buffer through the fanout to the
	// player buffers. Returns event.ErrClosedBuffer if the central
	// buffer has been closed.
	Pump() error
	// Returns the lobby state.
	Lobby() lobby.Lobby
	// Returns the Unix timestamp of the last recorded event.
	LastEvent() int64
	// Puts event into central event  buffer, propagating the event to
	// all players in the lobby. Returns event.ErrClosedBuffer if the
	// central buffer has been closed.
	Broadcast(e event.Event) error
	// Creates a new player with provided name in the lobby state.
	// Returns the generated token if successful, player.ErrInvalidChar
	// or player.ErrNameTooLong if the name is invalid, or
	// lobby.ErrLobbyFull  if lobby is already full.
	Create(name string) (string, error)
	// Creates and returns a player, ready with initialized and
	// registered event buffer if token is associated to a player in the
	// lobby state. Returns lobby.ErrPlayerNotFound if no player with the
	// token exists or event.ErrClosedBuffer if the central event buffer
	// of the lobby has been closed.
	Join(token string) (Player, error)
}

type manager struct {
	lobby     lobby.Lobby
	buffer    event.Buffer
	fanout    event.Fanout
	active    map[string]struct{}
	lastEvent atomic.Int64
	lock      sync.RWMutex
}

var _ Manager = &manager{}

// Updates the last event timestamp to the current time.
func (m *manager) updateEvent() {
	m.lastEvent.Store(time.Now().UTC().Unix())
}

// Returns a manager initialized with central event buffer and event
// fanout.
func New(lobby lobby.Lobby) *manager {
	buffer := event.NewBuffer()
	fanout := event.NewFanout(buffer)
	mng := &manager{lobby: lobby, buffer: buffer, fanout: fanout}
	mng.updateEvent()
	return mng
}

func (m *manager) Close(reason int) {
	event := event.New().SetType(event.CloseEvent).SetReason(reason)
	m.buffer.Send(event)
	m.buffer.Close()
}

func (m *manager) Pump() error {
	return m.fanout.Pump()
}

func (m *manager) Lobby() lobby.Lobby {
	return m.lobby
}

func (m *manager) LastEvent() int64 {
	return m.lastEvent.Load()
}

func (m *manager) Broadcast(e event.Event) error {
	return m.buffer.Send(e)
}

func (m *manager) Create(name string) (string, error) {
	return m.lobby.Create(name)
}

// Callback for player instance to call with their token when they wish
// to leave the lobby and transition into inactive state. Broadcasts a
// leave event notifying all players. Doesn't close the player's event
// buffer.
func (m *manager) leave(token string) {
	players, err := m.lobby.Leave(token)
	if err != nil {
		return
	}
	event := event.New().SetType(event.LeaveEvent).SetSender(token).
		SetPayload(event.NewPayload().SetPlayers(players))
	m.Broadcast(event)
}

func (m *manager) Join(token string) (Player, error) {
	players, err := m.lobby.Join(token)
	if err != nil {
		return nil, err
	}
	m.updateEvent()
	buffer := event.NewBuffer()
	m.fanout.Register(token, buffer)

	event := event.New().SetType(event.JoinEvent).SetSender(token).
		SetPayload(event.NewPayload().SetPlayers(players))

	if err := m.Broadcast(event); err != nil {
		return nil, err
	}

	return NewPlayer(
		token,
		m.lobby,
		buffer,
		m.Broadcast,
		m.leave,
		m.updateEvent,
	), nil
}
