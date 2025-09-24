package sudoku

import "testing"

func TestIs(t *testing.T) {
	t.Run("exact same boards", func(t *testing.T) {
		org := &Sudoku{
			{5, 3, 4, 6, 7, 8, 9, 1, 2},
			{6, 7, 2, 1, 9, 5, 3, 4, 8},
			{1, 9, 8, 3, 4, 2, 5, 6, 7},
			{8, 5, 9, 7, 6, 1, 4, 2, 3},
			{4, 2, 6, 8, 5, 3, 7, 9, 1},
			{7, 1, 3, 9, 2, 4, 8, 5, 6},
			{9, 6, 1, 5, 3, 7, 2, 8, 4},
			{2, 8, 7, 4, 1, 9, 6, 3, 5},
			{3, 4, 5, 2, 8, 6, 1, 7, 9},
		}
		comp := &Sudoku{
			{5, 3, 4, 6, 7, 8, 9, 1, 2},
			{6, 7, 2, 1, 9, 5, 3, 4, 8},
			{1, 9, 8, 3, 4, 2, 5, 6, 7},
			{8, 5, 9, 7, 6, 1, 4, 2, 3},
			{4, 2, 6, 8, 5, 3, 7, 9, 1},
			{7, 1, 3, 9, 2, 4, 8, 5, 6},
			{9, 6, 1, 5, 3, 7, 2, 8, 4},
			{2, 8, 7, 4, 1, 9, 6, 3, 5},
			{3, 4, 5, 2, 8, 6, 1, 7, 9},
		}

		if !org.Is(comp) {
			t.Error("Expected identical boards to be equal")
		}
		if !comp.Is(org) {
			t.Error("Expected identical boards to be bidirectional equal")
		}
	})

	t.Run("single number different", func(t *testing.T) {
		org := &Sudoku{
			{5, 3, 4, 6, 7, 8, 9, 1, 2},
			{6, 7, 2, 1, 9, 5, 3, 4, 8},
			{1, 9, 8, 3, 4, 2, 5, 6, 7},
			{8, 5, 9, 7, 6, 1, 4, 2, 3},
			{4, 2, 6, 8, 5, 3, 7, 9, 1},
			{7, 1, 3, 9, 2, 4, 8, 5, 6},
			{9, 6, 1, 5, 3, 7, 2, 8, 4},
			{2, 8, 7, 4, 1, 9, 6, 3, 5},
			{3, 4, 5, 2, 8, 6, 1, 7, 9},
		}
		comp := &Sudoku{
			{9, 3, 4, 6, 7, 8, 9, 1, 2}, // Changed first cell from 5 to 9
			{6, 7, 2, 1, 9, 5, 3, 4, 8},
			{1, 9, 8, 3, 4, 2, 5, 6, 7},
			{8, 5, 9, 7, 6, 1, 4, 2, 3},
			{4, 2, 6, 8, 5, 3, 7, 9, 1},
			{7, 1, 3, 9, 2, 4, 8, 5, 6},
			{9, 6, 1, 5, 3, 7, 2, 8, 4},
			{2, 8, 7, 4, 1, 9, 6, 3, 5},
			{3, 4, 5, 2, 8, 6, 1, 7, 9},
		}

		if org.Is(comp) {
			t.Error("Expected boards with single different number to be different")
		}
		if comp.Is(org) {
			t.Error("Expected boards with single different number to be bidirectional different")
		}
	})

	t.Run("two rows flipped", func(t *testing.T) {
		org := &Sudoku{
			{5, 3, 4, 6, 7, 8, 9, 1, 2},
			{6, 7, 2, 1, 9, 5, 3, 4, 8},
			{1, 9, 8, 3, 4, 2, 5, 6, 7},
			{8, 5, 9, 7, 6, 1, 4, 2, 3},
			{4, 2, 6, 8, 5, 3, 7, 9, 1},
			{7, 1, 3, 9, 2, 4, 8, 5, 6},
			{9, 6, 1, 5, 3, 7, 2, 8, 4},
			{2, 8, 7, 4, 1, 9, 6, 3, 5},
			{3, 4, 5, 2, 8, 6, 1, 7, 9},
		}
		comp := &Sudoku{
			{6, 7, 2, 1, 9, 5, 3, 4, 8}, // Row 0 and 1 swapped
			{5, 3, 4, 6, 7, 8, 9, 1, 2},
			{1, 9, 8, 3, 4, 2, 5, 6, 7},
			{8, 5, 9, 7, 6, 1, 4, 2, 3},
			{4, 2, 6, 8, 5, 3, 7, 9, 1},
			{7, 1, 3, 9, 2, 4, 8, 5, 6},
			{9, 6, 1, 5, 3, 7, 2, 8, 4},
			{2, 8, 7, 4, 1, 9, 6, 3, 5},
			{3, 4, 5, 2, 8, 6, 1, 7, 9},
		}

		if org.Is(comp) {
			t.Error("Expected boards with flipped rows to be different")
		}
		if comp.Is(org) {
			t.Error("Expected boards with flipped rows to be bidirectional different")
		}
	})

	t.Run("two columns flipped", func(t *testing.T) {
		org := &Sudoku{
			{5, 3, 4, 6, 7, 8, 9, 1, 2},
			{6, 7, 2, 1, 9, 5, 3, 4, 8},
			{1, 9, 8, 3, 4, 2, 5, 6, 7},
			{8, 5, 9, 7, 6, 1, 4, 2, 3},
			{4, 2, 6, 8, 5, 3, 7, 9, 1},
			{7, 1, 3, 9, 2, 4, 8, 5, 6},
			{9, 6, 1, 5, 3, 7, 2, 8, 4},
			{2, 8, 7, 4, 1, 9, 6, 3, 5},
			{3, 4, 5, 2, 8, 6, 1, 7, 9},
		}
		comp := &Sudoku{
			{3, 5, 4, 6, 7, 8, 9, 1, 2}, // Column 0 and 1 swapped
			{7, 6, 2, 1, 9, 5, 3, 4, 8},
			{9, 1, 8, 3, 4, 2, 5, 6, 7},
			{5, 8, 9, 7, 6, 1, 4, 2, 3},
			{2, 4, 6, 8, 5, 3, 7, 9, 1},
			{1, 7, 3, 9, 2, 4, 8, 5, 6},
			{6, 9, 1, 5, 3, 7, 2, 8, 4},
			{8, 2, 7, 4, 1, 9, 6, 3, 5},
			{4, 3, 5, 2, 8, 6, 1, 7, 9},
		}

		if org.Is(comp) {
			t.Error("Expected boards with flipped columns to be different")
		}
		if comp.Is(org) {
			t.Error("Expected boards with flipped columns to be bidirectional different")
		}
	})

	t.Run("two boxes switched", func(t *testing.T) {
		org := &Sudoku{
			{5, 3, 4, 6, 7, 8, 9, 1, 2},
			{6, 7, 2, 1, 9, 5, 3, 4, 8},
			{1, 9, 8, 3, 4, 2, 5, 6, 7},
			{8, 5, 9, 7, 6, 1, 4, 2, 3},
			{4, 2, 6, 8, 5, 3, 7, 9, 1},
			{7, 1, 3, 9, 2, 4, 8, 5, 6},
			{9, 6, 1, 5, 3, 7, 2, 8, 4},
			{2, 8, 7, 4, 1, 9, 6, 3, 5},
			{3, 4, 5, 2, 8, 6, 1, 7, 9},
		}
		comp := &Sudoku{
			{6, 7, 8, 5, 3, 4, 9, 1, 2}, // Top-left and top-middle boxes switched
			{1, 9, 5, 6, 7, 2, 3, 4, 8},
			{3, 4, 2, 1, 9, 8, 5, 6, 7},
			{8, 5, 9, 7, 6, 1, 4, 2, 3},
			{4, 2, 6, 8, 5, 3, 7, 9, 1},
			{7, 1, 3, 9, 2, 4, 8, 5, 6},
			{9, 6, 1, 5, 3, 7, 2, 8, 4},
			{2, 8, 7, 4, 1, 9, 6, 3, 5},
			{3, 4, 5, 2, 8, 6, 1, 7, 9},
		}

		if org.Is(comp) {
			t.Error("Expected boards with switched boxes to be different")
		}
		if comp.Is(org) {
			t.Error("Expected boards with switched boxes to be bidirectional different")
		}
	})

	t.Run("only last number different", func(t *testing.T) {
		org := &Sudoku{
			{5, 3, 4, 6, 7, 8, 9, 1, 2},
			{6, 7, 2, 1, 9, 5, 3, 4, 8},
			{1, 9, 8, 3, 4, 2, 5, 6, 7},
			{8, 5, 9, 7, 6, 1, 4, 2, 3},
			{4, 2, 6, 8, 5, 3, 7, 9, 1},
			{7, 1, 3, 9, 2, 4, 8, 5, 6},
			{9, 6, 1, 5, 3, 7, 2, 8, 4},
			{2, 8, 7, 4, 1, 9, 6, 3, 5},
			{3, 4, 5, 2, 8, 6, 1, 7, 9},
		}
		comp := &Sudoku{
			{5, 3, 4, 6, 7, 8, 9, 1, 2},
			{6, 7, 2, 1, 9, 5, 3, 4, 8},
			{1, 9, 8, 3, 4, 2, 5, 6, 7},
			{8, 5, 9, 7, 6, 1, 4, 2, 3},
			{4, 2, 6, 8, 5, 3, 7, 9, 1},
			{7, 1, 3, 9, 2, 4, 8, 5, 6},
			{9, 6, 1, 5, 3, 7, 2, 8, 4},
			{2, 8, 7, 4, 1, 9, 6, 3, 5},
			{3, 4, 5, 2, 8, 6, 1, 7, 1}, // Last cell changed from 9 to 1
		}

		if org.Is(comp) {
			t.Error("Expected boards with only last number different to be different")
		}
		if comp.Is(org) {
			t.Error("Expected boards with only last number different to be bidirectional different")
		}
	})

	t.Run("every number different", func(t *testing.T) {
		org := &Sudoku{
			{5, 3, 4, 6, 7, 8, 9, 1, 2},
			{6, 7, 2, 1, 9, 5, 3, 4, 8},
			{1, 9, 8, 3, 4, 2, 5, 6, 7},
			{8, 5, 9, 7, 6, 1, 4, 2, 3},
			{4, 2, 6, 8, 5, 3, 7, 9, 1},
			{7, 1, 3, 9, 2, 4, 8, 5, 6},
			{9, 6, 1, 5, 3, 7, 2, 8, 4},
			{2, 8, 7, 4, 1, 9, 6, 3, 5},
			{3, 4, 5, 2, 8, 6, 1, 7, 9},
		}
		comp := &Sudoku{
			{1, 2, 3, 4, 5, 6, 7, 8, 9},
			{2, 3, 4, 5, 6, 7, 8, 9, 1},
			{3, 4, 5, 6, 7, 8, 9, 1, 2},
			{4, 5, 6, 7, 8, 9, 1, 2, 3},
			{5, 6, 7, 8, 9, 1, 2, 3, 4},
			{6, 7, 8, 9, 1, 2, 3, 4, 5},
			{7, 8, 9, 1, 2, 3, 4, 5, 6},
			{8, 9, 1, 2, 3, 4, 5, 6, 7},
			{9, 1, 2, 3, 4, 5, 6, 7, 8},
		}

		if org.Is(comp) {
			t.Error("Expected boards with every number different to be different")
		}
		if comp.Is(org) {
			t.Error("Expected boards with every number different to be bidriectional different")
		}
	})

	t.Run("empty boards", func(t *testing.T) {
		org := New()
		comp := New()

		if !org.Is(comp) {
			t.Error("Expected empty boards to be equal")
		}
		if !comp.Is(org) {
			t.Error("Expected empty boards to be bidirectional equal")
		}
	})
}

