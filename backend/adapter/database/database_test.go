package database

import (
	"sudojo/pkg/lobby"
	"testing"

	"github.com/google/uuid"
)

func setup() *database {
	db, err := New("localhost", "5432", "postgres", "postgres", "password")
	if err != nil {
		panic(err)
	}
	return db
}

func assert(t *testing.T, want, got lobby.Lobby) {
	if want.Id() != got.Id() {
		t.Errorf("expected lobby ID '%s', got '%s'", want.Id(), got.Id())
	}

	if !want.Game().Current().Equal(got.Game().Current()) {
		t.Errorf(
			"expected current board %v, got %v",
			want.Game().Current(),
			got.Game().Current(),
		)
	}
	if !want.Game().Initial().Equal(got.Game().Initial()) {
		t.Errorf(
			"expected initial board %v, got %v",
			want.Game().Initial(),
			got.Game().Initial(),
		)
	}
	if !want.Game().Solution().Equal(got.Game().Solution()) {
		t.Errorf(
			"expected solution %v, got %v",
			want.Game().Solution(),
			got.Game().Solution(),
		)
	}

	if want.Strict() != got.Strict() {
		t.Errorf("expected strict '%t', got '%t'", want.Strict(), got.Strict())
	}
	if want.MaxPlayer() != got.MaxPlayer() {
		t.Errorf("expected max player '%d', got '%d'", want.MaxPlayer(), got.MaxPlayer())
	}

	if (want.Game().Started()) == nil != (got.Game().Started() == nil) {
		t.Errorf(
			"expected started nil timestamp '%t', got '%t'",
			(want.Game().Started()) == nil,
			(got.Game().Started()) == nil,
		)
	}
	if want.Game().Started() != nil {
		if *want.Game().Started() != *got.Game().Started() {
			t.Errorf(
				"expected started timestamp '%d', got '%d'",
				*want.Game().Started(), *got.Game().Started(),
			)
		}
	}

	if (want.Game().Finished()) == nil != (got.Game().Finished() == nil) {
		t.Errorf(
			"expected finished nil timestamp '%t', got '%t'",
			(want.Game().Finished()) == nil,
			(got.Game().Finished()) == nil,
		)
	}
	if want.Game().Finished() != nil {
		if *want.Game().Finished() != *got.Game().Finished() {
			t.Errorf(
				"expected finished timestamp '%d', got '%d'",
				*want.Game().Finished(), *got.Game().Finished(),
			)
		}
	}
}

func TestLobby(t *testing.T) {
	db := setup()
	defer db.Close()

	t.Run("insert and retrieve lobby", func(t *testing.T) {
		want := lobby.Open(8, false)
		if err := db.InsertLobby(want); err != nil {
			t.Fatalf("unexpected error inserting lobby '%v'", err)
		}
		got, err := db.Lobby(want.Id())
		if err != nil {
			t.Fatalf("unexpected error retrieving lobby '%v'", err)
		}
		if got == nil {
			t.Fatal("expected lobby, got nil")
		}

		assert(t, want, got)
	})

	t.Run("non-existing lobby", func(t *testing.T) {
		got, err := db.Lobby(uuid.NewString())
		if err != nil {
			t.Fatalf("unexpected error retrieving lobby '%v'", err)
		}
		if got != nil {
			t.Error("found unexpected lobby")
		}
	})

	t.Run("update lobby", func(t *testing.T) {
		want := lobby.Open(8, false)
		if err := db.InsertLobby(want); err != nil {
			t.Fatalf("unexpected error inserting lobby '%v'", err)
		}
		for i := range len(want.Game().Solution().Int()) {
			for j := range len(want.Game().Solution().Int()[i]) {
				val := want.Game().Solution().Int()[i][j]
				want.Insert(i, j, val)
			}
		}
		if err := db.UpdateLobby(want); err != nil {
			t.Fatalf("unexpected error updating lobby '%v'", err)
		}
		got, err := db.Lobby(want.Id())
		if err != nil {
			t.Fatalf("unexpected error retrieving lobby '%v'", err)
		}
		if got == nil {
			t.Fatal("expected lobby, got nil")
		}

		assert(t, want, got)
	})
}
