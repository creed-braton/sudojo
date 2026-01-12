package session

import (
	"errors"
	"fmt"
	"log/slog"
	"sudojo/adp/database"
	"sudojo/adp/metrics"
	"sudojo/adp/socket"
	"sudojo/pkg/game"
	"sudojo/pkg/lobby"
	"sudojo/pkg/sudoku"
	"sync"
	"sync/atomic"
	"time"
)

type Service interface {
	Lobby() lobby.Lobby
	Close(code int, msg string)
	Player(token string) bool
	CreatePlayer(name string) (string, error)
	JoinPlayer(token string, client socket.Socket) error
	CreatedAt() int64
}

type service struct {
	lobby   lobby.Lobby
	db      database.Database
	logger  *slog.Logger
	metrics metrics.Metrics
	created int64
	clients map[string]socket.Socket
	lock    sync.Mutex
	once    sync.Once
	closed  atomic.Bool
}

var _ Service = &service{}

func New(
	lobby lobby.Lobby,
	db database.Database,
	logger *slog.Logger,
	metrics metrics.Metrics,
) *service {
	return &service{
		lobby:   lobby,
		db:      db,
		logger:  logger,
		metrics: metrics,
		created: time.Now().UTC().UnixNano(),
		clients: make(map[string]socket.Socket),
	}
}

func (s *service) Lobby() lobby.Lobby {
	return s.lobby
}

func (s *service) Close(code int, msg string) {
	s.once.Do(func() {
		s.closed.Store(true)
		s.lock.Lock()
		for _, c := range s.clients {
			c.Close(code, msg)
		}
		if err := s.db.UpdateLobby(s.lobby); err != nil {
			s.logger.Error(
				fmt.Sprintf("update lobby: %v", err),
			)
		}
		s.lock.Unlock()
	})
}

func (s *service) Player(token string) bool {
	return s.lobby.Player(token) != nil
}

func (s *service) CreatePlayer(name string) (string, error) {
	token, err := s.lobby.Create(name)
	if err != nil {
		return "", err
	}
	if err := s.db.InsertPlayer(s.lobby.Id(), token, name); err != nil {
		s.logger.Error(
			fmt.Sprintf("player create: %v", err),
			"player_token", token, "player_name", name,
		)
		return "", err
	}
	s.logger.Info("player created", "player_token", token)
	return token, nil
}

func (s *service) broadcast(msg *socket.Message) {
	for t, c := range s.clients {
		err := c.Send(msg)
		if err == socket.ErrClosed {
			delete(s.clients, t)
		}
	}
}

func (s *service) listen(token string, client socket.Socket) {
	client.Listen()

	s.lock.Lock()
	defer s.lock.Unlock()

	c, exist := s.clients[token]
	if !exist || c.Id() != client.Id() {
		return
	}

	delete(s.clients, token)

	update, err := s.lobby.Leave(token)
	if err != nil {
		s.logger.Error(
			fmt.Sprintf("player leave: %v", err),
			"player_token", token,
		)
	}

	players := []*socket.Player{}
	for _, p := range update {
		players = append(players, &socket.Player{
			Name: p.Name(), Active: p.Active(),
		})
	}
	msg := &socket.Message{Type: "leave", Players: players}
	s.broadcast(msg)
}

func (s *service) insert(token string, client socket.Socket, in *socket.Message) {
	if in.Row == nil || in.Column == nil || in.Value == nil {
		return
	}

	now := time.Now().UTC().UnixNano()
	out := &socket.Message{
		Type:   in.Type,
		Trace:  in.Trace,
		Row:    in.Row,
		Column: in.Column,
		Value:  in.Value,
	}

	current, err := s.lobby.Insert(*in.Row, *in.Column, *in.Value, token, now)
	if err == game.ErrIncorrect || err == game.ErrRowConflict ||
		err == game.ErrColConflict || err == game.ErrBoxConflict {
		out.Conflict = err.Error()
	} else if err != nil {
		out.Error = err.Error()
	}
	if current != nil {
		out.Current = current.Int()
	}

	if current != nil || len(out.Conflict) > 0 {
		s.lock.Lock()
		s.broadcast(out)
		s.lock.Unlock()
		return
	}
	if len(out.Error) > 0 {
		client.Send(out)
	}
}

func (s *service) ping(client socket.Socket, in *socket.Message) {
	if in.Row == nil || in.Column == nil {
		return
	}
	out := &socket.Message{
		Type:   in.Type,
		Trace:  in.Trace,
		Row:    in.Row,
		Column: in.Column,
	}

	if !sudoku.ValidBounds(*in.Row, *in.Column) {
		out.Error = game.ErrOutOfBounds.Error()
		client.Send(out)
		return
	}

	s.lock.Lock()
	s.broadcast(out)
	s.lock.Unlock()
}

func (s *service) state(client socket.Socket, in *socket.Message) {
	config := &socket.Config{
		Strict:  s.lobby.Config().Strict(),
		Ping:    s.lobby.Config().Ping(),
		Notes:   s.lobby.Config().Notes(),
		MaxSize: s.lobby.Config().MaxSize(),
	}

	players := []*socket.Player{}
	for _, p := range s.lobby.Players() {
		players = append(players, &socket.Player{
			Name: p.Name(), Active: p.Active(),
		})
	}

	out := &socket.Message{
		Type:    "state",
		Trace:   in.Trace,
		Config:  config,
		Players: players,
		Current: s.lobby.Game().Current().Int(),
		Initial: s.lobby.Game().Initial().Int(),
	}
	client.Send(out)
}

func (s *service) handle(token string, client socket.Socket) {
	for {
		in, err := client.Receive()
		if err != nil {
			break
		}

		switch in.Type {
		case "insert":
			s.insert(token, client, in)
		case "ping":
			s.ping(client, in)
		case "state":
			s.state(client, in)
		}
	}
}

func (s *service) JoinPlayer(token string, client socket.Socket) error {
	update, err := s.lobby.Join(token)
	if err != nil {
		client.Close(
			socket.CloseNotFound,
			"player not found in lobby",
		)
		return err
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	if s.closed.Load() {
		client.Close(socket.CloseStale, "lobby is stale")
		return errors.New("session is closed")
	}

	c, exist := s.clients[token]
	if exist {
		c.Close(socket.CloseTakeover, "connection taken over")
	}
	s.clients[token] = client

	players := []*socket.Player{}
	for _, p := range update {
		players = append(players, &socket.Player{
			Name: p.Name(), Active: p.Active(),
		})
	}
	msg := &socket.Message{Type: "join", Players: players}
	s.broadcast(msg)

	go s.listen(token, client)
	go s.handle(token, client)

	return nil
}

func (s *service) CreatedAt() int64 {
	return s.created
}