func TestValidVal(t *testing.T) {
	t.Run("min value", func(t *testing.T) {
		if !validVal(MinValue) {
			t.Error("Expected min value to be valid")
		}
	})

	t.Run("at max value border", func(t *testing.T) {
		if !validVal(MaxValue) {
			t.Error("Expected max value to be valid")
		}
	})

	t.Run("over max value border", func(t *testing.T) {
		if validVal(MaxValue + 1) {
			t.Error("Expected max value + 1 to be invalid")
		}
	})

	t.Run("high number", func(t *testing.T) {
		if validVal(10000) {
			t.Error("Expected 10000 to be invalid")
		}
	})

	t.Run("negative number", func(t *testing.T) {
		if validVal(-1) {
			t.Error("Expected -1 to be invalid")
		}
	})
}

func TestValidRow(t *testing.T) {
	t.Run("valid value in row", func(t *testing.T) {
		s := &Sudoku{
			{1, 2, 3, 4, 5, 6, 7, 8, 0}, // Row 0 missing 9
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
		}

		if !s.validRow(0, 9) {
			t.Error("Expected 9 to be valid in row 0 (9 is not present)")
		}
	})

	t.Run("invalid value already in row", func(t *testing.T) {
		s := &Sudoku{
			{1, 2, 3, 4, 5, 6, 7, 8, 9}, // Row 0 has all numbers 1-9
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
		}

		if s.validRow(0, 5) {
			t.Error("Expected 5 to be invalid in row 0 (5 is already present)")
		}
	})

	t.Run("valid value in non-first row", func(t *testing.T) {
		s := &Sudoku{
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{1, 2, 3, 4, 5, 6, 7, 8, 0}, // Row 3 missing 9
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
		}

		if !s.validRow(3, 9) {
			t.Error("Expected 9 to be valid in row 3 (9 is not present)")
		}
	})

	t.Run("invalid value in non-first row", func(t *testing.T) {
		s := &Sudoku{
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{2, 4, 6, 8, 1, 3, 5, 7, 9}, // Row 5 has all numbers 1-9
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
		}

		if s.validRow(5, 7) {
			t.Error("Expected 7 to be invalid in row 5 (7 is already present)")
		}
	})
}

