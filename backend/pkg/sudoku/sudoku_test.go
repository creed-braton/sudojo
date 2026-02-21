package sudoku

import "testing"

func TestReadWrite(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input [3]int // [row, col, val]
	}{
		{name: "top-left cell", input: [3]int{0, 0, 1}},
		{name: "center cell", input: [3]int{4, 4, 5}},
		{name: "bottom-right cell", input: [3]int{8, 8, 9}},
		{name: "top-right cell", input: [3]int{0, 8, 7}},
		{name: "bottom-left cell", input: [3]int{8, 0, 3}},
		{name: "random inner cell", input: [3]int{3, 7, 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			row, col, val := tt.input[0], tt.input[1], tt.input[2]

			t.Run("consistency", func(t *testing.T) {
				s := New()
				s.SetCell(row, col, val)
				got := s.Cell(row, col)
				if got != val {
					t.Errorf("consistency violated: Cell(%d,%d) = %d, want %d", row, col, got, val)
				}
			})

			t.Run("isolation", func(t *testing.T) {
				s := New()
				s.SetCell(row, col, val)

				// Ensure all other cells remain 0
				for r := 0; r < 9; r++ {
					for c := 0; c < 9; c++ {
						if r == row && c == col {
							continue
						}
						if v := s.Cell(r, c); v != 0 {
							t.Errorf("isolation violated: Cell(%d,%d) = %d, want 0", r, c, v)
						}
					}
				}
			})

			t.Run("idempotence", func(t *testing.T) {
				s := New()
				s.SetCell(row, col, val)
				before := s.Cell(row, col)
				s.SetCell(row, col, val)
				after := s.Cell(row, col)
				if before != after {
					t.Errorf("idempotence violated at Cell(%d,%d): %d != %d", row, col, before, after)
				}
			})
		})
	}
}

