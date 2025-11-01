package lobby

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestValidName(t *testing.T) {
	var tests = []struct {
		name  string
		input string
		want  error
	}{
		{name: "empty name", input: "", want: nil},
		{name: "just too long name", input: "username123456789", want: ErrNameTooLong},
		{name: "almost too long name", input: "username12345678", want: nil},
		{name: "illegal character", input: "user+name", want: ErrInvalidChar},
		{name: "legal special character", input: "user-name", want: nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := validName(test.input)
			if test.want != got {
				t.Errorf("want: %v, got: %v", test.want, got)
			}
		})
	}
}

func TestLobbyPlayer(t *testing.T) {
	t.Run("player retrieval", func(t *testing.T) {
		logger := make(chan *Log, 1024)
		maxPlayer := 16
		l := New(true, maxPlayer, logger)

		players := make([]*Player, 0, maxPlayer)

		for i := range maxPlayer {
			p, err := l.Create(fmt.Sprintf("player%d", i))
			if err != nil {
				t.Fatal(err)
			}
			players = append(players, p)
		}

		for _, p := range players {
			retrieved := l.Player(p.Token)
			if retrieved == nil {
				t.Errorf("want player %s, got nil", p.Name)
				continue
			}
			if retrieved.Token != p.Token {
				t.Errorf("want token %s, got %s", p.Token, retrieved.Token)
			}
			if retrieved.Name != p.Name {
				t.Errorf("want name %s, got %s", p.Name, retrieved.Name)
			}
		}

		if l.Player("invalid-token") != nil {
			t.Error("wanted nil for invalid token, got a player")
		}
	})

	t.Run("unique player tokens", func(t *testing.T) {
		logger := make(chan *Log, 1024)
		maxPlayer := 16
		l := New(true, maxPlayer, logger)
		for range maxPlayer {
			_, err := l.Create("player")
			if err != nil {
				t.Fatal(err)
			}
		}

		tokens := make(map[string]bool)
		for token := range l.Players {
			if tokens[token] {
				t.Errorf("duplicate token found: %s", token)
			}
			tokens[token] = true
		}
	})

	t.Run("maximum player size", func(t *testing.T) {
		logger := make(chan *Log, 1024)
		maxPlayer := 16
		l := New(true, maxPlayer, logger)
		for range maxPlayer {
			_, err := l.Create("in-size")
			if err != nil {
				t.Fatal(err)
			}
		}
		_, err := l.Create("out-of-size")
		if err == nil {
			t.Error("wanted error, got nil")
			return
		}
		if err != ErrLobbyFull {
			t.Errorf("want: %v, got: %v", ErrLobbyFull, err)
		}
	})

	t.Run("lobby closed", func(t *testing.T) {
		logger := make(chan *Log, 1024)
		l := New(true, 16, logger)
		p, err := l.Create("valid")
		l.close()

		_, err = l.Create("invalid")
		if err == nil {
			t.Error("wanted error, got nil")
		} else if err != ErrLobbyClosed {
			t.Errorf("want: %v, got: %v", ErrLobbyClosed, err)
		}

		err = l.Join(p)
		if err == nil {
			t.Error("wanted error, got nil")
		} else if err != ErrLobbyClosed {
			t.Errorf("want: %v, got: %v", ErrLobbyClosed, err)
		}
	})
}

func TestLobbyLifecycle(t *testing.T) {
	t.Run("load lobby state", func(t *testing.T) {
		logger := make(chan *Log, 1024)
		maxPlayer := 16
		l := New(true, maxPlayer, logger)
		for i := range maxPlayer {
			_, err := l.Create(fmt.Sprintf("player%d", i))
			if err != nil {
				t.Fatal(err)
			}
		}
		l.close()
		l.Init(logger)

		msg := "test"
		b, err := (&outbound{
			Error: msg,
		}).marshal()
		if err != nil {
			t.Fatal(err)
		}
		l.broadcast(b)

		for _, p := range l.Players {
			select {
			case b := <-p.Out:
				o := &outbound{}
				if err := json.Unmarshal(b, o); err != nil {
					t.Fatal(err)
				}
				if o.Error != msg {
					t.Errorf("want: %s, got: %s", msg, o.Error)
				}
			case <-time.After(time.Second):
				t.Fatal("timeout waiting for broadcast")
			}
		}
	})
}
