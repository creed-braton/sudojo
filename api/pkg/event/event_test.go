package event

import (
	"sudojo/pkg/game"
	"sudojo/pkg/lobby"
	"sudojo/pkg/player"
	"testing"

	"github.com/google/uuid"
)

func TestEvent(t *testing.T) {
	trace := uuid.NewString()

	t.Run("leave event", func(t *testing.T) {
		p := player.New("token", "john")
		event := New(LeaveEvent, int64(42), trace).
			SetPlayers([]player.Player{p})
	})

	t.Run("join event", func(t *testing.T) {
		p := player.New("token", "john")
		p.SetActive(true)
		event := New(JoinEvent, int64(42), trace).
			SetPlayers([]player.Player{p})
	})

	t.Run("state event", func(t *testing.T) {
		players := []player.Player{
			player.New("a", "john"),
			player.New("b", "emily"),
		}
		players[0].SetActive(true)
		config, err := lobby.NewConfig(true, false, true, 4)
		if err != nil {
			t.Fatalf("unexpected config error: '%v'", err)
		}
		game := game.NewMock(true)

		event := New(StateEvent, int64(42), trace).
			SetPlayers(players).SetConfig(config).
			SetCurrent(game.Current()).SetInitial(game.Initial())
	})

	t.Run("insert event", func(t *testing.T) {
		game := game.NewMock(true)
		row, col := 0, 0
		val := game.Solution().Cell(row, col)

		current, err := game.Strict(row, col, val, int64(42))
		if err != nil {
			t.Fatalf("unexpected insert error: '%v'", err)
		}

		event := New(StateEvent, int64(42), trace).
			SetRow(row).SetColumn(col).SetValue(val).
			SetCurrent(current)
	})

	t.Run("ping event", func(t *testing.T) {
		row, col := 0, 0
		event := New(PingEvent, int64(42), trace).
			SetRow(row).SetColumn(col)
	})
}