func TestEqual(t *testing.T) {
	t.Parallel()
	var tests = []struct {
		name  string
		input [2]*sudoku
		want  bool
	}{
		{
			name:  "empty sudoku boards",
			input: [2]*sudoku{New(), New()},
			want:  false,
		},
		{
			name: "last cell different",
			input: [2]*sudoku{
				{
					{5, 3, 4, 6, 7, 8, 9, 1, 2},
					{6, 7, 2, 1, 9, 5, 3, 4, 8},
					{1, 9, 8, 3, 4, 2, 5, 6, 7},
					{8, 5, 9, 7, 6, 1, 4, 2, 3},
					{4, 2, 6, 8, 5, 3, 7, 9, 1},
					{7, 1, 3, 9, 2, 4, 8, 5, 6},
					{9, 6, 1, 5, 3, 7, 2, 8, 4},
					{2, 8, 7, 4, 1, 9, 6, 3, 5},
					{3, 4, 5, 2, 8, 6, 1, 7, 9},
				},
				{
					{5, 3, 4, 6, 7, 8, 9, 1, 2},
					{6, 7, 2, 1, 9, 5, 3, 4, 8},
					{1, 9, 8, 3, 4, 2, 5, 6, 7},
					{8, 5, 9, 7, 6, 1, 4, 2, 3},
					{4, 2, 6, 8, 5, 3, 7, 9, 1},
					{7, 1, 3, 9, 2, 4, 8, 5, 6},
					{9, 6, 1, 5, 3, 7, 2, 8, 4},
					{2, 8, 7, 4, 1, 9, 6, 3, 5},
					{3, 4, 5, 2, 8, 6, 1, 7, 8}, // last cell 9 instead of 8
				},
			},
			want: false,
		},
		{
			name: "middle cell different",
			input: [2]*sudoku{
				{
					{0, 0, 0, 0, 0, 0, 0, 1, 0},
					{0, 0, 0, 0, 0, 2, 0, 0, 3},
					{0, 0, 0, 4, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 5, 0, 0},
					{4, 0, 1, 6, 0, 0, 0, 0, 0},
					{0, 0, 7, 1, 0, 0, 0, 0, 0},
					{0, 5, 0, 0, 0, 0, 2, 0, 0},
					{0, 0, 0, 0, 8, 0, 0, 4, 0},
					{0, 3, 0, 9, 1, 0, 0, 0, 0},
				},
				{
					{0, 0, 0, 0, 0, 0, 0, 1, 0},
					{0, 0, 0, 0, 0, 2, 0, 0, 3},
					{0, 0, 0, 4, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 5, 0, 0},
					{4, 0, 1, 6, 1, 0, 0, 0, 0}, // middle cell 1 instead of 0
					{0, 0, 7, 1, 0, 0, 0, 0, 0},
					{0, 5, 0, 0, 0, 0, 2, 0, 0},
					{0, 0, 0, 0, 8, 0, 0, 4, 0},
					{0, 3, 0, 9, 1, 0, 0, 0, 0},
				},
			},
			want: false,
		},
		{
			name: "equal sudoku boards",
			input: [2]*sudoku{
				{
					{5, 3, 4, 6, 7, 8, 9, 1, 2},
					{6, 7, 2, 1, 9, 5, 3, 4, 8},
					{1, 9, 8, 3, 4, 2, 5, 6, 7},
					{8, 5, 9, 7, 6, 1, 4, 2, 3},
					{4, 2, 6, 8, 5, 3, 7, 9, 1},
					{7, 1, 3, 9, 2, 4, 8, 5, 6},
					{9, 6, 1, 5, 3, 7, 2, 8, 4},
					{2, 8, 7, 4, 1, 9, 6, 3, 5},
					{3, 4, 5, 2, 8, 6, 1, 7, 9},
				},
				{
					{5, 3, 4, 6, 7, 8, 9, 1, 2},
					{6, 7, 2, 1, 9, 5, 3, 4, 8},
					{1, 9, 8, 3, 4, 2, 5, 6, 7},
					{8, 5, 9, 7, 6, 1, 4, 2, 3},
					{4, 2, 6, 8, 5, 3, 7, 9, 1},
					{7, 1, 3, 9, 2, 4, 8, 5, 6},
					{9, 6, 1, 5, 3, 7, 2, 8, 4},
					{2, 8, 7, 4, 1, 9, 6, 3, 5},
					{3, 4, 5, 2, 8, 6, 1, 7, 9},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// boards should always be in both directions equal
			got := tt.input[0].Equal(tt.input[1]) &&
				tt.input[1].Equal(tt.input[0])
			if tt.want && !got {
				t.Errorf("want: %t, got: %t", tt.want, got)
			}
		})
	}
}

func TestSerialization(t *testing.T) {
	t.Parallel()
	s := &sudoku{
		{5, 3, 0, 0, 7, 0, 0, 0, 0},
		{6, 0, 0, 1, 9, 5, 0, 0, 0},
		{0, 9, 8, 0, 0, 0, 0, 6, 0},
		{8, 0, 0, 0, 6, 0, 0, 0, 3},
		{4, 0, 0, 8, 0, 3, 0, 0, 1},
		{7, 0, 0, 0, 2, 0, 0, 0, 6},
		{0, 6, 0, 0, 0, 0, 2, 8, 0},
		{0, 0, 0, 4, 1, 9, 0, 0, 5},
		{0, 0, 0, 0, 8, 0, 0, 7, 9},
	}

	t.Run("serialization format", func(t *testing.T) {
		matrix := s.Int()
		if len(matrix) != BoardSize {
			t.Errorf("invalid column length, want: %d, got: %d", BoardSize, len(matrix))
		}
		for i, row := range matrix {
			if len(row) != BoardSize {
				t.Errorf(
					"invalid length for row: %d, want: %d, got: %d",
					i, BoardSize, len(row),
				)
			}
		}
	})

	t.Run("round trip serialization", func(t *testing.T) {
		c := NewFromInts(s.Int())
		if !s.Equal(c) {
			t.Error("serialized and deserialized sudoku not equal to original")
		}
	})
}

func TestValidBounds(t *testing.T) {
	t.Parallel()
	var tests = []struct {
		name  string
		input [2]int
		want  bool
	}{
		{name: "in-bound: top-left corner", input: [2]int{0, 0}, want: true},
		{name: "in-bound: bottom-left corner", input: [2]int{8, 0}, want: true},
		{name: "in-bound: top-right corner", input: [2]int{0, 8}, want: true},
		{name: "in-bound: bottom-right corner", input: [2]int{8, 8}, want: true},
		{name: "out-of-bound: negative row", input: [2]int{-1, 4}, want: false},
		{name: "out-of-bound: negative column", input: [2]int{4, -1}, want: false},
		{name: "out-of-bound: row", input: [2]int{9, 0}, want: false},
		{name: "out-of-bound: column", input: [2]int{0, 9}, want: false},
		{name: "in-bound: center", input: [2]int{4, 4}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ValidBounds(tt.input[0], tt.input[1])
			if tt.want != got {
				t.Errorf("got: %t, want: %t", got, tt.want)
			}
		})
	}
}

