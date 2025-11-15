package lobby

import (
	"sudojo/pkg/event"
	"sudojo/pkg/game"
	"sudojo/pkg/lobby"
	"sudojo/pkg/player"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	Shutdown()
}

type service struct {
	lobby lobby.Lobby
	done  chan struct{}
	once  sync.Once
}

func (s *service) pump() {
	select {
	case <-s.done:
		return
	default:
		s.lobby.Pump()
	}
}

func New() *service {
	id := uuid.NewString()
	bus := event.NewEventBus()
	s := &service{lobby: lobby.New(
		id,
		false,
		game.Generate(time.Now().UTC().UnixNano()),
		bus,
		player.NewPlayerPool(make(map[string]string), 8, bus),
	)}
	go s.pump()

	return s
}

func (s *service) Shutdown() {
	s.once.Do(func() {
		close(s.done)
		s.lobby.Close()
	})
}