func TestValidCol(t *testing.T) {
	t.Run("valid value in column", func(t *testing.T) {
		s := &Sudoku{
			{1, 0, 0, 0, 0, 0, 0, 0, 0},
			{2, 0, 0, 0, 0, 0, 0, 0, 0},
			{3, 0, 0, 0, 0, 0, 0, 0, 0},
			{4, 0, 0, 0, 0, 0, 0, 0, 0},
			{5, 0, 0, 0, 0, 0, 0, 0, 0},
			{6, 0, 0, 0, 0, 0, 0, 0, 0},
			{7, 0, 0, 0, 0, 0, 0, 0, 0},
			{8, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0}, // Column 0 missing 9
		}

		if !s.validCol(0, 9) {
			t.Error("Expected 9 to be valid in column 0 (9 is not present)")
		}
	})

	t.Run("invalid value already in column", func(t *testing.T) {
		s := &Sudoku{
			{1, 0, 0, 0, 0, 0, 0, 0, 0},
			{2, 0, 0, 0, 0, 0, 0, 0, 0},
			{3, 0, 0, 0, 0, 0, 0, 0, 0},
			{4, 0, 0, 0, 0, 0, 0, 0, 0},
			{5, 0, 0, 0, 0, 0, 0, 0, 0},
			{6, 0, 0, 0, 0, 0, 0, 0, 0},
			{7, 0, 0, 0, 0, 0, 0, 0, 0},
			{8, 0, 0, 0, 0, 0, 0, 0, 0},
			{9, 0, 0, 0, 0, 0, 0, 0, 0}, // Column 0 has all numbers 1-9
		}

		if s.validCol(0, 5) {
			t.Error("Expected 5 to be invalid in column 0 (5 is already present)")
		}
	})

	t.Run("valid value in non-first column", func(t *testing.T) {
		s := &Sudoku{
			{0, 0, 0, 1, 0, 0, 0, 0, 0},
			{0, 0, 0, 2, 0, 0, 0, 0, 0},
			{0, 0, 0, 3, 0, 0, 0, 0, 0},
			{0, 0, 0, 4, 0, 0, 0, 0, 0},
			{0, 0, 0, 5, 0, 0, 0, 0, 0},
			{0, 0, 0, 6, 0, 0, 0, 0, 0},
			{0, 0, 0, 7, 0, 0, 0, 0, 0},
			{0, 0, 0, 8, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0}, // Column 3 missing 9
		}

		if !s.validCol(3, 9) {
			t.Error("Expected 9 to be valid in column 3 (9 is not present)")
		}
	})

	t.Run("invalid value in non-first column", func(t *testing.T) {
		s := &Sudoku{
			{0, 0, 0, 0, 0, 0, 0, 2, 0},
			{0, 0, 0, 0, 0, 0, 0, 4, 0},
			{0, 0, 0, 0, 0, 0, 0, 6, 0},
			{0, 0, 0, 0, 0, 0, 0, 8, 0},
			{0, 0, 0, 0, 0, 0, 0, 1, 0},
			{0, 0, 0, 0, 0, 0, 0, 3, 0},
			{0, 0, 0, 0, 0, 0, 0, 5, 0},
			{0, 0, 0, 0, 0, 0, 0, 7, 0},
			{0, 0, 0, 0, 0, 0, 0, 9, 0}, // Column 7 has all numbers 1-9
		}

		if s.validCol(7, 7) {
			t.Error("Expected 7 to be invalid in column 7 (7 is already present)")
		}
	})
}