func TestValidVal(t *testing.T) {
	t.Parallel()
	var tests = []struct {
		name  string
		input int
		want  bool
	}{
		{name: "largest value", input: MaxValue, want: true},
		{name: "empty value", input: EmptyCell, want: false},
		{name: "value overflow by one", input: MaxValue + 1, want: false},
		{name: "negative value", input: -1, want: false},
		{name: "median value", input: 5, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ValidVal(tt.input)
			if tt.want != got {
				t.Errorf("got: %t, want: %t", got, tt.want)
			}
		})
	}
}

func TestValidRow(t *testing.T) {
	t.Parallel()
	type input struct {
		s   *sudoku
		row int
		val int
	}

	s := &sudoku{
		{1, 2, 3, 0, 0, 0, 0, 0, 0},
		{1, 2, 3, 0, 0, 0, 0, 0, 9},
		{1, 2, 3, 4, 5, 6, 7, 8, 9},
		{1, 2, 3, 4, 0, 6, 7, 8, 9},
		{5, 5, 5, 5, 5, 5, 5, 5, 5},
		{5, 5, 5, 5, 5, 5, 5, 5, 5},
		{5, 5, 5, 5, 5, 5, 5, 5, 5},
		{5, 5, 5, 5, 5, 5, 5, 5, 5},
		{5, 5, 5, 5, 5, 5, 5, 5, 5},
	}

	var tests = []struct {
		name  string
		input input
		want  bool
	}{
		{name: "empty board", input: input{s: New(), row: 8, val: 5}, want: true},
		{name: "conflict on first column", input: input{s: s, row: 0, val: 1}, want: false},
		{name: "conflict on last column", input: input{s: s, row: 1, val: 9}, want: false},
		{name: "full diverse row", input: input{s: s, row: 2, val: 5}, want: false},
		{name: "almost full row", input: input{s: s, row: 3, val: 5}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.input.s.ValidRow(tt.input.row, tt.input.val)
			if tt.want != got {
				t.Errorf("got: %t, want: %t", got, tt.want)
			}
		})
	}
}

func TestValidCol(t *testing.T) {
	t.Parallel()
	type input struct {
		s   *sudoku
		col int
		val int
	}

	s := &sudoku{
		{1, 1, 1, 1, 5, 5, 5, 5, 5},
		{2, 2, 2, 2, 5, 5, 5, 5, 5},
		{3, 3, 3, 3, 5, 5, 5, 5, 5},
		{0, 0, 4, 4, 5, 5, 5, 5, 5},
		{0, 0, 5, 0, 5, 5, 5, 5, 5},
		{0, 0, 6, 6, 5, 5, 5, 5, 5},
		{0, 0, 7, 7, 5, 5, 5, 5, 5},
		{0, 0, 8, 8, 5, 5, 5, 5, 5},
		{0, 9, 9, 9, 5, 5, 5, 5, 5},
	}

	var tests = []struct {
		name  string
		input input
		want  bool
	}{
		{name: "empty board", input: input{s: New(), col: 8, val: 5}, want: true},
		{name: "conflict on first row", input: input{s: s, col: 0, val: 1}, want: false},
		{name: "conflict on last row", input: input{s: s, col: 1, val: 9}, want: false},
		{name: "full diverse column", input: input{s: s, col: 2, val: 5}, want: false},
		{name: "almost full column", input: input{s: s, col: 3, val: 5}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.input.s.ValidCol(tt.input.col, tt.input.val)
			if tt.want != got {
				t.Errorf("got: %t, want: %t", got, tt.want)
			}
		})
	}
}

