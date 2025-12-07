package socket

import "sudojo/core/event"

var (
	rateLimitMsg = []byte(`{"type":"system","error":"rate limit exceeded"}`)
	invalidMsg   = []byte(`{"type":"system","error":"invalid message format"}`)
)

type playerStatus struct {
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

type Message struct {
	Type      string          `json:"type"`
	Trace     string          `json:"trace_id,omitempty"`
	Error     string          `json:"error,omitempty"`
	Current   [][]int         `json:"current_state,omitempty"`
	Initial   [][]int         `json:"initial_state,omitempty"`
	Conflict  string          `json:"conflict,omitempty"`
	Row       *int            `json:"row,omitempty"`
	Column    *int            `json:"column,omitempty"`
	Value     *int            `json:"value,omitempty"`
	Players   []*playerStatus `json:"players,omitempty"`
	MaxPlayer int             `json:"max_player,omitempty"`
	Strict    *bool           `json:"strict,omitempty"`
}

func newMessage(e event.Event) *Message {
	msg := &Message{
		Type:  e.Type(),
		Trace: e.Trace(),
		Error: e.Error(),
	}
	if e.Payload() != nil {
		if e.Payload().Current() != nil {
			msg.Current = e.Payload().Current().Int()
		}
		if e.Payload().Initial() != nil {
			msg.Initial = e.Payload().Initial().Int()
		}
		msg.Conflict = e.Payload().Conflict()
		msg.Row = e.Payload().Row()
		msg.Column = e.Payload().Column()
		msg.Value = e.Payload().Value()
		msg.MaxPlayer = e.Payload().MaxPlayer()
		msg.Strict = e.Payload().Strict()
		if e.Payload().Players() != nil {
			msg.Players = []*playerStatus{}
			for _, p := range e.Payload().Players() {
				msg.Players = append(msg.Players, &playerStatus{
					Name:   p.Name(),
					Active: p.Active(),
				})
			}
		}
	}
	return msg
}