func TestValidBox(t *testing.T) {
	t.Run("valid value in box", func(t *testing.T) {
		s := &Sudoku{
			{1, 2, 3, 0, 0, 0, 0, 0, 0},
			{4, 5, 6, 0, 0, 0, 0, 0, 0},
			{7, 8, 0, 0, 0, 0, 0, 0, 0}, // Top-left box missing 9
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
		}

		if !s.validBox(2, 2, 9) {
			t.Error("Expected 9 to be valid in top-left box (9 is not present)")
		}
	})

	t.Run("invalid value already in box", func(t *testing.T) {
		s := &Sudoku{
			{1, 2, 3, 0, 0, 0, 0, 0, 0},
			{4, 5, 6, 0, 0, 0, 0, 0, 0},
			{7, 8, 9, 0, 0, 0, 0, 0, 0}, // Top-left box has all numbers 1-9
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
		}

		if s.validBox(1, 1, 5) {
			t.Error("Expected 5 to be invalid in top-left box (5 is already present)")
		}
	})

	t.Run("valid value in non-first box", func(t *testing.T) {
		s := &Sudoku{
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 1, 2, 3},
			{0, 0, 0, 0, 0, 0, 4, 5, 6},
			{0, 0, 0, 0, 0, 0, 7, 8, 0}, // Bottom-right box missing 9
		}

		if !s.validBox(8, 8, 9) {
			t.Error("Expected 9 to be valid in bottom-right box (9 is not present)")
		}
	})

	t.Run("invalid value in non-first box", func(t *testing.T) {
		s := &Sudoku{
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 1, 2, 3, 0, 0, 0},
			{0, 0, 0, 4, 5, 6, 0, 0, 0},
			{0, 0, 0, 7, 8, 9, 0, 0, 0}, // Middle box has all numbers 1-9
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
		}

		if s.validBox(4, 4, 8) {
			t.Error("Expected 8 to be invalid in middle box (8 is already present)")
		}
	})
}

