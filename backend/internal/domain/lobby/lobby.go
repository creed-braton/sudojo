package lobby

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sudojo/internal/domain/game"
	"sudojo/internal/domain/sudoku"
	"sync"
	"time"
	"unicode"

	"github.com/google/uuid"
)

const (
	msgTypeMove  = "move"
	msgTypeState = "state"
	msgTypePing  = "ping"
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
	Initial  *sudoku.Sudoku `json:"initial_state,omitempty"`
	Current  *sudoku.Sudoku `json:"current_state,omitempty"`
	Cell     [2]int         `json:"cell,omitempty"`
	Error    string         `json:"error,omitempty"`
	Conflict string         `json:"conflict,omitempty"`
}

func (o *outbound) marshal() ([]byte, error) {
	b, err := json.Marshal(o)
	if err != nil {
		return nil, fmt.Errorf("failed serializing outbound message: %v", err)
	}
	return b, nil
}

type Log struct {
	Id     string `json:"id"`
	Lobby  string `json:"lobby_id"`
	Row    int    `json:"row"`
	Column int    `json:"column"`
	Value  int    `json:"value"`
	Player string `json:"player_token"`
	Time   int64  `json:"timestamp"`
}

type Player struct {
	Token  string `json:"token"`
	Name   string `json:"name"`
	Out    chan []byte
	active bool
}

type Lobby struct {
	Id        string             `json:"id"`
	Game      *game.Game         `json:"game"`
	Players   map[string]*Player `json:"players"`
	MaxPlayer int                `json:"max_player"`
	Strict    bool               `json:"strict"`
	Created   int64              `json:"created_at"`
	Finished  *int64             `json:"finished_at"`
	activity  int64
	lock      sync.RWMutex
	done      chan struct{}
	once      sync.Once
	logger    chan *Log
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
	l.activity = time.Now().UTC().UnixNano()
	for _, p := range l.Players {
		p.Out = make(chan []byte)
	}
}

func New(strict bool, maxPlayer int, logger chan *Log) *Lobby {
	seed := time.Now().UTC().UnixNano()
	l := &Lobby{
		Game:      game.New(seed),
		Players:   make(map[string]*Player),
		MaxPlayer: maxPlayer,
	}
	l.Id = uuid.NewString()
	l.Init(logger)
	l.Game.Start()
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

var (
	ErrNameTooLong = errors.New("player name too long")
	ErrInvalidChar = errors.New("player name contains invalid character")
	ErrLobbyClosed = errors.New("lobby is getting closed")
	ErrLobbyFull   = errors.New("lobby is already full")
)

func validName(name string) error {
	if len(name) > 12 {
		return ErrNameTooLong
	}
	for _, c := range name {
		if unicode.IsDigit(c) || unicode.IsLetter(c) {
			continue
		}
		if string(c) != "-" && string(c) != "_" {
			return ErrInvalidChar
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
		return nil, ErrLobbyClosed
	default:
	}

	if l.Finished != nil {
		return nil, game.ErrAlreadyFinished
	}

	if len(l.Players) >= l.MaxPlayer {
		return nil, ErrLobbyFull
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
		return ErrLobbyClosed
	default:
	}

	p.active = true
	l.activity = time.Now().UTC().UnixNano()
	return nil
}

func (l *Lobby) Leave(p *Player) {
	l.lock.Lock()
	p.active = false
	l.activity = time.Now().UTC().UnixNano()
	l.lock.Unlock()
}

func (l *Lobby) Idle(interval int64) bool {
	l.lock.Lock()
	defer l.lock.Unlock()

	now := time.Now().UTC().UnixNano()
	if now-l.activity < interval {
		return false
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

func (l *Lobby) strictMove(msg *inbound, player *Player) error {
	var conflict string
	var cell [2]int
	update, err := l.Game.Strict(msg.Row, msg.Column, msg.Value)
	incorrect := err != nil && err == game.ErrIncorrect

	if update != nil || incorrect {
		id := uuid.NewString()
		l.logger <- &Log{
			Id:     id,
			Lobby:  l.Id,
			Row:    msg.Row,
			Column: msg.Column,
			Value:  msg.Value,
			Player: player.Token,
			Time:   time.Now().UTC().UnixNano(),
		}
	}

	if err != nil {
		if incorrect {
			conflict = err.Error()
			cell = [2]int{msg.Row, msg.Column}
		} else {
			res, err := (&outbound{
				Error: err.Error(),
			}).marshal()
			if err != nil {
				return err
			}
			l.send(res, player)
			return nil
		}
	}

	res, err := (&outbound{
		Current:  update,
		Conflict: conflict,
		Cell:     cell,
	}).marshal()
	if err != nil {
		return err
	}
	l.broadcast(res)
	return nil
}

func (l *Lobby) laxMove(msg *inbound, player *Player) error {
	update, err := l.Game.Lax(msg.Row, msg.Column, msg.Value)
	invalid := err != nil && (err == game.ErrRowConflict || err == game.ErrColConflict || err == game.ErrBoxConflict)

	if update != nil || invalid {
		id := uuid.NewString()
		l.logger <- &Log{
			Id:     id,
			Lobby:  l.Id,
			Row:    msg.Row,
			Column: msg.Column,
			Value:  msg.Value,
			Player: player.Token,
			Time:   time.Now().UTC().UnixNano(),
		}
	}

	if err != nil {
		res, err := (&outbound{
			Conflict: err.Error(),
			Cell:     [2]int{msg.Row, msg.Column},
		}).marshal()
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

	return nil
}

func (l *Lobby) state(player *Player) error {
	res, err := (&outbound{
		Initial: l.Game.Initial,
		Current: l.Game.State(),
	}).marshal()
	if err != nil {
		return err
	}

	l.send(res, player)
	return nil
}

func (l *Lobby) ping(msg *inbound, player *Player) error {
	if !sudoku.ValidBounds(msg.Row, msg.Column) {
		res, err := (&outbound{
			Error: game.ErrOutOfBounds.Error(),
		}).marshal()
		if err != nil {
			return err
		}
		l.send(res, player)
		return nil
	}

	res, err := (&outbound{
		Cell: [2]int{msg.Row, msg.Column},
	}).marshal()
	if err != nil {
		return err
	}
	l.broadcast(res)
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
	case msgTypeMove:
		if l.Strict {
			return l.strictMove(req, player)
		} else {
			return l.laxMove(req, player)
		}
	case msgTypeState:
		return l.state(player)
	case msgTypePing:
		return l.ping(req, player)
	default:
		res, err := (&outbound{Error: "invalid json format"}).marshal()
		if err != nil {
			return err
		}
		l.send(res, player)
		return nil
	}
}
