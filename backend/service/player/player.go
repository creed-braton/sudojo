package player

import (
	"log/slog"
	"sudojo/adapter/socket"
	"sudojo/pkg/event"
	"sudojo/pkg/lobby"
)

type Service interface {
	Start()
}

type service struct {
	buffer    event.Buffer
	client    socket.Client
	lobby     lobby.Lobby
	broadcast func(e event.Event) error
	event     func()
	token     string
	logger    *slog.Logger
}

var _ Service = &service{}

func New(
	buffer event.Buffer,
	client socket.Client,
	lobby lobby.Lobby,
	broadcast func(e event.Event) error,
	event func(),
	token string,
	logger *slog.Logger,
) *service {
	return &service{
		buffer:    buffer,
		client:    client,
		lobby:     lobby,
		broadcast: broadcast,
		event:     event,
		token:     token,
		logger:    logger,
	}
}

func (s *service) handleInsert(row, col, val int, trace string) error {
	p, err := s.lobby.Insert(row, col, val)
	event := event.New().
		SetType(event.InsertEvent).
		SetSender(s.token).
		SetTrace(trace).
		SetPayload(p)

	if err != nil {
		return s.buffer.Send(event.SetError(err.Error()))
	}
	if p != nil {
		return s.broadcast(event)
	}
	return nil
}

func (s *service) handlePing(row, col int, trace string) error {
	p, err := s.lobby.Ping(row, col)
	event := event.New().
		SetType(event.PingEvent).
		SetSender(s.token).
		SetTrace(trace).
		SetPayload(p)

	if err != nil {
		return s.buffer.Send(event.SetError(err.Error()))
	}
	if p != nil {
		s.broadcast(event)
	}
	return nil
}

func (s *service) handleState() error {
	return s.buffer.Send(
		event.New().
			SetType(event.StateEvent).
			SetPayload(s.lobby.State()),
	)
}

func (s *service) handler() {
	defer func() {
		s.buffer.Close()
		p, err := s.lobby.Leave(s.token)
		if err != nil {
			s.logger.Error(err.Error())
			return
		}
		event := event.New().
			SetType(event.LeaveEvent).
			SetSender(s.token).
			SetPayload(p)
		if err := s.broadcast(event); err != nil {
			s.logger.Error(err.Error())
		}
		s.logger.Info("player left")
	}()

	for {
		msg, err := s.client.Receive()
		if err != nil {
			return
		}
		s.event()
		if msg.Type == event.InsertEvent {
			s.handleInsert(*msg.Row, *msg.Column, *msg.Value, msg.Trace)
		} else if msg.Type == event.PingEvent {
			s.handlePing(*msg.Row, *msg.Column, msg.Trace)
		} else {
			s.handleState()
		}
	}
}

func (s *service) Start() {
	go func() {
		_ = s.client.WritePump()
	}()
	go func() {
		_ = s.client.ReadPump()
	}()

	go func() {
		defer func() {
			s.buffer.Close()
			s.client.Close()
		}()

		for {
			e, err := s.buffer.Receive()
			if err != nil {
				return
			}
			if err := s.client.Send(e); err != nil {
				return
			}
		}
	}()

	go s.handler()
}