func TestComplete(t *testing.T) {
	t.Run("completely filled valid sudoku", func(t *testing.T) {
		s := &Sudoku{
			{5, 3, 4, 6, 7, 8, 9, 1, 2},
			{6, 7, 2, 1, 9, 5, 3, 4, 8},
			{1, 9, 8, 3, 4, 2, 5, 6, 7},
			{8, 5, 9, 7, 6, 1, 4, 2, 3},
			{4, 2, 6, 8, 5, 3, 7, 9, 1},
			{7, 1, 3, 9, 2, 4, 8, 5, 6},
			{9, 6, 1, 5, 3, 7, 2, 8, 4},
			{2, 8, 7, 4, 1, 9, 6, 3, 5},
			{3, 4, 5, 2, 8, 6, 1, 7, 9},
		}

		if !s.Complete() {
			t.Error("Expected completely filled valid sudoku to be complete")
		}
	})

	t.Run("completely empty sudoku", func(t *testing.T) {
		s := &Sudoku{
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
		}

		if s.Complete() {
			t.Error("Expected completely empty sudoku to be incomplete")
		}
	})

	t.Run("single empty cell at start", func(t *testing.T) {
		s := &Sudoku{
			{0, 3, 4, 6, 7, 8, 9, 1, 2}, // First cell empty
			{6, 7, 2, 1, 9, 5, 3, 4, 8},
			{1, 9, 8, 3, 4, 2, 5, 6, 7},
			{8, 5, 9, 7, 6, 1, 4, 2, 3},
			{4, 2, 6, 8, 5, 3, 7, 9, 1},
			{7, 1, 3, 9, 2, 4, 8, 5, 6},
			{9, 6, 1, 5, 3, 7, 2, 8, 4},
			{2, 8, 7, 4, 1, 9, 6, 3, 5},
			{3, 4, 5, 2, 8, 6, 1, 7, 9},
		}

		if s.Complete() {
			t.Error("Expected sudoku with single empty cell at start to be incomplete")
		}
	})

	t.Run("single empty cell in middle", func(t *testing.T) {
		s := &Sudoku{
			{5, 3, 4, 6, 7, 8, 9, 1, 2},
			{6, 7, 2, 1, 9, 5, 3, 4, 8},
			{1, 9, 8, 3, 4, 2, 5, 6, 7},
			{8, 5, 9, 7, 6, 1, 4, 2, 3},
			{4, 2, 6, 8, 0, 3, 7, 9, 1}, // Middle cell empty
			{7, 1, 3, 9, 2, 4, 8, 5, 6},
			{9, 6, 1, 5, 3, 7, 2, 8, 4},
			{2, 8, 7, 4, 1, 9, 6, 3, 5},
			{3, 4, 5, 2, 8, 6, 1, 7, 9},
		}

		if s.Complete() {
			t.Error("Expected sudoku with single empty cell in middle to be incomplete")
		}
	})

	t.Run("single empty cell at end", func(t *testing.T) {
		s := &Sudoku{
			{5, 3, 4, 6, 7, 8, 9, 1, 2},
			{6, 7, 2, 1, 9, 5, 3, 4, 8},
			{1, 9, 8, 3, 4, 2, 5, 6, 7},
			{8, 5, 9, 7, 6, 1, 4, 2, 3},
			{4, 2, 6, 8, 5, 3, 7, 9, 1},
			{7, 1, 3, 9, 2, 4, 8, 5, 6},
			{9, 6, 1, 5, 3, 7, 2, 8, 4},
			{2, 8, 7, 4, 1, 9, 6, 3, 5},
			{3, 4, 5, 2, 8, 6, 1, 7, 0}, // Last cell empty
		}

		if s.Complete() {
			t.Error("Expected sudoku with single empty cell at end to be incomplete")
		}
	})

	t.Run("multiple empty cells scattered", func(t *testing.T) {
		s := &Sudoku{
			{5, 0, 4, 6, 7, 8, 9, 1, 2},
			{6, 7, 2, 1, 9, 5, 3, 4, 8},
			{1, 9, 8, 3, 4, 2, 5, 6, 7},
			{8, 5, 9, 7, 6, 1, 4, 2, 3},
			{4, 2, 6, 8, 5, 3, 7, 9, 1},
			{7, 1, 3, 9, 2, 4, 8, 5, 6},
			{9, 6, 1, 5, 3, 7, 2, 8, 4},
			{2, 8, 7, 4, 1, 9, 6, 3, 5},
			{3, 4, 5, 2, 8, 6, 1, 0, 9}, // Two empty cells
		}

		if s.Complete() {
			t.Error("Expected sudoku with multiple empty cells to be incomplete")
		}
	})

	t.Run("board with multiple zeros in different positions", func(t *testing.T) {
		s := &Sudoku{
			{0, 3, 4, 6, 7, 8, 9, 1, 2}, // Zero at start
			{6, 7, 2, 1, 9, 5, 3, 4, 8},
			{1, 9, 8, 3, 4, 2, 5, 6, 7},
			{8, 5, 9, 7, 6, 1, 4, 2, 3},
			{4, 2, 6, 8, 0, 3, 7, 9, 1}, // Zero in middle
			{7, 1, 3, 9, 2, 4, 8, 5, 6},
			{9, 6, 1, 5, 3, 7, 2, 8, 4},
			{2, 8, 7, 4, 1, 9, 6, 3, 5},
			{3, 4, 5, 2, 8, 6, 1, 7, 0}, // Zero at end
		}

		if s.Complete() {
			t.Error("Expected board with multiple zeros to be incomplete")
		}
	})

	t.Run("board with zeros in corners", func(t *testing.T) {
		s := &Sudoku{
			{0, 3, 4, 6, 7, 8, 9, 1, 0}, // Zeros in top corners
			{6, 7, 2, 1, 9, 5, 3, 4, 8},
			{1, 9, 8, 3, 4, 2, 5, 6, 7},
			{8, 5, 9, 7, 6, 1, 4, 2, 3},
			{4, 2, 6, 8, 5, 3, 7, 9, 1},
			{7, 1, 3, 9, 2, 4, 8, 5, 6},
			{9, 6, 1, 5, 3, 7, 2, 8, 4},
			{2, 8, 7, 4, 1, 9, 6, 3, 5},
			{0, 4, 5, 2, 8, 6, 1, 7, 0}, // Zeros in bottom corners
		}

		if s.Complete() {
			t.Error("Expected board with zeros in corners to be incomplete")
		}
	})

	t.Run("board with single zero in each row", func(t *testing.T) {
		s := &Sudoku{
			{0, 3, 4, 6, 7, 8, 9, 1, 2},
			{6, 0, 2, 1, 9, 5, 3, 4, 8},
			{1, 9, 0, 3, 4, 2, 5, 6, 7},
			{8, 5, 9, 0, 6, 1, 4, 2, 3},
			{4, 2, 6, 8, 0, 3, 7, 9, 1},
			{7, 1, 3, 9, 2, 0, 8, 5, 6},
			{9, 6, 1, 5, 3, 7, 0, 8, 4},
			{2, 8, 7, 4, 1, 9, 6, 0, 5},
			{3, 4, 5, 2, 8, 6, 1, 7, 0},
		}

		if s.Complete() {
			t.Error("Expected board with zeros in each row to be incomplete")
		}
	})

	t.Run("board with alternating pattern of zeros", func(t *testing.T) {
		s := &Sudoku{
			{1, 0, 1, 0, 1, 0, 1, 0, 1},
			{0, 1, 0, 1, 0, 1, 0, 1, 0},
			{1, 0, 1, 0, 1, 0, 1, 0, 1},
			{0, 1, 0, 1, 0, 1, 0, 1, 0},
			{1, 0, 1, 0, 1, 0, 1, 0, 1},
			{0, 1, 0, 1, 0, 1, 0, 1, 0},
			{1, 0, 1, 0, 1, 0, 1, 0, 1},
			{0, 1, 0, 1, 0, 1, 0, 1, 0},
			{1, 0, 1, 0, 1, 0, 1, 0, 1},
		}

		if s.Complete() {
			t.Error("Expected board with alternating zeros to be incomplete")
		}
	})

	t.Run("board filled with all same number", func(t *testing.T) {
		s := &Sudoku{
			{5, 5, 5, 5, 5, 5, 5, 5, 5},
			{5, 5, 5, 5, 5, 5, 5, 5, 5},
			{5, 5, 5, 5, 5, 5, 5, 5, 5},
			{5, 5, 5, 5, 5, 5, 5, 5, 5},
			{5, 5, 5, 5, 5, 5, 5, 5, 5},
			{5, 5, 5, 5, 5, 5, 5, 5, 5},
			{5, 5, 5, 5, 5, 5, 5, 5, 5},
			{5, 5, 5, 5, 5, 5, 5, 5, 5},
			{5, 5, 5, 5, 5, 5, 5, 5, 5},
		}

		if !s.Complete() {
			t.Error("Expected board with all non-zero values to be complete")
		}
	})

	t.Run("partially filled typical puzzle", func(t *testing.T) {
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

		if s.Complete() {
			t.Error("Expected partially filled sudoku to be incomplete")
		}
	})

	t.Run("all cells filled with 1 except one zero", func(t *testing.T) {
		s := &Sudoku{
			{1, 1, 1, 1, 1, 1, 1, 1, 1},
			{1, 1, 1, 1, 1, 1, 1, 1, 1},
			{1, 1, 1, 1, 1, 1, 1, 1, 1},
			{1, 1, 1, 1, 1, 1, 1, 1, 1},
			{1, 1, 1, 1, 0, 1, 1, 1, 1}, // One zero in the middle
			{1, 1, 1, 1, 1, 1, 1, 1, 1},
			{1, 1, 1, 1, 1, 1, 1, 1, 1},
			{1, 1, 1, 1, 1, 1, 1, 1, 1},
			{1, 1, 1, 1, 1, 1, 1, 1, 1},
		}

		if s.Complete() {
			t.Error("Expected board with single zero to be incomplete")
		}
	})
}