func TestValidBox(t *testing.T) {
	t.Parallel()
	type input struct {
		s   *sudoku
		row int
		col int
		val int
	}

	s := &sudoku{
		{5, 3, 0, 0, 7, 0, 0, 1, 0},
		{6, 0, 0, 1, 9, 5, 0, 0, 0},
		{0, 9, 8, 0, 0, 0, 0, 6, 0},
		{8, 0, 0, 0, 6, 0, 0, 0, 3},
		{4, 0, 0, 8, 0, 3, 0, 0, 1},
		{7, 0, 0, 0, 2, 0, 0, 0, 6},
		{0, 6, 0, 0, 0, 0, 2, 8, 0},
		{0, 0, 0, 4, 1, 9, 0, 0, 5},
		{0, 1, 0, 0, 8, 0, 0, 7, 9},
	}

	var tests = []struct {
		name  string
		input input
		want  bool
	}{
		{name: "in-bound: top-left box", input: input{s: s, row: 2, col: 2, val: 5}, want: false},
		{name: "out-of-bound: top-left box", input: input{s: s, row: 3, col: 3, val: 5}, want: true},
		{name: "in-bound: middle box", input: input{s: s, row: 5, col: 5, val: 6}, want: false},
		{name: "out-of-bound: middle box", input: input{s: s, row: 6, col: 6, val: 6}, want: true},
		{name: "in-bound: bottom-right box", input: input{s: s, row: 6, col: 6, val: 7}, want: false},
		{name: "out-of-bound: bottom-right box", input: input{s: s, row: 5, col: 5, val: 7}, want: true},
		{name: "in-bound: bottom-left box", input: input{s: s, row: 6, col: 2, val: 1}, want: false},
		{name: "out-of-bound: bottom-left box", input: input{s: s, row: 5, col: 3, val: 1}, want: true},
		{name: "in-bound: top-right box", input: input{s: s, row: 2, col: 6, val: 1}, want: false},
		{name: "out-of-bound: top-right box", input: input{s: s, row: 4, col: 5, val: 1}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.input.s.ValidBox(
				tt.input.row,
				tt.input.col,
				tt.input.val,
			)
			if tt.want != got {
				t.Errorf("got: %t, want: %t", got, tt.want)
			}
		})
	}
}

