package history

import (
	"sync"
	"testing"
)

func TestArtifact(t *testing.T) {
	player, ts, row, col, val := "test-player", int64(1234567890), 5, 3, 8
	a := NewArtifact(player, ts, row, col, val)

	if a == nil {
		t.Fatal("unexpected nil artifact")
	}

	if a.Player() != player {
		t.Errorf("expected player '%s', got '%s'", player, a.Player())
	}
	if a.Timestamp() != ts {
		t.Errorf("expected timestamp '%d', got '%d'", ts, a.Timestamp())
	}
	if a.Row() != row {
		t.Errorf("expected row '%d', got '%d'", row, a.Row())
	}
	if a.Column() != col {
		t.Errorf("expected column '%d', got '%d'", col, a.Column())
	}
	if a.Value() != val {
		t.Errorf("expected value '%d', got '%d'", val, a.Value())
	}
}

func TestAppend(t *testing.T) {
	t.Run("drops duplicates", func(t *testing.T) {
		initial := []Artifact{
			NewArtifact("player", 100, 1, 2, 3),
		}
		h := New(initial)
		h.Append(NewArtifact("player", 200, 1, 2, 3))
		got := h.Artifacts()
		if len(got) != 1 {
			t.Errorf("expected 1 artifact, got %d", len(got))
		}
	})

	t.Run("concurrent appends", func(t *testing.T) {
		h := New(nil)
		worker := 8
		num := 100

		var wg sync.WaitGroup
		wg.Add(worker)

		for i := 0; i < worker; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < num; j++ {
					a := NewArtifact("player", int64(id*1000+j), id*num+j, 0, 1)
					h.Append(a)
				}
			}(i)
		}

		wg.Wait()

		got := h.Artifacts()
		want := worker * num
		if len(got) != want {
			t.Errorf("expected %d artifacts, got %d", want, len(got))
		}
	})
}

func TestFlush(t *testing.T) {
	t.Run("returns appended artifacts", func(t *testing.T) {
		h := New(nil)
		num := 100

		for i := 0; i < num; i++ {
			a := NewArtifact("player", int64(i), i, i, i)
			h.Append(a)
		}

		f := h.Flush()
		if len(f) != num {
			t.Errorf("expected %d artifacts in flush, got %d", num, len(f))
		}
	})

	t.Run("flush initialized history", func(t *testing.T) {
		initial := []Artifact{
			NewArtifact("", 0, 1, 0, 0),
			NewArtifact("", 0, 0, 1, 0),
			NewArtifact("", 0, 0, 0, 1),
			NewArtifact("", 0, 1, 1, 0),
			NewArtifact("", 0, 1, 0, 1),
		}
		h := New(initial)

		f := h.Flush()
		if len(f) != 0 {
			t.Errorf("expected 0 artifacts in flush, got %d", len(f))
		}
	})

	t.Run("second flush empty", func(t *testing.T) {
		h := New(nil)

		for i := 0; i < 100; i++ {
			a := NewArtifact("player", int64(i), i, i, i)
			h.Append(a)
		}
		h.Flush()

		f := h.Flush()
		if len(f) != 0 {
			t.Errorf("expected 0 artifacts in second flush, got %d", len(f))
		}
	})
}

func TestArtifacts(t *testing.T) {
	t.Run("only initial artifacts", func(t *testing.T) {
		initial := []Artifact{
			NewArtifact("", 0, 1, 0, 0),
			NewArtifact("", 0, 0, 1, 0),
			NewArtifact("", 0, 0, 0, 1),
			NewArtifact("", 0, 1, 1, 0),
			NewArtifact("", 0, 1, 0, 1),
		}
		h := New(initial)
		got := h.Artifacts()
		if len(got) != 5 {
			t.Errorf("expected 5 artifacts, got %d", len(got))
		}
	})

	t.Run("empty artifacts", func(t *testing.T) {
		h := New(nil)
		got := h.Artifacts()
		if len(got) != 0 {
			t.Errorf("expected 0 artifacts, got %d", len(got))
		}
	})

	t.Run("initial and current artifacts", func(t *testing.T) {
		initial := []Artifact{
			NewArtifact("", 0, 1, 0, 0),
			NewArtifact("", 0, 0, 1, 0),
			NewArtifact("", 0, 0, 0, 1),
			NewArtifact("", 0, 1, 1, 0),
			NewArtifact("", 0, 1, 0, 1),
		}
		h := New(initial)
		for i := 0; i < 5; i++ {
			h.Append(NewArtifact("", 0, i, i, i))
		}
		got := h.Artifacts()
		if len(got) != 10 {
			t.Errorf("expected 10 artifacts, got %d", len(got))
		}
	})

	t.Run("current artifacts after flush", func(t *testing.T) {
		h := New(nil)
		for i := 0; i < 5; i++ {
			h.Append(NewArtifact("", 0, i, i, i))
		}
		h.Flush()
		got := h.Artifacts()
		if len(got) != 5 {
			t.Errorf("expected 5 artifacts after flush, got %d", len(got))
		}
	})

	t.Run("concurrent reads", func(t *testing.T) {
		initial := []Artifact{
			NewArtifact("", 0, 1, 0, 0),
			NewArtifact("", 0, 0, 1, 0),
		}
		h := New(initial)
		for i := 0; i < 3; i++ {
			h.Append(NewArtifact("", 0, i, i, i))
		}

		reader := 8
		num := 100
		var wg sync.WaitGroup
		wg.Add(reader)

		for i := 0; i < reader; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < num; j++ {
					_ = h.Artifacts()
				}
			}()
		}
		wg.Wait()

		got := h.Artifacts()
		if len(got) != 5 {
			t.Errorf("expected 5 artifacts after concurrent reads, got %d", len(got))
		}
	})

	t.Run("returns true copy", func(t *testing.T) {
		initial := []Artifact{
			NewArtifact("", 0, 1, 0, 0),
			NewArtifact("", 0, 0, 1, 0),
		}
		h := New(initial)
		h.Append(NewArtifact("", 0, 0, 0, 1))

		got := h.Artifacts()
		got = append(got, NewArtifact("", 0, 1, 1, 0))

		check := h.Artifacts()
		if len(check) != 3 {
			t.Errorf("expected 3 artifacts after modification, got %d", len(check))
		}
	})
}