func TestFill(t *testing.T) {
	t.Run("fill generates complete sudoku for seeds 0-1000", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			s := New()
			s.Fill(int64(i))
			if !s.Complete() {
				t.Errorf("Fill with seed %d did not produce a complete sudoku", i)
			}
		}
	})
}

func TestInsert(t *testing.T) {
	t.Run("valid insertion into empty cell", func(t *testing.T) {
		s := New()
		_, err := s.Insert(0, 0, 5)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if s[0][0] != 5 {
			t.Errorf("Expected cell [0][0] to be 5, got %d", s[0][0])
		}
	})

	t.Run("valid insertion at boundary positions", func(t *testing.T) {
		s := New()

		// Test all four corners
		testCases := []struct {
			row, col, val int
		}{
			{0, 0, 1}, // top-left
			{0, 8, 2}, // top-right
			{8, 0, 3}, // bottom-left
			{8, 8, 4}, // bottom-right
		}

		for _, tc := range testCases {
			_, err := s.Insert(tc.row, tc.col, tc.val)
			if err != nil {
				t.Errorf("Expected no error for position [%d][%d], got: %v", tc.row, tc.col, err)
			}
			if s[tc.row][tc.col] != tc.val {
				t.Errorf("Expected cell [%d][%d] to be %d, got %d", tc.row, tc.col, tc.val, s[tc.row][tc.col])
			}
		}
	})

	t.Run("insert empty cell value", func(t *testing.T) {
		s := New()
		s[4][4] = 7 // Set a value first

		_, err := s.Insert(4, 4, EmptyCell)
		if err != nil {
			t.Errorf("Expected no error when inserting EmptyCell, got: %v", err)
		}
		if s[4][4] != EmptyCell {
			t.Errorf("Expected cell [4][4] to be %d, got %d", EmptyCell, s[4][4])
		}
	})

	t.Run("insert same value that already exists in cell", func(t *testing.T) {
		s := New()
		s[3][3] = 6

		_, err := s.Insert(3, 3, 6)
		if err != nil {
			t.Errorf("Expected no error when inserting same value, got: %v", err)
		}
		if s[3][3] != 6 {
			t.Errorf("Expected cell [3][3] to remain 6, got %d", s[3][3])
		}
	})

	t.Run("overwrite existing value with valid new value", func(t *testing.T) {
		s := New()
		s[2][2] = 3

		_, err := s.Insert(2, 2, 8)
		if err != nil {
			t.Errorf("Expected no error when overwriting with valid value, got: %v", err)
		}
		if s[2][2] != 8 {
			t.Errorf("Expected cell [2][2] to be 8, got %d", s[2][2])
		}
	})

	t.Run("row out of bounds - negative", func(t *testing.T) {
		s := New()
		_, err := s.Insert(-1, 0, 5)
		if err == nil {
			t.Error("Expected error for negative row, got nil")
		}
		if err.Error() != "position out of bounds" {
			t.Errorf("Expected 'position out of bounds' error, got: %v", err)
		}
	})

	t.Run("row out of bounds - too large", func(t *testing.T) {
		s := New()
		_, err := s.Insert(BoardSize, 0, 5)
		if err == nil {
			t.Error("Expected error for row >= BoardSize, got nil")
		}
		if err.Error() != "position out of bounds" {
			t.Errorf("Expected 'position out of bounds' error, got: %v", err)
		}
	})

	t.Run("column out of bounds - negative", func(t *testing.T) {
		s := New()
		_, err := s.Insert(0, -1, 5)
		if err == nil {
			t.Error("Expected error for negative column, got nil")
		}
		if err.Error() != "position out of bounds" {
			t.Errorf("Expected 'position out of bounds' error, got: %v", err)
		}
	})

	t.Run("column out of bounds - too large", func(t *testing.T) {
		s := New()
		_, err := s.Insert(0, BoardSize, 5)
		if err == nil {
			t.Error("Expected error for column >= BoardSize, got nil")
		}
		if err.Error() != "position out of bounds" {
			t.Errorf("Expected 'position out of bounds' error, got: %v", err)
		}
	})

	t.Run("both row and column out of bounds", func(t *testing.T) {
		s := New()
		_, err := s.Insert(-1, -1, 5)
		if err == nil {
			t.Error("Expected error for both row and column out of bounds, got nil")
		}
		if err.Error() != "position out of bounds" {
			t.Errorf("Expected 'position out of bounds' error, got: %v", err)
		}
	})

	t.Run("value below minimum", func(t *testing.T) {
		s := New()
		// MinValue is 1, so MinValue-1 is 0 (EmptyCell), which is allowed
		// Test with -1 instead to get a true invalid value
		_, err := s.Insert(0, 0, -1)
		if err == nil {
			t.Error("Expected error for value below minimum, got nil")
		}
	})

	t.Run("value above maximum", func(t *testing.T) {
		s := New()
		_, err := s.Insert(0, 0, MaxValue+1)
		if err == nil {
			t.Error("Expected error for value above maximum, got nil")
		}
	})

	t.Run("valid boundary values", func(t *testing.T) {
		s := New()

		// Test MinValue
		_, err := s.Insert(0, 0, MinValue)
		if err != nil {
			t.Errorf("Expected no error for MinValue (%d), got: %v", MinValue, err)
		}

		// Test MaxValue
		_, err = s.Insert(0, 1, MaxValue)
		if err != nil {
			t.Errorf("Expected no error for MaxValue (%d), got: %v", MaxValue, err)
		}
	})

	t.Run("value already exists in row", func(t *testing.T) {
		s := New()
		s[0][0] = 5 // Set value 5 in first cell of row 0

		_, err := s.Insert(0, 8, 5) // Try to insert 5 in last cell of same row
		if err == nil {
			t.Error("Expected error for value already in row, got nil")
		}
	})

	t.Run("value already exists in column", func(t *testing.T) {
		s := New()
		s[0][0] = 7 // Set value 7 in first cell of column 0

		_, err := s.Insert(8, 0, 7) // Try to insert 7 in last cell of same column
		if err == nil {
			t.Error("Expected error for value already in column, got nil")
		}
	})

	t.Run("value already exists in box", func(t *testing.T) {
		s := New()
		s[0][0] = 9 // Set value 9 in top-left box

		_, err := s.Insert(2, 2, 9) // Try to insert 9 in same box (bottom-right of top-left box)
		if err == nil {
			t.Error("Expected error for value already in box, got nil")
		}
	})

	t.Run("value exists in different box - should succeed", func(t *testing.T) {
		s := New()
		s[0][0] = 4 // Set value 4 in top-left box

		_, err := s.Insert(6, 6, 4) // Insert 4 in bottom-right box - should be valid
		if err != nil {
			t.Errorf("Expected no error for value in different box, got: %v", err)
		}
		if s[6][6] != 4 {
			t.Errorf("Expected cell [6][6] to be 4, got %d", s[6][6])
		}
	})

	t.Run("complex scenario - partially filled board", func(t *testing.T) {
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

		// Valid insertion
		_, err := s.Insert(0, 2, 4)
		if err != nil {
			t.Errorf("Expected no error for valid insertion, got: %v", err)
		}

		// Invalid - conflicts with row
		_, err = s.Insert(0, 3, 5) // 5 already exists in row 0
		if err == nil {
			t.Error("Expected error for row conflict, got nil")
		}

		// Invalid - conflicts with column
		_, err = s.Insert(1, 0, 5) // 5 already exists in column 0
		if err == nil {
			t.Error("Expected error for column conflict, got nil")
		}

		// Invalid - conflicts with box
		_, err = s.Insert(1, 1, 5) // 5 already exists in top-left box
		if err == nil {
			t.Error("Expected error for box conflict, got nil")
		}
	})

	t.Run("insert into all positions of a box", func(t *testing.T) {
		s := New()

		// Fill top-left box with values 1-9 (except position [2][2])
		values := [][]int{
			{1, 2, 3},
			{4, 5, 6},
			{7, 8, 0}, // Leave [2][2] empty
		}

		for r := 0; r < 3; r++ {
			for c := 0; c < 3; c++ {
				if values[r][c] != 0 {
					s[r][c] = values[r][c]
				}
			}
		}

		// Should be able to insert 9 in the empty position
		_, err := s.Insert(2, 2, 9)
		if err != nil {
			t.Errorf("Expected no error for valid box insertion, got: %v", err)
		}

		// Should not be able to insert any value 1-9 in a different position of this box now
		// Try to insert value 1 (which already exists at [0][0]) into position [1][0]
		_, err = s.Insert(1, 0, 1) // Try to insert 1 in a different position in same box
		if err == nil {
			t.Error("Expected error when trying to insert duplicate value in same box, got nil")
		}

		// Try to insert value 5 (which already exists at [1][1]) into position [0][1]
		_, err = s.Insert(0, 1, 5) // Try to insert 5 in a different position in same box
		if err == nil {
			t.Error("Expected error when trying to insert duplicate value in same box, got nil")
		}
	})

	t.Run("large invalid values", func(t *testing.T) {
		s := New()

		testValues := []int{100, 1000, -100, -1000}
		for _, val := range testValues {
			_, err := s.Insert(4, 4, val)
			if err == nil {
				t.Errorf("Expected error for invalid value %d, got nil", val)
			}
		}
	})

	t.Run("insert zero explicitly", func(t *testing.T) {
		s := New()
		s[5][5] = 3 // Set initial value

		_, err := s.Insert(5, 5, 0) // Insert zero (EmptyCell)
		if err != nil {
			t.Errorf("Expected no error when inserting 0, got: %v", err)
		}
		if s[5][5] != 0 {
			t.Errorf("Expected cell [5][5] to be 0, got %d", s[5][5])
		}
	})

	t.Run("multiple valid insertions in sequence", func(t *testing.T) {
		s := New()

		insertions := []struct {
			row, col, val int
		}{
			{0, 0, 1},
			{0, 1, 2},
			{0, 2, 3},
			{1, 0, 4},
			{1, 1, 5},
			{1, 2, 6},
			{2, 0, 7},
			{2, 1, 8},
			{2, 2, 9},
		}

		for _, ins := range insertions {
			_, err := s.Insert(ins.row, ins.col, ins.val)
			if err != nil {
				t.Errorf("Expected no error for insertion [%d][%d]=%d, got: %v", ins.row, ins.col, ins.val, err)
			}
			if s[ins.row][ins.col] != ins.val {
				t.Errorf("Expected cell [%d][%d] to be %d, got %d", ins.row, ins.col, ins.val, s[ins.row][ins.col])
			}
		}
	})
}