func TestComplete(t *testing.T) {
	t.Parallel()
	var tests = []struct {
		name  string
		input *sudoku
		want  bool
	}{
		{name: "empty board", input: New(), want: false},
		{
			name: "last cell empty",
			input: &sudoku{
				{1, 2, 3, 4, 5, 6, 7, 8, 9},
				{2, 3, 4, 5, 6, 7, 8, 9, 1},
				{3, 4, 5, 6, 7, 8, 9, 1, 2},
				{4, 5, 6, 7, 8, 9, 1, 2, 3},
				{5, 6, 7, 8, 9, 1, 2, 3, 4},
				{6, 7, 8, 9, 1, 2, 3, 4, 5},
				{7, 8, 9, 1, 2, 3, 4, 5, 6},
				{8, 9, 1, 2, 3, 4, 5, 6, 7},
				{9, 1, 2, 3, 4, 5, 6, 7, 0},
			},
			want: false,
		},
		{
			name: "first cell empty",
			input: &sudoku{
				{0, 2, 3, 4, 5, 6, 7, 8, 9},
				{2, 3, 4, 5, 6, 7, 8, 9, 1},
				{3, 4, 5, 6, 7, 8, 9, 1, 2},
				{4, 5, 6, 7, 8, 9, 1, 2, 3},
				{5, 6, 7, 8, 9, 1, 2, 3, 4},
				{6, 7, 8, 9, 1, 2, 3, 4, 5},
				{7, 8, 9, 1, 2, 3, 4, 5, 6},
				{8, 9, 1, 2, 3, 4, 5, 6, 7},
				{9, 1, 2, 3, 4, 5, 6, 7, 8},
			},
			want: false,
		},
		{
			name: "middle cell empty",
			input: &sudoku{
				{1, 2, 3, 4, 5, 6, 7, 8, 9},
				{2, 3, 4, 5, 6, 7, 8, 9, 1},
				{3, 4, 5, 6, 7, 8, 9, 1, 2},
				{4, 5, 6, 7, 8, 9, 1, 2, 3},
				{5, 6, 7, 8, 0, 1, 2, 3, 4},
				{6, 7, 8, 9, 1, 2, 3, 4, 5},
				{7, 8, 9, 1, 2, 3, 4, 5, 6},
				{8, 9, 1, 2, 3, 4, 5, 6, 7},
				{9, 1, 2, 3, 4, 5, 6, 7, 8},
			},
			want: false,
		},
		{
			name: "multiple empty cells",
			input: &sudoku{
				{0, 2, 3, 4, 5, 6, 7, 8, 9},
				{2, 3, 4, 5, 6, 7, 8, 9, 1},
				{3, 4, 5, 6, 7, 8, 9, 1, 2},
				{4, 5, 6, 7, 8, 9, 1, 2, 3},
				{5, 6, 7, 8, 0, 1, 2, 3, 4},
				{6, 7, 8, 9, 1, 2, 3, 4, 5},
				{7, 8, 9, 1, 2, 3, 4, 5, 6},
				{8, 9, 1, 2, 3, 4, 5, 6, 7},
				{9, 1, 2, 3, 4, 5, 6, 7, 0},
			},
			want: false,
		},
		{
			name: "fully filled board",
			input: &sudoku{
				{5, 3, 4, 6, 7, 8, 9, 1, 2},
				{6, 7, 2, 1, 9, 5, 3, 4, 8},
				{1, 9, 8, 3, 4, 2, 5, 6, 7},
				{8, 5, 9, 7, 6, 1, 4, 2, 3},
				{4, 2, 6, 8, 5, 3, 7, 9, 1},
				{7, 1, 3, 9, 2, 4, 8, 5, 6},
				{9, 6, 1, 5, 3, 7, 2, 8, 4},
				{2, 8, 7, 4, 1, 9, 6, 3, 5},
				{3, 4, 5, 2, 8, 6, 1, 7, 9},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.input.Complete()
			if tt.want != got {
				t.Errorf("got: %t, want: %t", got, tt.want)
			}
		})
	}
}

func TestFill(t *testing.T) {
	t.Parallel()
	for seed := range 1000 {
		s := New()
		s.Fill(int64(seed))

		if !s.Complete() {
			t.Errorf("board not complete for seed %d", seed)
		}

		for row := 0; row < BoardSize; row++ {
			for col := 0; col < BoardSize; col++ {
				val := s.Cell(row, col)
				s.SetCell(row, col, EmptyCell)
				if !s.ValidRow(row, val) || !s.ValidCol(col, val) || !s.ValidBox(row, col, val) {
					t.Errorf("invalid board for seed %d at cell (%d,%d)", seed, row, col)
				}
				s.SetCell(row, col, val)
			}
		}
	}
}

func TestCopy(t *testing.T) {
	t.Parallel()
	t.Run("partially filled board", func(t *testing.T) {
		t.Parallel()
		s := &sudoku{
			{0, 0, 0, 0, 0, 0, 0, 1, 0},
			{0, 0, 0, 0, 0, 2, 0, 0, 3},
			{0, 0, 0, 4, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 5, 0, 0},
			{4, 0, 1, 6, 0, 0, 0, 0, 0},
			{0, 0, 7, 1, 0, 0, 0, 0, 0},
			{0, 5, 0, 0, 0, 0, 2, 0, 0},
			{0, 0, 0, 0, 8, 0, 0, 4, 0},
			{0, 3, 0, 9, 1, 0, 0, 0, 0},
		}

		c := New()
		s.Copy(c)

		if !s.Equal(c) {
			t.Error("copy not equal")
		}
		s.SetCell(4, 4, 5)
		if s.Equal(c) {
			t.Error("copy pointer not independent")
		}
	})

	t.Run("complete boards", func(t *testing.T) {
		t.Parallel()
		for seed := range 1000 {
			s := New()
			s.Fill(int64(seed))

			c := New()
			s.Copy(c)

			if !s.Equal(c) {
				t.Errorf("copy not equal for seed %d", seed)
			}
			s.SetCell(4, 4, 0)
			if s.Equal(c) {
				t.Errorf("copy pointer not independent for seed %d", seed)
			}
		}
	})
}

