package ctrl

import (
	"sudojo/core/event"
	"sudojo/pkg/lobby"
	"sync"
	"sync/atomic"
	"time"
)

// Holds lobby state and handles event flow logic.
type Controller interface {
	// Closes the central event buffer, causing the fanout to close
	// out all player buffers if Pump() is being called.
	Close()
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
	// Returns the generated token if successful, ErrInvalidChar or
	// ErrNameTooLong if the name is invalid, or ErrLobbyFull if lobby
	// is already full.
	Create(name string) (string, error)
	// Creates and returns a player, ready with initialized and
	// registered event buffer if token is associated to a player in the
	// lobby state. Returns ErrPlayerNotFound if no player with the token
	// exists or event.ErrClosedBuffer if the central event buffer of the
	// lobby has been closed.
	Join(token string) (Player, error)
}

type controller struct {
	lobby     lobby.Lobby
	buffer    event.Buffer
	fanout    event.Fanout
	active    map[string]struct{}
	lastEvent atomic.Int64
	lock      sync.RWMutex
}

var _ Controller = &controller{}

// Returns a controller initialized with central event buffer and event
// fanout.
func New(lobby lobby.Lobby) *controller {

	buffer := event.NewBuffer()
	fanout := event.NewFanout(buffer)
	ctrl := &controller{lobby: lobby, buffer: buffer, fanout: fanout}
	ctrl.updateEvent()
	return ctrl
}

func (c *controller) Close() {
	c.buffer.Close()
}

func (c *controller) Pump() error {
	return c.fanout.Pump()
}

func (c *controller) Lobby() lobby.Lobby {
	return c.lobby
}

// Updates the last event timestamp to the current time.
func (c *controller) updateEvent() {
	c.lastEvent.Store(time.Now().UTC().Unix())
}

func (c *controller) LastEvent() int64 {
	return c.lastEvent.Load()
}

func (c *controller) Broadcast(e event.Event) error {
	return c.buffer.Send(e)
}

func (c *controller) Create(name string) (string, error) {
	return c.lobby.Create(name)
}

// Callback for player instance to call with their token when they wish
// to leave the lobby and transition into inactive state. Broadcasts a
// leave event notifying all players. Doesn't close the player's event
// buffer.
func (c *controller) leave(token string) {
	event := event.New().SetType(event.LeaveEvent).SetSender(token).
		SetPayload(event.NewPayload().SetPlayers(c.lobby.Leave(token)))
	c.Broadcast(event)
}

func (c *controller) Join(token string) (Player, error) {
	players, err := c.lobby.Join(token)
	if err != nil {
		return nil, err
	}
	c.updateEvent()
	buffer := event.NewBuffer()
	c.fanout.Register(token, buffer)

	event := event.New().SetType(event.JoinEvent).SetSender(token).
		SetPayload(event.NewPayload().SetPlayers(players))

	if err := c.Broadcast(event); err != nil {
		return nil, err
	}

	return NewPlayer(
		token,
		c.lobby,
		buffer,
		c.Broadcast,
		c.leave,
		c.updateEvent,
	), nil
}
