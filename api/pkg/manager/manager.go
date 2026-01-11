package manager

import (
	"sudojo/pkg/event"
	"sudojo/pkg/game"
	"sudojo/pkg/lobby"
	"sudojo/pkg/message"
	"sudojo/pkg/sudoku"
	"time"
)

// Orchestrates game sessions by coordinating player connections, event
// distribution, and message processing between clients and the lobby.
type Manager interface {
	// Returns the lobby associated with the manager.
	Lobby() lobby.Lobby
	// Activates the player associated with the token in the lobby and
	// registers a new event buffer for the player. Broadcasts a join
	// event to all connected players. Returns the event buffer for
	// receiving events. Returns lobby.ErrPlayerNotFound if no player
	// is associated with the token, game.ErrFinished if the game is
	// already finished, or event.ErrHubClosed if the hub has been closed.
	Join(token string) (event.Buffer, error)
	// Deactivates the player associated with the token in the lobby
	// and deregisters his event buffer from the hub. Broadcasts a leave
	// event to all connected players. Returns lobby.ErrPlayerNotFound
	// if no player is associated with the token.
	Leave(token string) error
	// Routes the message to the appropriate handler based on message
	// type. Insert messages validate input and broadcast board updates
	// or send error events to the sender. Ping messages broadcast cell
	// highlights or send error events for invalid bounds. State messages
	// send current game state to the sender.
	Process(msg message.Message)
}

type manager struct {
	lobby lobby.Lobby
	hub   event.Hub
}

// Creates a new manager with provided lobby and hub.
func New(lobby lobby.Lobby, hub event.Hub) *manager {
	return &manager{lobby: lobby, hub: hub}
}

var _ Manager = &manager{}

func (m *manager) Lobby() lobby.Lobby {
	return m.lobby
}

func (m *manager) Join(token string) (event.Buffer, error) {
	now := time.Now().UTC().UnixNano()
	players, err := m.lobby.Join(token)
	if err != nil {
		return nil, err
	}

	buffer := event.NewBuffer(128)
	if err := m.hub.Register(token, buffer); err != nil {
		return nil, err
	}

	event := event.New(event.JoinEvent, now, "").SetPlayers(players)
	m.hub.Broadcast(event)

	return buffer, nil
}

func (m *manager) Leave(token string) error {
	now := time.Now().UTC().UnixNano()
	players, err := m.lobby.Leave(token)
	if err != nil {
		return err
	}

	m.hub.Deregister(token)
	event := event.New(event.LeaveEvent, now, "").SetPlayers(players)
	m.hub.Broadcast(event)

	return nil
}

func (m *manager) insert(msg message.Message) {
	now := time.Now().UTC().UnixNano()
	event := event.New(event.InsertEvent, now, msg.Trace())

	current, err := m.lobby.Insert(*msg.Row(), *msg.Column(), *msg.Value(), msg.Sender(), now)
	event.SetCurrent(current)

	if err != nil && err != game.ErrIncorrect && err != game.ErrRowConflict &&
		err != game.ErrColConflict && err != game.ErrBoxConflict {
		event.SetError(err.Error())
		m.hub.Send(msg.Sender(), event)
		return
	}

	if err != nil {
		event.SetConflict(err.Error())
	}

	m.hub.Broadcast(event)
}

func (m *manager) ping(msg message.Message) {
	now, row, col := time.Now().UTC().UnixNano(), *msg.Row(), *msg.Column()
	event := event.New(event.PingEvent, now, msg.Trace()).
		SetRow(row).SetColumn(col)

	if !sudoku.ValidBounds(row, col) {
		event.SetError(game.ErrOutOfBounds.Error())
		m.hub.Send(msg.Sender(), event)
		return
	}

	m.hub.Broadcast(event)
}

func (m *manager) state(msg message.Message) {
	now := time.Now().UTC().UnixNano()
	event := event.New(event.StateEvent, now, msg.Trace()).
		SetCurrent(m.lobby.Game().Current()).
		SetInitial(m.lobby.Game().Initial()).
		SetPlayers(m.lobby.Players()).
		SetConfig(m.lobby.Config())

	m.hub.Send(msg.Sender(), event)
}

func (m *manager) Process(msg message.Message) {
	if msg.Type() == message.InsertMsg {
		m.insert(msg)
	}
	if msg.Type() == message.StateMsg {
		m.state(msg)
	}
	if msg.Type() == message.PingMsg && m.lobby.Config().Ping() {
		m.ping(msg)
	}
}
