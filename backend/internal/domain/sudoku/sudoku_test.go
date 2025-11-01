package sudoku

import "testing"

func TestEqual(t *testing.T) {
	var tests = []struct {
		name  string
		input [2]*Sudoku
		want  bool
	}{
		{
			name:  "empty sudoku boards",
			input: [2]*Sudoku{New(), New()},
			want:  false,
		},
		{
			name: "last cell different",
			input: [2]*Sudoku{
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
			input: [2]*Sudoku{
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
			input: [2]*Sudoku{
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// boards should always be in both directions equal
			got := test.input[0].Equal(test.input[1]) &&
				test.input[1].Equal(test.input[0])
			if test.want && !got {
				t.Errorf("want: %t, got: %t", test.want, got)
			}
		})
	}
}

func TestSerialization(t *testing.T) {
	s := &Sudoku{
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
		c := Load(s.Int())
		if !s.Equal(c) {
			t.Error("serialized and deserialized sudoku not equal to original")
		}
	})
}

func TestValidBounds(t *testing.T) {
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ValidBounds(test.input[0], test.input[1])
			if test.want != got {
				t.Errorf("got: %t, want: %t", got, test.want)
			}
		})
	}
}

func TestValidVal(t *testing.T) {
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ValidVal(test.input)
			if test.want != got {
				t.Errorf("got: %t, want: %t", got, test.want)
			}
		})
	}
}

func TestValidRow(t *testing.T) {
	type input struct {
		s   *Sudoku
		row int
		val int
	}

	s := &Sudoku{
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.input.s.ValidRow(test.input.row, test.input.val)
			if test.want != got {
				t.Errorf("got: %t, want: %t", got, test.want)
			}
		})
	}
}

func TestValidCol(t *testing.T) {
	type input struct {
		s   *Sudoku
		col int
		val int
	}

	s := &Sudoku{
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.input.s.ValidCol(test.input.col, test.input.val)
			if test.want != got {
				t.Errorf("got: %t, want: %t", got, test.want)
			}
		})
	}
}

func TestValidBox(t *testing.T) {
	type input struct {
		s   *Sudoku
		row int
		col int
		val int
	}

	s := &Sudoku{
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.input.s.ValidBox(
				test.input.row,
				test.input.col,
				test.input.val,
			)
			if test.want != got {
				t.Errorf("got: %t, want: %t", got, test.want)
			}
		})
	}
}

func TestComplete(t *testing.T) {
	var tests = []struct {
		name  string
		input *Sudoku
		want  bool
	}{
		{name: "empty board", input: New(), want: false},
		{
			name: "last cell empty",
			input: &Sudoku{
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
			input: &Sudoku{
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
			input: &Sudoku{
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
			input: &Sudoku{
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
			input: &Sudoku{
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.input.Complete()
			if test.want != got {
				t.Errorf("got: %t, want: %t", got, test.want)
			}
		})
	}
}

func TestFill(t *testing.T) {
	for seed := range 1000 {
		s := New()
		s.Fill(int64(seed))

		if !s.Complete() {
			t.Errorf("board not complete for seed %d", seed)
		}

		for row := 0; row < BoardSize; row++ {
			for col := 0; col < BoardSize; col++ {
				val := s[row][col]
				s[row][col] = EmptyCell
				if !s.ValidRow(row, val) || !s.ValidCol(col, val) || !s.ValidBox(row, col, val) {
					t.Errorf("invalid board for seed %d at cell (%d,%d)", seed, row, col)
				}
				s[row][col] = val
			}
		}
	}
}

func TestCopy(t *testing.T) {
	t.Run("partially filled board", func(t *testing.T) {
		s := &Sudoku{
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
		s[4][4] = 5
		if s.Equal(c) {
			t.Error("copy pointer not independent")
		}
	})

	t.Run("complete boards", func(t *testing.T) {
		for seed := range 1000 {
			s := New()
			s.Fill(int64(seed))

			c := New()
			s.Copy(c)

			if !s.Equal(c) {
				t.Errorf("copy not equal for seed %d", seed)
			}
			s[4][4] = 0
			if s.Equal(c) {
				t.Errorf("copy pointer not independent for seed %d", seed)
			}
		}
	})
}

func TestClues(t *testing.T) {
	var tests = []struct {
		name  string
		input *Sudoku
		want  int
	}{
		{name: "empty board", input: New(), want: 0},
		{
			name: "17 clue board",
			input: &Sudoku{
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
			input: &Sudoku{
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
			input: &Sudoku{
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.input.clues()
			if test.want != got {
				t.Errorf("got: %d, want: %d", got, test.want)
			}
		})
	}
}

func TestUniqueSolution(t *testing.T) {
	var tests = []struct {
		name  string
		input *Sudoku
		want  bool
	}{
		{name: "empty board", input: New(), want: false},
		{
			name: "17 clue board",
			input: &Sudoku{
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
			input: &Sudoku{
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
			input: &Sudoku{
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := New()
			test.input.Copy(c)
			got := c.UniqueSolution()

			if test.want != got {
				t.Errorf("got: %t, want: %t", got, test.want)
			}

			if test.want {
				if !c.Complete() {
					t.Error("solution not complete")
					return
				}

				for row := 0; row < BoardSize; row++ {
					for col := 0; col < BoardSize; col++ {
						val := c[row][col]
						c[row][col] = EmptyCell
						if !c.ValidRow(row, val) || !c.ValidCol(col, val) || !c.ValidBox(row, col, val) {
							t.Errorf("invalid solution at cell (%d,%d)", row, col)
						}
						c[row][col] = val
					}
				}
			} else {
				if !test.input.Equal(c) {
					t.Error("non unique solution board overwritten")
				}
			}
		})
	}
}

func TestGeneratePuzzle(t *testing.T) {
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
