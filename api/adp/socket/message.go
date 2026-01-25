package socket

type Player struct {
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

type Config struct {
	Strict  bool `json:"strict_mode"`
	Pings   bool `json:"pings_allowed"`
	Notes   bool `json:"notes_allowed"`
	MaxSize int  `json:"max_player"`
}

type Message struct {
	Type       string    `json:"type"`
	Trace      string    `json:"trace,omitempty"`
	Error      string    `json:"error,omitempty"`
	Current    [][]int   `json:"current_board,omitempty"`
	Initial    [][]int   `json:"initial_board,omitempty"`
	Conflict   string    `json:"conflict,omitempty"`
	Row        *int      `json:"row,omitempty"`
	Column     *int      `json:"column,omitempty"`
	Value      *int      `json:"value,omitempty"`
	Players    []*Player `json:"players,omitempty"`
	Config     *Config   `json:"config,omitempty"`
	Difficulty string    `json:"difficulty,omitempty"`
}
