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
	msgTypeMove  = "move"  // Message type for player moves
	msgTypeState = "state" // Message type for game state requests
	msgTypePing  = "ping"  // Message type for cell highlighting
)

// Represents incoming messages from players containing move data,
// state requests, or ping commands for cell highlighting.
type inbound struct {
	Type   string `json:"type"`   // message type (move, state, ping)
	Row    int    `json:"row"`    // target row coordinate
	Column int    `json:"column"` // target column coordinate
	Value  int    `json:"value"`  // value to place in cell
}

// Deserializes bytes into an inbound message structure.
func unmarshal(b []byte) (*inbound, error) {
	in := &inbound{}
	if err := json.Unmarshal(b, in); err != nil {
		return nil, fmt.Errorf("failed deserializing inbound message: %v", err)
	}
	return in, nil
}

// Represents outgoing messages sent to players containing game state
// updates, error messages, or conflict notifications.
type outbound struct {
	Initial  *sudoku.Sudoku `json:"initial_state,omitempty"` // initial puzzle state
	Current  *sudoku.Sudoku `json:"current_state,omitempty"` // current game state
	Cell     [2]int         `json:"cell,omitempty"`          // highlighted cell coordinates
	Error    string         `json:"error,omitempty"`         // error message
	Conflict string         `json:"conflict,omitempty"`      // conflict description
}

// Serializes the outbound message structure into bytes.
func (o *outbound) marshal() ([]byte, error) {
	b, err := json.Marshal(o)
	if err != nil {
		return nil, fmt.Errorf("failed serializing outbound message: %v", err)
	}
	return b, nil
}

// Represents a game move log entry containing player action details and
// metadata for audit and replay purposes.
type Log struct {
	Id     string `json:"id"`           // unique log entry identifier
	Lobby  string `json:"lobby_id"`     // lobby identifier
	Row    int    `json:"row"`          // row coordinate
	Column int    `json:"column"`       // column coordinate
	Value  int    `json:"value"`        // placed value
	Player string `json:"player_token"` // player token who made the move
	Time   int64  `json:"timestamp"`    // timestamp when move was made
}

// Represents a player in the lobby with communication channel and status tracking.
type Player struct {
	Token  string      `json:"token"` // unique player identifier
	Name   string      `json:"name"`  // player display name
	Out    chan []byte // outbound message channel
	active bool        // connection status
}

// Represents a multiplayer Sudoku game lobby that manages players, game state,
// and message routing. Provides thread-safe operations for player management
// and game interaction.
type Lobby struct {
	Id        string             `json:"id"`          // unique lobby identifier
	Game      *game.Game         `json:"game"`        // underlying Sudoku game
	Players   map[string]*Player `json:"players"`     // players indexed by token
	MaxPlayer int                `json:"max_player"`  // maximum allowed players
	Strict    bool               `json:"strict"`      // validation mode (strict/lax)
	Created   int64              `json:"created_at"`  // creation timestamp
	Finished  *int64             `json:"finished_at"` // completion timestamp, nil if ongoing
	activity  int64              // last activity timestamp for idle detection
	lock      sync.RWMutex       // mutex for thread-safe operations
	done      chan struct{}      // channel to signal lobby closure
	once      sync.Once          // ensures cleanup happens only once
	logger    chan *Log          // channel for sending logs
}

// Generates a cryptographically secure random token for player identification.
func newToken() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

// Initializes lobby channels and sets up player communication channels.
func (l *Lobby) Init(logger chan *Log) {
	l.logger = logger
	l.lock = sync.RWMutex{}
	l.done = make(chan struct{})
	l.once = sync.Once{}
	l.activity = time.Now().UTC().UnixNano()
	for _, p := range l.Players {
		p.Out = make(chan []byte, 256)
	}
}

// Creates a new lobby with specified validation mode, player limit, and logger.
// Generates a new Sudoku game and starts it immediately.
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

// Closes the lobby by signaling shutdown and closing all player channels.
func (l *Lobby) close() {
	l.once.Do(func() {
		close(l.done)
		for _, p := range l.Players {
			close(p.Out)
		}
	})
}

// Returns the player associated with the given token in a thread-safe manner.
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

// Validates player name length and character restrictions.
func validName(name string) error {
	if len(name) > 16 {
		return ErrNameTooLong
	}
	for _, c := range name {
		if unicode.IsDigit(c) || unicode.IsLetter(c) {
			continue
		}
		if string(c) != "-" && string(c) != "_" && string(c) != " " {
			return ErrInvalidChar
		}
	}
	return nil
}

// Creates a new player with the given name and adds them to the lobby.
// Returns an error if the lobby is closed, full, finished, or name is invalid.
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

// Marks a player as active and updates lobby activity timestamp.
// Returns an error if the lobby is being closed.
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

// Marks a player as inactive and updates lobby activity timestamp.
func (l *Lobby) Leave(p *Player) {
	l.lock.Lock()
	p.active = false
	l.activity = time.Now().UTC().UnixNano()
	l.lock.Unlock()
}

// Checks if the lobby has been idle longer than the specified interval.
// If idle, closes the lobby and returns true.
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

// Sends a message to a specific player if the lobby is not being closed.
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

// Sends a message to all players in the lobby.
func (l *Lobby) broadcast(msg []byte) {
	l.lock.RLock()
	for _, p := range l.Players {
		l.send(msg, p)
	}
	l.lock.RUnlock()
}

// Processes a move in strict mode where only correct values are accepted.
// Logs legal moves and broadcasts updates to all players.
func (l *Lobby) strictMove(msg *inbound, player *Player) error {
	var conflict string
	var cell [2]int
	update, err := l.Game.Strict(msg.Row, msg.Column, msg.Value)
	illegal := err != nil && err != game.ErrIncorrect

	if update != nil || !illegal {
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
		if illegal {
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

// Processes a move in lax mode where incorrect but non conflicting values are allowed.
// Logs legal moves and sends conflict notifications for rule violations.
func (l *Lobby) laxMove(msg *inbound, player *Player) error {
	update, err := l.Game.Lax(msg.Row, msg.Column, msg.Value)
	illegal := err != nil && (err != game.ErrRowConflict && err != game.ErrColConflict && err != game.ErrBoxConflict)

	if update != nil || !illegal {
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

// Sends the game state (initial and current board) to a player.
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

// Processes a ping message to highlight a cell for all players.
// Validates cell coordinates and broadcasts the cell position.
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

// Processes incoming messages from players and routes them to
// appropriate handlers based on message type.
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
