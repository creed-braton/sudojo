package event

import (
	"sudojo/pkg/game"
	"sudojo/pkg/lobby"
	"sudojo/pkg/player"
	"testing"

	"github.com/google/uuid"
)

func TestEvent(t *testing.T) {
	trace, now := uuid.NewString(), int64(42)

	t.Run("leave event", func(t *testing.T) {
		p := player.New("token", "john")
		event := New(LeaveEvent, now, trace).
			SetPlayers([]player.Player{p})

		if event.Type() != LeaveEvent {
			t.Errorf("expected type '%s', got '%s'", LeaveEvent, event.Type())
		}
		if event.Timestamp() != now {
			t.Errorf("expected timestamp %d, got %d", now, event.Timestamp())
		}
		if event.Trace() != trace {
			t.Errorf("expected trace '%s', got '%s'", trace, event.Trace())
		}
		if len(event.Players()) != 1 {
			t.Fatalf("expected 1 player, got %d", len(event.Players()))
		}
		if event.Players()[0].Token() != "token" {
			t.Errorf("expected player token 'token', got '%s'", event.Players()[0].Token())
		}
		if event.Players()[0].Name() != "john" {
			t.Errorf("expected player name 'john', got '%s'", event.Players()[0].Name())
		}
		if event.Error() != "" {
			t.Errorf("expected empty error, got '%s'", event.Error())
		}
		if event.Conflict() != "" {
			t.Errorf("expected empty conflict, got '%s'", event.Conflict())
		}
		if event.Config() != nil {
			t.Error("expected nil config, got config")
		}
		if event.Current() != nil {
			t.Error("expected nil current, got current")
		}
		if event.Initial() != nil {
			t.Error("expected nil initial, got initial")
		}
		if event.Row() != nil {
			t.Error("expected nil row, got row")
		}
		if event.Column() != nil {
			t.Error("expected nil column, got column")
		}
		if event.Value() != nil {
			t.Error("expected nil value, got value")
		}
	})

	t.Run("join event", func(t *testing.T) {
		p := player.New("token", "john")
		p.SetActive(true)
		event := New(JoinEvent, now, trace).
			SetPlayers([]player.Player{p})

		if event.Type() != JoinEvent {
			t.Errorf("expected type '%s', got '%s'", JoinEvent, event.Type())
		}
		if event.Timestamp() != now {
			t.Errorf("expected timestamp %d, got %d", now, event.Timestamp())
		}
		if event.Trace() != trace {
			t.Errorf("expected trace '%s', got '%s'", trace, event.Trace())
		}
		if len(event.Players()) != 1 {
			t.Fatalf("expected 1 player, got %d", len(event.Players()))
		}
		if event.Players()[0].Token() != "token" {
			t.Errorf("expected player token 'token', got '%s'", event.Players()[0].Token())
		}
		if !event.Players()[0].Active() {
			t.Error("expected player to be active")
		}
		if event.Error() != "" {
			t.Errorf("expected empty error, got '%s'", event.Error())
		}
		if event.Conflict() != "" {
			t.Errorf("expected empty conflict, got '%s'", event.Conflict())
		}
		if event.Config() != nil {
			t.Error("expected nil config, got config")
		}
		if event.Current() != nil {
			t.Error("expected nil current, got current")
		}
		if event.Initial() != nil {
			t.Error("expected nil initial, got initial")
		}
		if event.Row() != nil {
			t.Error("expected nil row, got row")
		}
		if event.Column() != nil {
			t.Error("expected nil column, got column")
		}
		if event.Value() != nil {
			t.Error("expected nil value, got value")
		}
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
		g := game.NewMock(true)

		event := New(StateEvent, now, trace).
			SetPlayers(players).SetConfig(config).
			SetCurrent(g.Current()).SetInitial(g.Initial())

		if event.Type() != StateEvent {
			t.Errorf("expected type '%s', got '%s'", StateEvent, event.Type())
		}
		if event.Timestamp() != now {
			t.Errorf("expected timestamp %d, got %d", now, event.Timestamp())
		}
		if event.Trace() != trace {
			t.Errorf("expected trace '%s', got '%s'", trace, event.Trace())
		}
		if len(event.Players()) != 2 {
			t.Fatalf("expected 2 players, got %d", len(event.Players()))
		}
		if event.Players()[0].Token() != "a" {
			t.Errorf("expected first player token 'a', got '%s'", event.Players()[0].Token())
		}
		if event.Players()[1].Token() != "b" {
			t.Errorf("expected second player token 'b', got '%s'", event.Players()[1].Token())
		}
		if !event.Players()[0].Active() {
			t.Error("expected first player to be active")
		}
		if event.Players()[1].Active() {
			t.Error("expected second player to be inactive")
		}
		if event.Config() == nil {
			t.Fatal("expected config, got nil")
		}
		if !event.Config().Strict() {
			t.Error("expected config strict to be true")
		}
		if event.Config().Ping() {
			t.Error("expected config ping to be false")
		}
		if !event.Config().Notes() {
			t.Error("expected config notes to be true")
		}
		if event.Config().MaxSize() != 4 {
			t.Errorf("expected config max size 4, got %d", event.Config().MaxSize())
		}
		if event.Current() == nil {
			t.Fatal("expected current, got nil")
		}
		if !event.Current().Equal(g.Current()) {
			t.Error("expected current to equal game current")
		}
		if event.Initial() == nil {
			t.Fatal("expected initial, got nil")
		}
		if !event.Initial().Equal(g.Initial()) {
			t.Error("expected initial to equal game initial")
		}
		if event.Error() != "" {
			t.Errorf("expected empty error, got '%s'", event.Error())
		}
		if event.Conflict() != "" {
			t.Errorf("expected empty conflict, got '%s'", event.Conflict())
		}
		if event.Row() != nil {
			t.Error("expected nil row, got row")
		}
		if event.Column() != nil {
			t.Error("expected nil column, got column")
		}
		if event.Value() != nil {
			t.Error("expected nil value, got value")
		}
	})

	t.Run("insert event", func(t *testing.T) {
		g := game.NewMock(true)
		row, col := 0, 0
		val := g.Solution().Cell(row, col)

		current, err := g.Strict(row, col, val, now)
		if err != nil {
			t.Fatalf("unexpected insert error: '%v'", err)
		}

		event := New(InsertEvent, now, trace).
			SetRow(row).SetColumn(col).SetValue(val).
			SetCurrent(current)

		if event.Type() != InsertEvent {
			t.Errorf("expected type '%s', got '%s'", InsertEvent, event.Type())
		}
		if event.Timestamp() != now {
			t.Errorf("expected timestamp %d, got %d", now, event.Timestamp())
		}
		if event.Trace() != trace {
			t.Errorf("expected trace '%s', got '%s'", trace, event.Trace())
		}
		if event.Row() == nil {
			t.Fatal("expected row, got nil")
		}
		if *event.Row() != row {
			t.Errorf("expected row %d, got %d", row, *event.Row())
		}
		if event.Column() == nil {
			t.Fatal("expected column, got nil")
		}
		if *event.Column() != col {
			t.Errorf("expected column %d, got %d", col, *event.Column())
		}
		if event.Value() == nil {
			t.Fatal("expected value, got nil")
		}
		if *event.Value() != val {
			t.Errorf("expected value %d, got %d", val, *event.Value())
		}
		if event.Current() == nil {
			t.Fatal("expected current, got nil")
		}
		if !event.Current().Equal(current) {
			t.Error("expected current to equal inserted current")
		}
		if event.Error() != "" {
			t.Errorf("expected empty error, got '%s'", event.Error())
		}
		if event.Conflict() != "" {
			t.Errorf("expected empty conflict, got '%s'", event.Conflict())
		}
		if event.Config() != nil {
			t.Error("expected nil config, got config")
		}
		if event.Initial() != nil {
			t.Error("expected nil initial, got initial")
		}
		if event.Players() != nil {
			t.Error("expected nil players, got players")
		}
	})

	t.Run("ping event", func(t *testing.T) {
		row, col := 0, 0
		event := New(PingEvent, now, trace).
			SetRow(row).SetColumn(col)

		if event.Type() != PingEvent {
			t.Errorf("expected type '%s', got '%s'", PingEvent, event.Type())
		}
		if event.Timestamp() != now {
			t.Errorf("expected timestamp %d, got %d", now, event.Timestamp())
		}
		if event.Trace() != trace {
			t.Errorf("expected trace '%s', got '%s'", trace, event.Trace())
		}
		if event.Row() == nil {
			t.Fatal("expected row, got nil")
		}
		if *event.Row() != row {
			t.Errorf("expected row %d, got %d", row, *event.Row())
		}
		if event.Column() == nil {
			t.Fatal("expected column, got nil")
		}
		if *event.Column() != col {
			t.Errorf("expected column %d, got %d", col, *event.Column())
		}
		if event.Error() != "" {
			t.Errorf("expected empty error, got '%s'", event.Error())
		}
		if event.Conflict() != "" {
			t.Errorf("expected empty conflict, got '%s'", event.Conflict())
		}
		if event.Config() != nil {
			t.Error("expected nil config, got config")
		}
		if event.Current() != nil {
			t.Error("expected nil current, got current")
		}
		if event.Initial() != nil {
			t.Error("expected nil initial, got initial")
		}
		if event.Value() != nil {
			t.Error("expected nil value, got value")
		}
		if event.Players() != nil {
			t.Error("expected nil players, got players")
		}
	})
}
