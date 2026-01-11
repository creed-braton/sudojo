package serial

import (
	"encoding/json"
	"sudojo/pkg/event"
	"sudojo/pkg/message"
)

type Serial interface {
	UnmarshalMsg(b []byte) (message.Message, error)
	MarshalEvent(e event.Event) ([]byte, error)
}

type serial struct{}

var _ Serial = &serial{}

func New() *serial {
	return &serial{}
}

func (s *serial) UnmarshalMsg(b []byte) (message.Message, error) {
	data := &struct {
		MsgType string `json:"type"`
		Trace   string `json:"trace"`
		Row     *int   `json:"row"`
		Col     *int   `json:"column"`
		Val     *int   `json:"value"`
	}{}
	if err := json.Unmarshal(b, data); err != nil {
		return nil, err
	}

	return message.New(
		data.MsgType,
		data.Row,
		data.Col,
		data.Val,
		data.Trace,
	), nil
}

type playerStatus struct {
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

type config struct {
	Strict  bool `json:"strict"`
	Ping    bool `json:"ping"`
	Notes   bool `json:"notes"`
	MaxSize int  `json:"max_player"`
}

type eventRes struct {
	EventType string          `json:"type"`
	Trace     string          `json:"trace"`
	Error     string          `json:"error,omitempty"`
	Conflict  string          `json:"conflict,omitempty"`
	Initial   [][]int         `json:"initial_board,omitempty"`
	Current   [][]int         `json:"current_board,omitempty"`
	Row       *int            `json:"row,omitempty"`
	Col       *int            `json:"column,omitempty"`
	Val       *int            `json:"value,omitempty"`
	Players   []*playerStatus `json:"players,omitempty"`
	Config    *config         `json:"config,omitempty"`
}

func (s *serial) MarshalEvent(e event.Event) ([]byte, error) {
	data := &eventRes{}
	data.EventType = e.Type()
	data.Trace = e.Trace()
	data.Error = e.Error()
	data.Conflict = e.Conflict()
	data.Initial = e.Initial().Int()
	data.Current = e.Current().Int()
	data.Row = e.Row()
	data.Col = e.Column()
	data.Val = e.Value()

	if e.Players() != nil {
		data.Players = []*playerStatus{}
		for _, p := range e.Players() {
			data.Players = append(data.Players, &playerStatus{
				Name:   p.Name(),
				Active: p.Active(),
			})
		}
	}

	if e.Config() != nil {
		data.Config = &config{
			Strict:  e.Config().Strict(),
			Ping:    e.Config().Ping(),
			Notes:   e.Config().Notes(),
			MaxSize: e.Config().MaxSize(),
		}
	}

	if b, err := json.Marshal(data); err != nil {
		return nil, err
	} else {
		return b, nil
	}
}
