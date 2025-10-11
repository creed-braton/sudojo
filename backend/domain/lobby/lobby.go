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
	"time"
	"unicode"
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

func unmarshal(b []byte) (*inbound, error) {
	in := &inbound{}
	if err := json.Unmarshal(b, in); err != nil {
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
	b, err := json.Marshal(o)
	if err != nil {
		return nil, fmt.Errorf("failed serializing outbound message: %v", err)
	}
	return b, nil
}

type Player struct {
	Token  string
	Name   string
	Out    chan []byte
	active bool
}

type Lobby struct {
	Id      string
	Game    *game.Game
	Players map[string]*Player
	lock    sync.RWMutex
	done    chan struct{}
	once    sync.Once
	logger  chan *Log
}

func newToken() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func (l *Lobby) Init(logger chan *Log) {
	l.logger = logger
	l.done = make(chan struct{})
}

func New(logger chan *Log) *Lobby {
	l := &Lobby{
		Game:    game.New(),
		Players: make(map[string]*Player),
	}
	l.Id = l.Game.Id.String()
	l.Init(logger)
	return l
}

func (l *Lobby) close() {
	l.once.Do(func() {
		close(l.done)
		for _, p := range l.Players {
			close(p.Out)
		}
	})
}

func (l *Lobby) Player(token string) *Player {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.Players[token]
}

func validName(name string) error {
	if len(name) > 12 {
		return errors.New("player name too long")
	}
	for _, c := range name {
		if unicode.IsDigit(c) || unicode.IsLetter(c) {
			continue
		}
		if string(c) != "-" && string(c) != "_" {
			return errors.New("player name contains invalid character")
		}
	}
	return nil
}

func (l *Lobby) Create(name string) (*Player, error) {
	if err := validName(name); err != nil {
		return nil, err
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	select {
	case <-l.done:
		return nil, errors.New("lobby is getting closed")
	default:
	}

	if len(l.Players) >= maxPlayer {
		return nil, errors.New("lobby is already full")
	}

	p := &Player{
		Token:  newToken(),
		Name:   name,
		Out:    make(chan []byte, 256),
		active: false,
	}
	l.Players[p.Token] = p

	return p, nil
}

func (l *Lobby) Join(p *Player) error {
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

	for _, p := range l.Players {
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
	for _, p := range l.Players {
		l.send(msg, p)
	}
	l.lock.RUnlock()
}

func (l *Lobby) move(msg *inbound, player *Player) error {
	update, err := l.Game.Insert(msg.Row, msg.Column, msg.Value)
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

	if update != nil || err != nil {
		l.logger <- &Log{
			LobbyId: l.Id,
			Row:     msg.Row,
			Column:  msg.Column,
			Value:   msg.Value,
			Player:  player.Token,
			Time:    time.Now().UTC().UnixNano(),
		}
	}

	if l.Game.Finished != nil {
		l.lock.Lock()
		l.close()
		l.lock.Unlock()
	}

	return nil
}

func (l *Lobby) state(player *Player) error {
	initial, current := l.Game.State()

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
