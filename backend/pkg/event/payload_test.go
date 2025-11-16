package event

import (
	"sudojo/pkg/player"
	"sudojo/pkg/sudoku"
	"testing"
)

func TestPayload(t *testing.T) {
	p := NewPayload()

	initial := sudoku.New()
	current := sudoku.New()
	row := 3
	column := 4
	value := 7
	conflict := "duplicate number"

	players := []player.Player{
		player.New("tokenA", "Alice"),
		player.New("tokenB", "Bob"),
	}

	p.SetInitial(initial).
		SetCurrent(current).
		SetRow(row).
		SetColumn(column).
		SetValue(value).
		SetConflict(conflict).
		SetPlayers(players)

	if got := p.Initial(); got != initial {
		t.Errorf("expected initial '%+v', got '%+v'", initial, got)
	}
	if got := p.Current(); got != current {
		t.Errorf("expected current '%+v', got '%+v'", current, got)
	}
	if got := p.Row(); got == nil || *got != row {
		t.Errorf("expected row '%d', got '%v'", row, got)
	}
	if got := p.Column(); got == nil || *got != column {
		t.Errorf("expected column '%d', got '%v'", column, got)
	}
	if got := p.Value(); got == nil || *got != value {
		t.Errorf("expected value '%d', got '%v'", value, got)
	}
	if got := p.Conflict(); got != conflict {
		t.Errorf("expected conflict '%s', got '%s'", conflict, got)
	}
	if got := p.Players(); len(got) != len(players) {
		t.Errorf("expected players length '%d', got '%d'", len(players), len(got))
	}
}
