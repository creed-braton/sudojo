package message

import (
	"errors"
	"fmt"
)

const (
	InsertMsg = "insert"
	PingMsg   = "ping"
	StateMsg  = "state"
)

type Message interface {
	Type() string
	Sender() string
	SetSender(sender string) Message
	Row() *int
	Column() *int
	Value() *int
	Validate() error
}

type message struct {
	msgType, sender string
	row, col, val   *int
}

var _ Message = &message{}

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
