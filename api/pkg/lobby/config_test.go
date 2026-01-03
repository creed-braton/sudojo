package lobby

import "testing"

func TestValidSize(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  bool
	}{
		{name: "too small size", input: 0, want: false},
		{name: "lower bound size", input: 1, want: true},
		{name: "middle value size", input: 4, want: true},
		{name: "upper bound size", input: 8, want: true},
		{name: "too big size", input: 9, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validSize(tt.input)
			if got != tt.want {
				t.Errorf("expected: '%t', got: '%t'", tt.want, got)
			}
		})
	}
}

func TestConfig(t *testing.T) {
	strict, ping, notes, maxSize := false, true, false, 6
	c, err := NewConfig(strict, ping, notes, maxSize)
	if err != nil {
		t.Fatalf("unexpected error: '%v'", err)
	}
	if c.Strict() != strict {
		t.Errorf("expected strict: '%t', got: '%t'", strict, c.Strict())
	}
	if c.Ping() != ping {
		t.Errorf("expected ping: '%t', got: '%t'", ping, c.Ping())
	}
	if c.Notes() != notes {
		t.Errorf("expected notes: '%t', got: '%t'", notes, c.Notes())
	}
	if c.MaxSize() != maxSize {
		t.Errorf("expected max size: %d, got: %d", maxSize, c.MaxSize())
	}
}
