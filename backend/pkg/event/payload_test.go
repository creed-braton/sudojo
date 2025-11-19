package event

import (
	"sudojo/pkg/player"
	"sudojo/pkg/sudoku"
	"testing"
)

func TestPayload(t *testing.T) {
	p := NewPayload()

	strict := true
	initial, current := sudoku.New(), sudoku.New()
	row, col, val, maxPlayer := 3, 4, 7, 8
	conflict := "duplicate number"

	players := []player.Player{
		player.New("tokenA", "Alice"),
		player.New("tokenB", "Bob"),
	}

	p.SetInitial(initial).
		SetCurrent(current).
		SetRow(row).
		SetColumn(col).
		SetValue(val).
		SetConflict(conflict).
		SetPlayers(players).
		SetMaxPlayer(maxPlayer).
		SetStrict(strict)

	if got := p.Initial(); got != initial {
		t.Errorf("expected initial '%+v', got '%+v'", initial, got)
	}
	if got := p.Current(); got != current {
		t.Errorf("expected current '%+v', got '%+v'", current, got)
	}
	if got := p.Row(); got == nil || *got != row {
		t.Errorf("expected row '%d', got '%v'", row, got)
	}
	if got := p.Column(); got == nil || *got != col {
		t.Errorf("expected column '%d', got '%v'", col, got)
	}
	if got := p.Value(); got == nil || *got != val {
		t.Errorf("expected value '%d', got '%v'", val, got)
	}
	if got := p.Conflict(); got != conflict {
		t.Errorf("expected conflict '%s', got '%s'", conflict, got)
	}
	if got := p.Players(); len(got) != len(players) {
		t.Errorf("expected players length '%d', got '%d'", len(players), len(got))
	}
	if got := p.MaxPlayer(); got != maxPlayer {
		t.Errorf("expected max player '%d', got '%d'", maxPlayer, got)
	}
	if got := p.Strict(); *got != strict {
		t.Errorf("expected strict '%t', got '%t'", strict, *got)
	}
}