func TestCopy(t *testing.T) {
	t.Run("copy complete sudoku", func(t *testing.T) {
		s := &Sudoku{
			{5, 3, 4, 6, 7, 8, 9, 1, 2},
			{6, 7, 2, 1, 9, 5, 3, 4, 8},
			{1, 9, 8, 3, 4, 2, 5, 6, 7},
			{8, 5, 9, 7, 6, 1, 4, 2, 3},
			{4, 2, 6, 8, 5, 3, 7, 9, 1},
			{7, 1, 3, 9, 2, 4, 8, 5, 6},
			{9, 6, 1, 5, 3, 7, 2, 8, 4},
			{2, 8, 7, 4, 1, 9, 6, 3, 5},
			{3, 4, 5, 2, 8, 6, 1, 7, 9},
		}

		c := New()
		s.Copy(c)

		if !s.Is(c) {
			t.Error("Expected copied sudoku to have same values as original")
		}
	})

	t.Run("copy empty sudoku", func(t *testing.T) {
		s := New()
		c := New()
		s.Copy(c)

		if !s.Is(c) {
			t.Error("Expected copied sudoku to have same values as original")
		}
	})
}

func TestClues(t *testing.T) {
	t.Run("complete sudoku", func(t *testing.T) {
		s := &Sudoku{
			{5, 3, 4, 6, 7, 8, 9, 1, 2},
			{6, 7, 2, 1, 9, 5, 3, 4, 8},
			{1, 9, 8, 3, 4, 2, 5, 6, 7},
			{8, 5, 9, 7, 6, 1, 4, 2, 3},
			{4, 2, 6, 8, 5, 3, 7, 9, 1},
			{7, 1, 3, 9, 2, 4, 8, 5, 6},
			{9, 6, 1, 5, 3, 7, 2, 8, 4},
			{2, 8, 7, 4, 1, 9, 6, 3, 5},
			{3, 4, 5, 2, 8, 6, 1, 7, 9},
		}

		got := s.Clues()
		want := 81
		if got != want {
			t.Errorf("got: %d, want: %d", got, want)
		}
	})

	t.Run("empty sudoku", func(t *testing.T) {
		s := New()
		got := s.Clues()
		want := 0
		if got != want {
			t.Errorf("got: %d, want: %d", got, want)
		}
	})

	t.Run("17 clue minimum sudoku", func(t *testing.T) {
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

		got := s.Clues()
		want := 17
		if got != want {
			t.Errorf("got: %d, want: %d", got, want)
		}
	})

	t.Run("16 clue non unique solution sudoku", func(t *testing.T) {
		s := &Sudoku{
			{5, 0, 0, 0, 7, 0, 0, 0, 0},
			{6, 0, 0, 1, 0, 0, 0, 0, 0},
			{0, 9, 0, 0, 0, 0, 0, 6, 0},
			{8, 0, 0, 0, 6, 0, 0, 0, 3},
			{0, 0, 0, 8, 0, 0, 0, 0, 0},
			{7, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 6, 0, 0, 0, 0, 0, 8, 0},
			{0, 0, 0, 4, 0, 0, 0, 0, 5},
			{0, 0, 0, 0, 8, 0, 0, 0, 0},
		}

		got := s.Clues()
		want := 16
		if got != want {
			t.Errorf("got: %d, want: %d", got, want)
		}
	})
}

