package layout

import (
	"strings"
	"testing"
)

func TestColumns(t *testing.T) {
	tests := []struct {
		name     string
		contents []string
		widths   []int
		gap      int
		expected string
	}{
		{
			name:     "Simple two columns",
			contents: []string{"Left", "Right"},
			widths:   []int{10, 10},
			gap:      2,
			expected: "Left        Right     ",
		},
		{
			name:     "Multi-line columns",
			contents: []string{"Line 1\nLine 2", "Item A\nItem B"},
			widths:   []int{10, 10},
			gap:      1,
			expected: "Line 1     Item A    \nLine 2     Item B    ",
		},
		{
			name:     "Different column widths",
			contents: []string{"Short", "This is longer"},
			widths:   []int{5, 15},
			gap:      3,
			expected: "Short   This is longer ",
		},
		{
			name:     "More content than widths",
			contents: []string{"A", "B", "C"},
			widths:   []int{5, 5},
			gap:      1,
			expected: "A     B     C    ",
		},
		{
			name:     "Content truncation",
			contents: []string{"This is too long", "OK"},
			widths:   []int{5, 5},
			gap:      1,
			expected: "This  OK   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Columns(tt.contents, tt.widths, tt.gap)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestRows(t *testing.T) {
	tests := []struct {
		name     string
		contents []string
		gap      int
		expected string
	}{
		{
			name:     "Simple rows",
			contents: []string{"Row 1", "Row 2", "Row 3"},
			gap:      0,
			expected: "Row 1\nRow 2\nRow 3",
		},
		{
			name:     "Rows with gap",
			contents: []string{"Row 1", "Row 2"},
			gap:      1,
			expected: "Row 1\n\nRow 2",
		},
		{
			name:     "Single row",
			contents: []string{"Only row"},
			gap:      2,
			expected: "Only row",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Rows(tt.contents, tt.gap)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestCenter(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		width    int
		height   int
		checkFor []string
	}{
		{
			name:    "Center single line",
			content: "Hi",
			width:   6,
			height:  3,
			checkFor: []string{
				"      ", // top padding
				"  Hi  ", // centered content
				"      ", // bottom padding
			},
		},
		{
			name:    "Center multi-line",
			content: "Line 1\nLine 2",
			width:   10,
			height:  5,
			checkFor: []string{
				"          ", // top padding
				"  Line 1  ",
				"  Line 2  ",
				"          ", // bottom padding
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Center(tt.content, tt.width, tt.height)
			lines := strings.Split(result, "\n")

			if len(lines) != tt.height {
				t.Errorf("Expected %d lines, got %d", tt.height, len(lines))
			}

			for i, expected := range tt.checkFor {
				if i < len(lines) && lines[i] != expected {
					t.Errorf("Line %d: expected %q, got %q", i, expected, lines[i])
				}
			}
		})
	}
}

func TestAlign(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		width      int
		height     int
		horizontal Alignment
		vertical   Alignment
		checkLine  int
		expected   string
	}{
		{
			name:       "Top left",
			content:    "TL",
			width:      5,
			height:     3,
			horizontal: AlignLeft,
			vertical:   AlignTop,
			checkLine:  0,
			expected:   "TL   ",
		},
		{
			name:       "Bottom right",
			content:    "BR",
			width:      5,
			height:     3,
			horizontal: AlignRight,
			vertical:   AlignBottom,
			checkLine:  2,
			expected:   "   BR",
		},
		{
			name:       "Center middle",
			content:    "X",
			width:      5,
			height:     3,
			horizontal: AlignCenter,
			vertical:   AlignMiddle,
			checkLine:  1,
			expected:   "  X  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Align(tt.content, tt.width, tt.height, tt.horizontal, tt.vertical)
			lines := strings.Split(result, "\n")

			if len(lines) != tt.height {
				t.Errorf("Expected %d lines, got %d", tt.height, len(lines))
			}

			if tt.checkLine < len(lines) && lines[tt.checkLine] != tt.expected {
				t.Errorf("Line %d: expected %q, got %q", tt.checkLine, tt.expected, lines[tt.checkLine])
			}
		})
	}
}

func TestMargin(t *testing.T) {
	content := "Text"
	result := Margin(content, 1, 2, 1, 3)

	lines := strings.Split(result, "\n")
	if len(lines) != 3 { // 1 top + 1 content + 1 bottom
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}

	// Check content line has correct margins
	contentLine := lines[1]
	if !strings.HasPrefix(contentLine, "   ") { // 3 spaces left margin
		t.Error("Missing left margin")
	}
	if !strings.Contains(contentLine, "Text") {
		t.Error("Missing content")
	}
}

func TestAddPadding(t *testing.T) {
	// Since AddPadding is just an alias for Margin, basic test
	content := "Test"
	marginResult := Margin(content, 1, 1, 1, 1)
	paddingResult := AddPadding(content, 1, 1, 1, 1)

	if marginResult != paddingResult {
		t.Error("AddPadding should produce same result as Margin")
	}
}

func TestGrid(t *testing.T) {
	grid := NewGrid(3, 2).SetGap(0)
	grid.SetCell(0, 0, "A")
	grid.SetCell(1, 0, "B")
	grid.SetCell(2, 0, "C")
	grid.SetCell(0, 1, "1")
	grid.SetCell(1, 1, "2")
	grid.SetCell(2, 1, "3")

	result := grid.Render()
	// Trim trailing newline if present
	result = strings.TrimRight(result, "\n")
	lines := strings.Split(result, "\n")

	if len(lines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(lines))
		t.Errorf("Result: %q", result)
	}

	// Check that all cells are present
	if !strings.Contains(result, "A") || !strings.Contains(result, "B") || !strings.Contains(result, "C") {
		t.Error("Missing cells from first row")
	}
	if !strings.Contains(result, "1") || !strings.Contains(result, "2") || !strings.Contains(result, "3") {
		t.Error("Missing cells from second row")
	}
}

func TestGridWithGap(t *testing.T) {
	grid := NewGrid(2, 2).SetGap(2)
	grid.SetCell(0, 0, "TopLeft")
	grid.SetCell(1, 0, "TopRight")
	grid.SetCell(0, 1, "BottomLeft")
	grid.SetCell(1, 1, "BottomRight")

	result := grid.Render()
	lines := strings.Split(result, "\n")

	// With gap=2, we expect 3 lines (2 content + 1 gap line)
	if len(lines) < 3 {
		t.Errorf("Expected at least 3 lines with gap, got %d", len(lines))
	}

	// Check horizontal gap
	firstLine := lines[0]
	if !strings.Contains(firstLine, "TopLeft") || !strings.Contains(firstLine, "TopRight") {
		t.Error("Missing content from first row")
	}

	// Count spaces between cells
	leftEnd := strings.Index(firstLine, "t") + 1 // end of "TopLeft"
	rightStart := strings.Index(firstLine, "T")  // start of "TopRight"
	if rightStart > leftEnd {
		gap := rightStart - leftEnd
		if gap < 2 {
			t.Errorf("Expected gap of at least 2, got %d", gap)
		}
	}
}

func TestGridFixedDimensions(t *testing.T) {
	grid := NewGrid(2, 1)
	grid.SetColumnWidth(0, 10)
	grid.SetColumnWidth(1, 5)
	grid.SetCell(0, 0, "Long")
	grid.SetCell(1, 0, "Short")

	result := grid.Render()

	// Check that columns have correct widths
	if !strings.Contains(result, "Long      ") { // "Long" + 6 spaces = 10 chars
		t.Error("First column not padded to 10 characters")
	}
}

func TestGridMultilineCell(t *testing.T) {
	grid := NewGrid(2, 1)
	grid.SetCell(0, 0, "Line1\nLine2")
	grid.SetCell(1, 0, "Single")

	result := grid.Render()
	lines := strings.Split(result, "\n")

	// Should have 2 lines due to multiline cell
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines for multiline cell, got %d", len(lines))
	}

	if !strings.Contains(lines[0], "Line1") || !strings.Contains(lines[0], "Single") {
		t.Error("First line missing expected content")
	}
	if !strings.Contains(lines[1], "Line2") {
		t.Error("Second line missing expected content")
	}
}

func TestPadOrTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		align    Alignment
		expected string
	}{
		{
			name:     "Pad left align",
			input:    "Hi",
			width:    5,
			align:    AlignLeft,
			expected: "Hi   ",
		},
		{
			name:     "Pad right align",
			input:    "Hi",
			width:    5,
			align:    AlignRight,
			expected: "   Hi",
		},
		{
			name:     "Pad center align",
			input:    "Hi",
			width:    5,
			align:    AlignCenter,
			expected: " Hi  ",
		},
		{
			name:     "Truncate",
			input:    "TooLong",
			width:    3,
			align:    AlignLeft,
			expected: "Too",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := padOrTruncate(tt.input, tt.width, tt.align)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
