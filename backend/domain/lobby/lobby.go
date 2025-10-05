package lobby

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sudojo/domain/game"
	"sudojo/domain/sudoku"
	"sync"
)

const (
	msgTypeMove  = "move"
	msgTypeState = "state"
	maxPlayer    = 8
)

type inbound struct {
	Type   string `json:"type"`
	Row    int    `json:"row"`
	Column int    `json:"column"`
	Value  int    `json:"value"`
}

func unmarshal(data []byte) (*inbound, error) {
	in := &inbound{}
	if err := json.Unmarshal(data, in); err != nil {
		return nil, fmt.Errorf("failed deserializing inbound message: %v", err)
	}
	return in, nil
}

type outbound struct {
	Initial *sudoku.Sudoku `json:"initial_state,omitempty"`
	Current *sudoku.Sudoku `json:"current_state,omitempty"`
	Error   string         `json:"error,omitempty"`
}

func (o *outbound) marshal() ([]byte, error) {
	data, err := json.Marshal(o)
	if err != nil {
		return nil, fmt.Errorf("failed serializing outbound message: %v", err)
	}
	return data, nil
}

type Player struct {
	Token  string
	name   string
	Out    chan []byte
	active bool
}

type Lobby struct {
	Id      string
	game    *game.Game
	players map[string]*Player
	lock    sync.RWMutex
	done    chan struct{}
	once    sync.Once
}

func newToken() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func New() *Lobby {
	l := &Lobby{
		game:    game.New(),
		players: make(map[string]*Player),
		done:    make(chan struct{}),
	}
	l.Id = l.game.Id.String()
	return l
}

func (l *Lobby) close() {
	l.once.Do(func() {
		close(l.done)
		for _, p := range l.players {
			close(p.Out)
		}
	})
}

func (l *Lobby) Player(token string) *Player {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.players[token]
}

func (l *Lobby) Join(name string) (*Player, error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	select {
	case <-l.done:
		return nil, errors.New("lobby is getting closed")
	default:
	}

	if len(l.players) >= maxPlayer {
		return nil, errors.New("lobby is already full")
	}

	p := &Player{
		Token:  newToken(),
		name:   name,
		Out:    make(chan []byte, 256),
		active: true,
	}
	l.players[p.Token] = p

	return p, nil
}

func (l *Lobby) Rejoin(p *Player) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	select {
	case <-l.done:
		return errors.New("lobby is getting closed")
	default:
	}

	p.active = true
	return nil
}

func (l *Lobby) Leave(p *Player) {
	l.lock.Lock()
	p.active = false
	l.lock.Unlock()
}

func (l *Lobby) Idle() bool {
	l.lock.Lock()
	defer l.lock.Unlock()

	for _, p := range l.players {
		if p.active {
			return false
		}
	}

	l.close()
	return true
}

func (l *Lobby) send(msg []byte, p *Player) {
	l.lock.RLock()
	defer l.lock.RUnlock()

	// check if lobby is not getting closed and channel is still safe to use
	select {
	case <-l.done:
		return
	default:
		p.Out <- msg
	}
}

func (l *Lobby) broadcast(msg []byte) {
	l.lock.RLock()
	for _, p := range l.players {
		l.send(msg, p)
	}
	l.lock.RUnlock()
}

func (l *Lobby) move(msg *inbound, player *Player) error {
	update, err := l.game.Insert(msg.Row, msg.Column, msg.Value)
	if err != nil {
		res, err := (&outbound{Error: err.Error()}).marshal()
		if err != nil {
			return err
		}
		l.send(res, player)
	}

	if update != nil {
		res, err := (&outbound{Current: update}).marshal()
		if err != nil {
			return err
		}
		l.broadcast(res)
	}

	if l.game.Finished != nil {
		l.lock.Lock()
		l.close()
		l.lock.Unlock()
	}

	return nil
}

func (l *Lobby) state(player *Player) error {
	initial, current := l.game.State()

	res, err := (&outbound{
		Initial: initial,
		Current: current,
	}).marshal()
	if err != nil {
		return err
	}

	l.send(res, player)
	return nil
}

func (l *Lobby) Process(msg []byte, player *Player) error {
	req, err := unmarshal(msg)
	if err != nil {
		res, err := (&outbound{Error: "invalid json format"}).marshal()
		if err != nil {
			return err
		}
		l.send(res, player)
		return nil
	}

	switch req.Type {
	case msgTypeState:
		return l.state(player)
	case msgTypeMove:
		return l.move(req, player)
	default:
		res, err := (&outbound{Error: "invalid json format"}).marshal()
		if err != nil {
			return err
		}
		l.send(res, player)
		return nil
	}
}