func TestClues(t *testing.T) {
	t.Parallel()
	var tests = []struct {
		name  string
		input *sudoku
		want  int
	}{
		{name: "empty board", input: New(), want: 0},
		{
			name: "17 clue board",
			input: &sudoku{
				{0, 0, 0, 0, 0, 0, 0, 1, 0},
				{0, 0, 0, 0, 0, 2, 0, 0, 3},
				{0, 0, 0, 4, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 5, 0, 0},
				{4, 0, 1, 6, 0, 0, 0, 0, 0},
				{0, 0, 7, 1, 0, 0, 0, 0, 0},
				{0, 5, 0, 0, 0, 0, 2, 0, 0},
				{0, 0, 0, 0, 8, 0, 0, 4, 0},
				{0, 3, 0, 9, 1, 0, 0, 0, 0},
			},
			want: 17,
		},
		{
			name: "16 clue board",
			input: &sudoku{
				{5, 0, 0, 0, 7, 0, 0, 0, 0},
				{6, 0, 0, 1, 0, 0, 0, 0, 0},
				{0, 9, 0, 0, 0, 0, 0, 6, 0},
				{8, 0, 0, 0, 6, 0, 0, 0, 3},
				{0, 0, 0, 8, 0, 0, 0, 0, 0},
				{7, 0, 0, 0, 0, 0, 0, 0, 0},
				{0, 6, 0, 0, 0, 0, 0, 8, 0},
				{0, 0, 0, 4, 0, 0, 0, 0, 5},
				{0, 0, 0, 0, 8, 0, 0, 0, 0},
			},
			want: 16,
		},
		{
			name: "complete board",
			input: &sudoku{
				{5, 3, 4, 6, 7, 8, 9, 1, 2},
				{6, 7, 2, 1, 9, 5, 3, 4, 8},
				{1, 9, 8, 3, 4, 2, 5, 6, 7},
				{8, 5, 9, 7, 6, 1, 4, 2, 3},
				{4, 2, 6, 8, 5, 3, 7, 9, 1},
				{7, 1, 3, 9, 2, 4, 8, 5, 6},
				{9, 6, 1, 5, 3, 7, 2, 8, 4},
				{2, 8, 7, 4, 1, 9, 6, 3, 5},
				{3, 4, 5, 2, 8, 6, 1, 7, 9},
			},
			want: BoardSize * BoardSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.input.clues()
			if tt.want != got {
				t.Errorf("got: %d, want: %d", got, tt.want)
			}
		})
	}
}

