package terminus

import (
	"testing"
)

func TestScreen(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "NewScreen creates correct dimensions",
			test: func(t *testing.T) {
				screen := NewScreen(80, 24)
				
				if screen.width != 80 {
					t.Errorf("Expected width 80, got %d", screen.width)
				}
				
				if screen.height != 24 {
					t.Errorf("Expected height 24, got %d", screen.height)
				}
				
				if len(screen.lines) != 24 {
					t.Errorf("Expected 24 lines, got %d", len(screen.lines))
				}
				
				for i, line := range screen.lines {
					if len(line) != 80 {
						t.Errorf("Line %d: expected 80 cells, got %d", i, len(line))
					}
				}
			},
		},
		{
			name: "SetCell and GetCell",
			test: func(t *testing.T) {
				screen := NewScreen(10, 10)
				style := NewStyle().Bold(true)
				
				screen.SetCell(5, 5, 'A', style)
				
				cell := screen.GetCell(5, 5)
				if cell.Rune != 'A' {
					t.Errorf("Expected rune 'A', got '%c'", cell.Rune)
				}
				
				// Test out of bounds
				screen.SetCell(-1, -1, 'B', style)
				screen.SetCell(10, 10, 'C', style)
				
				// Should return space for out of bounds
				cell = screen.GetCell(-1, -1)
				if cell.Rune != ' ' {
					t.Errorf("Expected space for out of bounds, got '%c'", cell.Rune)
				}
			},
		},
		{
			name: "Clear screen",
			test: func(t *testing.T) {
				screen := NewScreen(10, 10)
				
				// Fill with 'X'
				for y := 0; y < 10; y++ {
					for x := 0; x < 10; x++ {
						screen.SetCell(x, y, 'X', NewStyle())
					}
				}
				
				// Clear
				screen.Clear()
				
				// Check all cells are spaces
				for y := 0; y < 10; y++ {
					for x := 0; x < 10; x++ {
						cell := screen.GetCell(x, y)
						if cell.Rune != ' ' {
							t.Errorf("Cell at (%d,%d) should be space, got '%c'", x, y, cell.Rune)
						}
					}
				}
				
				// Check cursor reset
				if screen.cursor.x != 0 || screen.cursor.y != 0 {
					t.Error("Cursor should be reset to (0,0)")
				}
			},
		},
		{
			name: "RenderFromString simple text",
			test: func(t *testing.T) {
				screen := NewScreen(10, 5)
				screen.RenderFromString("Hello")
				
				// Check rendered text
				expected := "Hello"
				for i, r := range expected {
					cell := screen.GetCell(i, 0)
					if cell.Rune != r {
						t.Errorf("Position %d: expected '%c', got '%c'", i, r, cell.Rune)
					}
				}
			},
		},
		{
			name: "RenderFromString with newline",
			test: func(t *testing.T) {
				screen := NewScreen(10, 5)
				screen.RenderFromString("Hello\nWorld")
				
				// Check first line
				for i, r := range "Hello" {
					cell := screen.GetCell(i, 0)
					if cell.Rune != r {
						t.Errorf("Line 0, position %d: expected '%c', got '%c'", i, r, cell.Rune)
					}
				}
				
				// Check second line
				for i, r := range "World" {
					cell := screen.GetCell(i, 1)
					if cell.Rune != r {
						t.Errorf("Line 1, position %d: expected '%c', got '%c'", i, r, cell.Rune)
					}
				}
			},
		},
		{
			name: "RenderFromString with word wrap",
			test: func(t *testing.T) {
				screen := NewScreen(5, 3)
				screen.RenderFromString("HelloWorld")
				
				// First line should have "Hello"
				for i, r := range "Hello" {
					cell := screen.GetCell(i, 0)
					if cell.Rune != r {
						t.Errorf("Line 0, position %d: expected '%c', got '%c'", i, r, cell.Rune)
					}
				}
				
				// Second line should have "World"
				for i, r := range "World" {
					cell := screen.GetCell(i, 1)
					if cell.Rune != r {
						t.Errorf("Line 1, position %d: expected '%c', got '%c'", i, r, cell.Rune)
					}
				}
			},
		},
		{
			name: "RenderFromString with scrolling",
			test: func(t *testing.T) {
				screen := NewScreen(10, 3)
				screen.RenderFromString("Line1\nLine2\nLine3\nLine4")
				
				// Should only see last 3 lines
				lines := []string{"Line2", "Line3", "Line4"}
				for y, line := range lines {
					for x, r := range line {
						cell := screen.GetCell(x, y)
						if cell.Rune != r {
							t.Errorf("Line %d, position %d: expected '%c', got '%c'", y, x, r, cell.Rune)
						}
					}
				}
			},
		},
		{
			name: "ToString",
			test: func(t *testing.T) {
				screen := NewScreen(5, 3)
				screen.RenderFromString("AB\nCD\nEF")
				
				result := screen.ToString()
				
				// Should be padded with spaces
				expected := "AB   \nCD   \nEF   "
				if result != expected {
					t.Errorf("Expected:\n%q\nGot:\n%q", expected, result)
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

func TestANSIParser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []struct {
			r     rune
			style string
		}
	}{
		{
			name:  "Plain text",
			input: "Hello",
			expected: []struct {
				r     rune
				style string
			}{
				{r: 'H', style: "Style{}"},
				{r: 'e', style: "Style{}"},
				{r: 'l', style: "Style{}"},
				{r: 'l', style: "Style{}"},
				{r: 'o', style: "Style{}"},
			},
		},
		{
			name:  "Bold text",
			input: "\x1b[1mBold\x1b[0m",
			expected: []struct {
				r     rune
				style string
			}{
				{r: 'B', style: "Style{bold}"},
				{r: 'o', style: "Style{bold}"},
				{r: 'l', style: "Style{bold}"},
				{r: 'd', style: "Style{bold}"},
			},
		},
		{
			name:  "Colored text",
			input: "\x1b[31mRed\x1b[0m",
			expected: []struct {
				r     rune
				style string
			}{
				{r: 'R', style: "Style{fg:red}"},
				{r: 'e', style: "Style{fg:red}"},
				{r: 'd', style: "Style{fg:red}"},
			},
		},
		{
			name:  "Multiple attributes",
			input: "\x1b[1;31;44mText\x1b[0m",
			expected: []struct {
				r     rune
				style string
			}{
				{r: 'T', style: "Style{bold, fg:red, bg:blue}"},
				{r: 'e', style: "Style{bold, fg:red, bg:blue}"},
				{r: 'x', style: "Style{bold, fg:red, bg:blue}"},
				{r: 't', style: "Style{bold, fg:red, bg:blue}"},
			},
		},
		{
			name:  "Reset in middle",
			input: "\x1b[1mBold\x1b[0mNormal",
			expected: []struct {
				r     rune
				style string
			}{
				{r: 'B', style: "Style{bold}"},
				{r: 'o', style: "Style{bold}"},
				{r: 'l', style: "Style{bold}"},
				{r: 'd', style: "Style{bold}"},
				{r: 'N', style: "Style{}"},
				{r: 'o', style: "Style{}"},
				{r: 'r', style: "Style{}"},
				{r: 'm', style: "Style{}"},
				{r: 'a', style: "Style{}"},
				{r: 'l', style: "Style{}"},
			},
		},
		{
			name:  "UTF-8 characters",
			input: "Hello 世界",
			expected: []struct {
				r     rune
				style string
			}{
				{r: 'H', style: "Style{}"},
				{r: 'e', style: "Style{}"},
				{r: 'l', style: "Style{}"},
				{r: 'l', style: "Style{}"},
				{r: 'o', style: "Style{}"},
				{r: ' ', style: "Style{}"},
				{r: '世', style: "Style{}"},
				{r: '界', style: "Style{}"},
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewANSIParser(tt.input)
			
			for i, expected := range tt.expected {
				r, style, ok := parser.Next()
				if !ok {
					t.Fatalf("Parser ended early at position %d", i)
				}
				
				if r != expected.r {
					t.Errorf("Position %d: expected rune '%c', got '%c'", i, expected.r, r)
				}
				
				styleStr := style.String()
				if styleStr != expected.style {
					t.Errorf("Position %d: expected style %s, got %s", i, expected.style, styleStr)
				}
			}
			
			// Check no more characters
			_, _, ok := parser.Next()
			if ok {
				t.Error("Parser should have ended")
			}
		})
	}
}

func TestScreenWithANSI(t *testing.T) {
	screen := NewScreen(20, 5)
	
	// Render styled text
	input := "\x1b[1;31mBold Red\x1b[0m\n\x1b[32mGreen\x1b[0m"
	screen.RenderFromString(input)
	
	// Check first line has bold red style
	for i, r := range "Bold Red" {
		cell := screen.GetCell(i, 0)
		if cell.Rune != r {
			t.Errorf("Position %d: expected '%c', got '%c'", i, r, cell.Rune)
		}
		// We can't easily check the style without exposing it,
		// but the parser tests verify styles are parsed correctly
	}
	
	// Check second line
	for i, r := range "Green" {
		cell := screen.GetCell(i, 1)
		if cell.Rune != r {
			t.Errorf("Line 1, position %d: expected '%c', got '%c'", i, r, cell.Rune)
		}
	}
}

func TestScrollUp(t *testing.T) {
	screen := NewScreen(10, 3)
	
	// Fill screen with distinct content
	for y := 0; y < 3; y++ {
		for x := 0; x < 10; x++ {
			screen.SetCell(x, y, rune('A'+y), NewStyle())
		}
	}
	
	// Scroll up
	screen.scrollUp()
	
	// First line should now have content from second line (B)
	for x := 0; x < 10; x++ {
		cell := screen.GetCell(x, 0)
		if cell.Rune != 'B' {
			t.Errorf("Line 0, position %d: expected 'B', got '%c'", x, cell.Rune)
		}
	}
	
	// Second line should have content from third line (C)
	for x := 0; x < 10; x++ {
		cell := screen.GetCell(x, 1)
		if cell.Rune != 'C' {
			t.Errorf("Line 1, position %d: expected 'C', got '%c'", x, cell.Rune)
		}
	}
	
	// Last line should be empty
	for x := 0; x < 10; x++ {
		cell := screen.GetCell(x, 2)
		if cell.Rune != ' ' {
			t.Errorf("Line 2, position %d: expected space, got '%c'", x, cell.Rune)
		}
	}
}

func TestTabHandling(t *testing.T) {
	screen := NewScreen(20, 2)
	screen.RenderFromString("A\tB\tC")
	
	// Tabs should move to next 8-character boundary
	// A at 0, tab moves to 8, B at 8, tab moves to 16, C at 16
	
	if screen.GetCell(0, 0).Rune != 'A' {
		t.Error("Expected 'A' at position 0")
	}
	
	if screen.GetCell(8, 0).Rune != 'B' {
		t.Error("Expected 'B' at position 8")
	}
	
	if screen.GetCell(16, 0).Rune != 'C' {
		t.Error("Expected 'C' at position 16")
	}
}