func TestUniqueSolution(t *testing.T) {
	t.Run("17-clue solvable sudoku has unique solution", func(t *testing.T) {
		// Known 17-clue sudoku with unique solution
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

		result := s.UniqueSolution()
		if !result {
			t.Error("Expected 17-clue sudoku to have unique solution (should return true)")
		}

		// Verify the sudoku was solved completely
		if !s.Complete() {
			t.Error("Expected sudoku to be complete after UniqueSolution returns true")
		}
	})

	t.Run("16-clue sudoku hasn't a unique solution", func(t *testing.T) {
		s := &Sudoku{
			{5, 0, 0, 0, 7, 0, 0, 0, 0},
			{6, 0, 0, 1, 0, 0, 0, 0, 0},
			{0, 9, 0, 0, 0, 0, 0, 6, 0},
			{8, 0, 0, 0, 6, 0, 0, 0, 3},
			{0, 0, 0, 8, 0, 0, 0, 0, 0},
			{7, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 6, 0, 0, 0, 0, 0, 8, 0},
			{0, 0, 0, 4, 0, 0, 0, 0, 5},
			{0, 0, 0, 0, 8, 0, 0, 0, 0},
		}

		result := s.UniqueSolution()
		if result {
			t.Error("Expected 16-clue sudoku not to have unique solution (should return false)")
		}
	})

	t.Run("already complete sudoku has unique solution", func(t *testing.T) {
		s := &Sudoku{
			{5, 3, 4, 6, 7, 8, 9, 1, 2},
			{6, 7, 2, 1, 9, 5, 3, 4, 8},
			{1, 9, 8, 3, 4, 2, 5, 6, 7},
			{8, 5, 9, 7, 6, 1, 4, 2, 3},
			{4, 2, 6, 8, 5, 3, 7, 9, 1},
			{7, 1, 3, 9, 2, 4, 8, 5, 6},
			{9, 6, 1, 5, 3, 7, 2, 8, 4},
			{2, 8, 7, 4, 1, 9, 6, 3, 5},
			{3, 4, 5, 2, 8, 6, 1, 7, 9},
		}

		result := s.UniqueSolution()
		if !result {
			t.Error("Expected complete sudoku to have unique solution (should return true)")
		}

		// Should still be complete
		if !s.Complete() {
			t.Error("Expected complete sudoku to remain complete")
		}
	})
}
