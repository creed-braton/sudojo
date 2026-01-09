package message

import (
	"errors"
	"fmt"
)

const (
	InsertMsg = "insert" // Insert a value into the board at a specific cell.
	PingMsg   = "ping"   // Highlight a cell without inserting a value.
	StateMsg  = "state"  // Request or broadcast the current game state.
)

// Represents a client message used as input to proccess. Each message has a
// type identifying the kind of action (insert, ping, or state), a sender
// set after construction, and optional row, column, and value fields whose
// relevance depends on the message type.
type Message interface {
	// Type of the message.
	Type() string
	// Identifier of the player who sent the message. Required but must be set
	// via SetSender after initialization, not during construction.
	Sender() string
	// Sets the sender identifier of the message.
	SetSender(sender string) Message
	// Row of the targeted cell.
	Row() *int
	// Column of the targeted cell.
	Column() *int
	// Value to insert into the cell.
	Value() *int
	// Validates the message based on its type. Returns an error if the sender
	// is not set, if required fields are missing, or if forbidden fields are
	// present for the given type.
	Validate() error
}

type message struct {
	msgType, sender string
	row, col, val   *int
}

var _ Message = &message{}

// Creates a new message with the given type, row, column, and value.
func New(t string, row, col, val *int) *message {
	return &message{msgType: t, row: row, col: col, val: val}
}

func (m *message) Type() string {
	return m.msgType
}

func (m *message) Sender() string {
	return m.sender
}

func (m *message) SetSender(sender string) Message {
	m.sender = sender
	return m
}

func (m *message) Row() *int {
	return m.row
}

func (m *message) Column() *int {
	return m.col
}

func (m *message) Value() *int {
	return m.val
}

func (m *message) Validate() error {
	if len(m.sender) < 1 {
		return errors.New("sender could not be resolved")
	}

	switch m.msgType {
	case InsertMsg:
		if m.row == nil || m.col == nil || m.val == nil {
			return errors.New("insert message requires row, column, and value")
		}
	case PingMsg:
		if m.row == nil || m.col == nil {
			return errors.New("ping message requires row and column")
		}
		if m.val != nil {
			return errors.New("ping message must not have value")
		}
	case StateMsg:
		if m.row != nil || m.col != nil || m.val != nil {
			return errors.New("state message must not have row, column, or value")
		}
	default:
		return fmt.Errorf("unknown message type: %s", m.msgType)
	}
	return nil
}