func TestUniqueSolution(t *testing.T) {
	t.Parallel()
	var tests = []struct {
		name  string
		input *sudoku
		want  bool
	}{
		{name: "empty board", input: New(), want: false},
		{
			name: "17 clue board",
			input: &sudoku{
				{0, 0, 0, 0, 0, 0, 0, 1, 0},
				{0, 0, 0, 0, 0, 2, 0, 0, 3},
				{0, 0, 0, 4, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 5, 0, 0},
				{4, 0, 1, 6, 0, 0, 0, 0, 0},
				{0, 0, 7, 1, 0, 0, 0, 0, 0},
				{0, 5, 0, 0, 0, 0, 2, 0, 0},
				{0, 0, 0, 0, 8, 0, 0, 4, 0},
				{0, 3, 0, 9, 1, 0, 0, 0, 0},
			},
			want: true,
		},
		{
			name: "16 clue board",
			input: &sudoku{
				{5, 0, 0, 0, 7, 0, 0, 0, 0},
				{6, 0, 0, 1, 0, 0, 0, 0, 0},
				{0, 9, 0, 0, 0, 0, 0, 6, 0},
				{8, 0, 0, 0, 6, 0, 0, 0, 3},
				{0, 0, 0, 8, 0, 0, 0, 0, 0},
				{7, 0, 0, 0, 0, 0, 0, 0, 0},
				{0, 6, 0, 0, 0, 0, 0, 8, 0},
				{0, 0, 0, 4, 0, 0, 0, 0, 5},
				{0, 0, 0, 0, 8, 0, 0, 0, 0},
			},
			want: false,
		},
		{
			name: "complete board",
			input: &sudoku{
				{5, 3, 4, 6, 7, 8, 9, 1, 2},
				{6, 7, 2, 1, 9, 5, 3, 4, 8},
				{1, 9, 8, 3, 4, 2, 5, 6, 7},
				{8, 5, 9, 7, 6, 1, 4, 2, 3},
				{4, 2, 6, 8, 5, 3, 7, 9, 1},
				{7, 1, 3, 9, 2, 4, 8, 5, 6},
				{9, 6, 1, 5, 3, 7, 2, 8, 4},
				{2, 8, 7, 4, 1, 9, 6, 3, 5},
				{3, 4, 5, 2, 8, 6, 1, 7, 9},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := New()
			tt.input.Copy(c)
			got := c.UniqueSolution()

			if tt.want != got {
				t.Errorf("got: %t, want: %t", got, tt.want)
			}

			if tt.want {
				if !c.Complete() {
					t.Error("solution not complete")
					return
				}

				for row := 0; row < BoardSize; row++ {
					for col := 0; col < BoardSize; col++ {
						val := c.Cell(row, col)
						c.SetCell(row, col, EmptyCell)
						if !c.ValidRow(row, val) || !c.ValidCol(col, val) || !c.ValidBox(row, col, val) {
							t.Errorf("invalid solution at cell (%d,%d)", row, col)
						}
						c.SetCell(row, col, val)
					}
				}
			} else {
				if !tt.input.Equal(c) {
					t.Error("non unique solution board overwritten")
				}
			}
		})
	}
}

func TestHash(t *testing.T) {
	t.Parallel()
	var tests = []struct {
		name  string
		input [2]*sudoku
		want  bool
	}{
		{
			name:  "empty boards",
			input: [2]*sudoku{New(), New()},
			want:  true,
		},
		{
			name: "identical complete boards",
			input: [2]*sudoku{
				{
					{5, 3, 4, 6, 7, 8, 9, 1, 2},
					{6, 7, 2, 1, 9, 5, 3, 4, 8},
					{1, 9, 8, 3, 4, 2, 5, 6, 7},
					{8, 5, 9, 7, 6, 1, 4, 2, 3},
					{4, 2, 6, 8, 5, 3, 7, 9, 1},
					{7, 1, 3, 9, 2, 4, 8, 5, 6},
					{9, 6, 1, 5, 3, 7, 2, 8, 4},
					{2, 8, 7, 4, 1, 9, 6, 3, 5},
					{3, 4, 5, 2, 8, 6, 1, 7, 9},
				},
				{
					{5, 3, 4, 6, 7, 8, 9, 1, 2},
					{6, 7, 2, 1, 9, 5, 3, 4, 8},
					{1, 9, 8, 3, 4, 2, 5, 6, 7},
					{8, 5, 9, 7, 6, 1, 4, 2, 3},
					{4, 2, 6, 8, 5, 3, 7, 9, 1},
					{7, 1, 3, 9, 2, 4, 8, 5, 6},
					{9, 6, 1, 5, 3, 7, 2, 8, 4},
					{2, 8, 7, 4, 1, 9, 6, 3, 5},
					{3, 4, 5, 2, 8, 6, 1, 7, 9},
				},
			},
			want: true,
		},
		{
			name: "single cell different",
			input: [2]*sudoku{
				{
					{1, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
				{
					{2, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
			},
			want: false,
		},
		{
			name: "same value different position",
			input: [2]*sudoku{
				{
					{5, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
				{
					{0, 5, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.input[0].Hash() == tt.input[1].Hash()
			if tt.want != got {
				t.Errorf("got: %t, want: %t", got, tt.want)
			}
		})
	}
}

func TestGeneratePuzzle(t *testing.T) {
	t.Parallel()
	for seed := range 100 {
		s := New()
		s.Fill(int64(seed))
		s.GeneratePuzzle(int64(seed))

		if s.Complete() {
			t.Errorf("complete puzzle generated for seed %d", seed)
		}
		if !s.UniqueSolution() {
			t.Errorf("non unique solution puzzle generated for seed %d", seed)
		}
	}
}
