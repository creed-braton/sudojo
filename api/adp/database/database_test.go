package database

import (
	"sudojo/pkg/lobby"
	"testing"

	"github.com/google/uuid"
)

func setup(t *testing.T) *database {
	t.Helper()

	db, err := New("localhost", "5432", "postgres", "postgres", "password")
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func assert(t *testing.T, want, got lobby.Lobby) {
	t.Helper()

	if want.Id() != got.Id() {
		t.Errorf("expected lobby ID '%s', got '%s'", want.Id(), got.Id())
	}

	if want.Game().Hash() != got.Game().Hash() {
		t.Errorf("expected game hash '%s', got '%s'", want.Game().Hash(), got.Game().Hash())
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

	if want.Config().Strict() != got.Config().Strict() {
		t.Errorf("expected strict '%t', got '%t'", want.Config().Strict(), got.Config().Strict())
	}
	if want.Config().Ping() != got.Config().Ping() {
		t.Errorf("expected ping '%t', got '%t'", want.Config().Ping(), got.Config().Ping())
	}
	if want.Config().Notes() != got.Config().Notes() {
		t.Errorf("expected strict '%t', got '%t'", want.Config().Notes(), got.Config().Notes())
	}
	if want.Config().MaxSize() != got.Config().MaxSize() {
		t.Errorf("expected lobby size '%d', got '%d'", want.Config().MaxSize(), got.Config().MaxSize())
	}

	if (want.Game().StartedAt()) == nil != (got.Game().StartedAt() == nil) {
		t.Errorf(
			"expected started nil timestamp '%t', got '%t'",
			(want.Game().StartedAt()) == nil,
			(got.Game().StartedAt()) == nil,
		)
	}
	if want.Game().StartedAt() != nil {
		if *want.Game().StartedAt() != *got.Game().StartedAt() {
			t.Errorf(
				"expected started timestamp '%d', got '%d'",
				*want.Game().StartedAt(), *got.Game().StartedAt(),
			)
		}
	}

	if (want.Game().FinishedAt()) == nil != (got.Game().FinishedAt() == nil) {
		t.Errorf(
			"expected finished nil timestamp '%t', got '%t'",
			(want.Game().FinishedAt()) == nil,
			(got.Game().FinishedAt()) == nil,
		)
	}
	if want.Game().FinishedAt() != nil {
		if *want.Game().FinishedAt() != *got.Game().FinishedAt() {
			t.Errorf(
				"expected finished timestamp '%d', got '%d'",
				*want.Game().FinishedAt(), *got.Game().FinishedAt(),
			)
		}
	}
}

func TestLobby(t *testing.T) {
	db := setup(t)
	defer db.Close()

	t.Run("insert and retrieve lobby", func(t *testing.T) {
		want := lobby.NewMock(true, true, true, 8)
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
			t.Error("unexpected lobby found")
		}
	})

	t.Run("update lobby", func(t *testing.T) {
		want := lobby.NewMock(true, true, true, 8)
		if err := db.InsertLobby(want); err != nil {
			t.Fatalf("unexpected error inserting lobby '%v'", err)
		}
		for i := range len(want.Game().Solution().Int()) {
			for j := range len(want.Game().Solution().Int()[i]) {
				val := want.Game().Solution().Int()[i][j]
				want.Game().Lax(i, j, val, int64(42))
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
