package sudoku

import "testing"

func TestValidBounds(t *testing.T) {
	var tests = []struct {
		name  string
		input [2]int
		want  bool
	}{
		{
			name:  "top-left corner",
			input: [2]int{0, 0},
			want:  true,
		},
		{
			name:  "bottom-right corner",
			input: [2]int{8, 8},
			want:  true,
		},
		{
			name:  "center position",
			input: [2]int{4, 4},
			want:  true,
		},
		{
			name:  "top-right corner",
			input: [2]int{0, 8},
			want:  true,
		},
		{
			name:  "bottom-left corner",
			input: [2]int{8, 0},
			want:  true,
		},
		{
			name:  "negative row",
			input: [2]int{-1, 4},
			want:  false,
		},
		{
			name:  "negative column",
			input: [2]int{4, -1},
			want:  false,
		},
		{
			name:  "both negative",
			input: [2]int{-1, -1},
			want:  false,
		},
		{
			name:  "both too large",
			input: [2]int{9, 9},
			want:  false,
		},
		{
			name:  "just out of bounds row",
			input: [2]int{9, 0},
			want:  false,
		},
		{
			name:  "just out of bounds column",
			input: [2]int{0, 9},
			want:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := validBounds(test.input[0], test.input[1])
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
		{
			name:  "smallest value",
			input: minValue,
			want:  true,
		},
		{
			name:  "biggest value",
			input: maxValue,
			want:  true,
		},
		{
			name:  "empty value",
			input: emptyCell,
			want:  false,
		},
		{
			name:  "just too big value",
			input: maxValue + 1,
			want:  false,
		},
		{
			name:  "negative value",
			input: -1,
			want:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := validVal(test.input)
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

	var tests = []struct {
		name  string
		input *input
		want  bool
	}{
		{
			name: "empty board",
			input: &input{
				s:   &Sudoku{},
				row: 0,
				val: 1,
			},
			want: true,
		},
		{
			name: "partially filled row",
			input: &input{
				s: &Sudoku{
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{1, 2, 3, 0, 0, 0, 0, 0, 0},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
				},
				row: 2,
				val: 4,
			},
			want: true,
		},
		{
			name: "almost empty row",
			input: &input{
				s: &Sudoku{
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{0, 2, 0, 0, 0, 0, 0, 0, 0},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
				},
				row: 2,
				val: 2,
			},
			want: false,
		},
		{
			name: "almost filled row",
			input: &input{
				s: &Sudoku{
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{1, 2, 3, 4, 5, 6, 7, 8, 0},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
				},
				row: 5,
				val: 9,
			},
			want: true,
		},
		{
			name: "filled row",
			input: &input{
				s: &Sudoku{
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{1, 2, 3, 4, 5, 6, 7, 8, 9},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
				},
				row: 6,
				val: 3,
			},
			want: false,
		},
		{
			name: "last row",
			input: &input{
				s: &Sudoku{
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{0, 0, 0, 0, 5, 0, 0, 0, 0},
				},
				row: 8,
				val: 5,
			},
			want: false,
		},
		{
			name: "first row",
			input: &input{
				s: &Sudoku{
					{0, 0, 0, 0, 0, 0, 0, 0, 1},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
				},
				row: 0,
				val: 1,
			},
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.input.s.validRow(test.input.row, test.input.val)
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

	var tests = []struct {
		name  string
		input *input
		want  bool
	}{
		{
			name: "empty board",
			input: &input{
				s:   &Sudoku{},
				col: 4,
				val: 1,
			},
			want: true,
		},
		{
			name: "partially filled column",
			input: &input{
				s: &Sudoku{
					{5, 5, 1, 5, 5, 5, 5, 5, 5},
					{5, 5, 2, 5, 5, 5, 5, 5, 5},
					{5, 5, 3, 5, 5, 5, 5, 5, 5},
					{5, 5, 0, 5, 5, 5, 5, 5, 5},
					{5, 5, 0, 5, 5, 5, 5, 5, 5},
					{5, 5, 0, 5, 5, 5, 5, 5, 5},
					{5, 5, 0, 5, 5, 5, 5, 5, 5},
					{5, 5, 0, 5, 5, 5, 5, 5, 5},
					{5, 5, 0, 5, 5, 5, 5, 5, 5},
				},
				col: 2,
				val: 4,
			},
			want: true,
		},
		{
			name: "almost empty column",
			input: &input{
				s: &Sudoku{
					{5, 5, 0, 5, 5, 5, 5, 5, 5},
					{5, 5, 2, 5, 5, 5, 5, 5, 5},
					{5, 5, 0, 5, 5, 5, 5, 5, 5},
					{5, 5, 0, 5, 5, 5, 5, 5, 5},
					{5, 5, 0, 5, 5, 5, 5, 5, 5},
					{5, 5, 0, 5, 5, 5, 5, 5, 5},
					{5, 5, 0, 5, 5, 5, 5, 5, 5},
					{5, 5, 0, 5, 5, 5, 5, 5, 5},
					{5, 5, 0, 5, 5, 5, 5, 5, 5},
				},
				col: 2,
				val: 2,
			},
			want: false,
		},
		{
			name: "almost filled column",
			input: &input{
				s: &Sudoku{
					{5, 5, 1, 5, 5, 5, 5, 5, 5},
					{5, 5, 2, 5, 5, 5, 5, 5, 5},
					{5, 5, 3, 5, 5, 5, 5, 5, 5},
					{5, 5, 4, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 6, 5, 5, 5, 5, 5, 5},
					{5, 5, 7, 5, 5, 5, 5, 5, 5},
					{5, 5, 8, 5, 5, 5, 5, 5, 5},
					{5, 5, 0, 5, 5, 5, 5, 5, 5},
				},
				col: 2,
				val: 9,
			},
			want: true,
		},
		{
			name: "filled column",
			input: &input{
				s: &Sudoku{
					{5, 5, 1, 5, 5, 5, 5, 5, 5},
					{5, 5, 2, 5, 5, 5, 5, 5, 5},
					{5, 5, 3, 5, 5, 5, 5, 5, 5},
					{5, 5, 4, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 6, 5, 5, 5, 5, 5, 5},
					{5, 5, 7, 5, 5, 5, 5, 5, 5},
					{5, 5, 8, 5, 5, 5, 5, 5, 5},
					{5, 5, 9, 5, 5, 5, 5, 5, 5},
				},
				col: 2,
				val: 3,
			},
			want: false,
		},
		{
			name: "last column",
			input: &input{
				s: &Sudoku{
					{5, 5, 5, 5, 5, 5, 5, 5, 0},
					{5, 5, 5, 5, 5, 5, 5, 5, 0},
					{5, 5, 5, 5, 5, 5, 5, 5, 0},
					{5, 5, 5, 5, 5, 5, 5, 5, 0},
					{5, 5, 5, 5, 5, 5, 5, 5, 5},
					{5, 5, 5, 5, 5, 5, 5, 5, 0},
					{5, 5, 5, 5, 5, 5, 5, 5, 0},
					{5, 5, 5, 5, 5, 5, 5, 5, 0},
					{5, 5, 5, 5, 5, 5, 5, 5, 0},
				},
				col: 8,
				val: 5,
			},
			want: false,
		},
		{
			name: "first column",
			input: &input{
				s: &Sudoku{
					{0, 5, 5, 5, 5, 5, 5, 5, 5},
					{0, 5, 5, 5, 5, 5, 5, 5, 5},
					{0, 5, 5, 5, 5, 5, 5, 5, 5},
					{0, 5, 5, 5, 5, 5, 5, 5, 5},
					{0, 5, 5, 5, 5, 5, 5, 5, 5},
					{0, 5, 5, 5, 5, 5, 5, 5, 5},
					{0, 5, 5, 5, 5, 5, 5, 5, 5},
					{0, 5, 5, 5, 5, 5, 5, 5, 5},
					{1, 5, 5, 5, 5, 5, 5, 5, 5},
				},
				col: 0,
				val: 1,
			},
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.input.s.validCol(test.input.col, test.input.val)
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

	var tests = []struct {
		name  string
		input *input
		want  bool
	}{
		{
			name: "empty board - top-left box",
			input: &input{
				s:   &Sudoku{},
				row: 0,
				col: 0,
				val: 5,
			},
			want: true,
		},
		{
			name: "empty board - center box",
			input: &input{
				s:   &Sudoku{},
				row: 4,
				col: 4,
				val: 1,
			},
			want: true,
		},
		{
			name: "empty board - bottom-right box",
			input: &input{
				s:   &Sudoku{},
				row: 8,
				col: 8,
				val: 9,
			},
			want: true,
		},
		{
			name: "value exists in same box - top-left",
			input: &input{
				s: &Sudoku{
					{1, 2, 3, 0, 0, 0, 0, 0, 0},
					{4, 5, 6, 0, 0, 0, 0, 0, 0},
					{7, 8, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
				row: 2,
				col: 2,
				val: 5,
			},
			want: false,
		},
		{
			name: "value exists in center of box",
			input: &input{
				s: &Sudoku{
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 7, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
				row: 3,
				col: 3,
				val: 7,
			},
			want: false,
		},
		{
			name: "value exists in different box",
			input: &input{
				s: &Sudoku{
					{1, 2, 3, 0, 0, 0, 0, 0, 0},
					{4, 5, 6, 0, 0, 0, 0, 0, 0},
					{7, 8, 9, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
				row: 3,
				col: 3,
				val: 5,
			},
			want: true,
		},
		{
			name: "value in same row and column but different box",
			input: &input{
				s: &Sudoku{
					{0, 0, 0, 0, 0, 0, 7, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
				row: 0,
				col: 0,
				val: 7,
			},
			want: true,
		},
		{
			name: "top-middle box",
			input: &input{
				s: &Sudoku{
					{0, 0, 0, 1, 2, 3, 0, 0, 0},
					{0, 0, 0, 4, 5, 6, 0, 0, 0},
					{0, 0, 0, 7, 8, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
				row: 2,
				col: 5,
				val: 5,
			},
			want: false,
		},
		{
			name: "top-right box",
			input: &input{
				s: &Sudoku{
					{0, 0, 0, 0, 0, 0, 1, 2, 3},
					{0, 0, 0, 0, 0, 0, 4, 5, 6},
					{0, 0, 0, 0, 0, 0, 7, 8, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
				row: 2,
				col: 8,
				val: 9,
			},
			want: true,
		},
		{
			name: "middle-left box",
			input: &input{
				s: &Sudoku{
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{1, 2, 3, 0, 0, 0, 0, 0, 0},
					{4, 5, 6, 0, 0, 0, 0, 0, 0},
					{7, 8, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
				row: 5,
				col: 2,
				val: 2,
			},
			want: false,
		},
		{
			name: "middle-right box",
			input: &input{
				s: &Sudoku{
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 1, 2, 3},
					{0, 0, 0, 0, 0, 0, 4, 5, 6},
					{0, 0, 0, 0, 0, 0, 7, 8, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
				row: 5,
				col: 8,
				val: 9,
			},
			want: true,
		},
		{
			name: "bottom-left box",
			input: &input{
				s: &Sudoku{
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{1, 2, 3, 0, 0, 0, 0, 0, 0},
					{4, 5, 6, 0, 0, 0, 0, 0, 0},
					{7, 8, 0, 0, 0, 0, 0, 0, 0},
				},
				row: 8,
				col: 2,
				val: 9,
			},
			want: true,
		},
		{
			name: "bottom-middle box",
			input: &input{
				s: &Sudoku{
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 1, 2, 3, 0, 0, 0},
					{0, 0, 0, 4, 5, 6, 0, 0, 0},
					{0, 0, 0, 7, 8, 0, 0, 0, 0},
				},
				row: 8,
				col: 5,
				val: 5,
			},
			want: false,
		},
		{
			name: "bottom-right box",
			input: &input{
				s: &Sudoku{
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 1, 2, 3},
					{0, 0, 0, 0, 0, 0, 4, 5, 6},
					{0, 0, 0, 0, 0, 0, 7, 8, 0},
				},
				row: 8,
				col: 8,
				val: 9,
			},
			want: true,
		},
		{
			name: "box boundary calculation - position (2,2)",
			input: &input{
				s: &Sudoku{
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 3, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
				row: 0,
				col: 0,
				val: 3,
			},
			want: false,
		},
		{
			name: "box boundary calculation - position (3,3)",
			input: &input{
				s: &Sudoku{
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 3, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
				row: 3,
				col: 3,
				val: 3,
			},
			want: true,
		},
		{
			name: "almost full box - one slot remaining",
			input: &input{
				s: &Sudoku{
					{1, 2, 3, 0, 0, 0, 0, 0, 0},
					{4, 5, 6, 0, 0, 0, 0, 0, 0},
					{7, 8, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
				row: 2,
				col: 2,
				val: 9,
			},
			want: true,
		},
		{
			name: "full box - conflict",
			input: &input{
				s: &Sudoku{
					{1, 2, 3, 0, 0, 0, 0, 0, 0},
					{4, 5, 6, 0, 0, 0, 0, 0, 0},
					{7, 8, 9, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
				row: 1,
				col: 1,
				val: 1,
			},
			want: false,
		},
		{
			name: "corner positions - top-left corner",
			input: &input{
				s: &Sudoku{
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 5, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
				row: 0,
				col: 0,
				val: 5,
			},
			want: false,
		},
		{
			name: "corner positions - bottom-right corner",
			input: &input{
				s: &Sudoku{
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 6, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
				row: 8,
				col: 8,
				val: 6,
			},
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.input.s.validBox(
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
