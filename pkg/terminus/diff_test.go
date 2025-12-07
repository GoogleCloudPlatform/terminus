package terminus

import (
	"testing"
)

func TestDiffer(t *testing.T) {
	tests := []struct {
		name      string
		oldScreen *Screen
		newScreen *Screen
		expected  string // expected ANSI output
	}{
		{
			name:      "First render (nil old screen)",
			oldScreen: nil,
			newScreen: func() *Screen {
				s := NewScreen(10, 3)
				s.RenderFromString("Hello\nWorld")
				return s
			}(),
			expected: "\x1b[2J\x1b[H\x1b[1;1HHello\x1b[2;1HWorld\x1b[0m",
		},
		{
			name: "No changes",
			oldScreen: func() *Screen {
				s := NewScreen(10, 3)
				s.RenderFromString("Hello")
				return s
			}(),
			newScreen: func() *Screen {
				s := NewScreen(10, 3)
				s.RenderFromString("Hello")
				return s
			}(),
			expected: "\x1b[0m",
		},
		{
			name: "Single line change",
			oldScreen: func() *Screen {
				s := NewScreen(10, 3)
				s.RenderFromString("Hello\nWorld\nTest")
				return s
			}(),
			newScreen: func() *Screen {
				s := NewScreen(10, 3)
				s.RenderFromString("Hello\nChanged\nTest")
				return s
			}(),
			expected: "\x1b[2;1H\x1b[0KChanged\x1b[0m",
		},
		{
			name: "Multiple line changes",
			oldScreen: func() *Screen {
				s := NewScreen(10, 3)
				s.RenderFromString("AAA\nBBB\nCCC")
				return s
			}(),
			newScreen: func() *Screen {
				s := NewScreen(10, 3)
				s.RenderFromString("XXX\nBBB\nZZZ")
				return s
			}(),
			expected: "\x1b[1;1H\x1b[0KXXX\x1b[3;1H\x1b[0KZZZ\x1b[0m",
		},
		{
			name: "Dimension change forces full redraw",
			oldScreen: func() *Screen {
				s := NewScreen(10, 3)
				s.RenderFromString("Hello")
				return s
			}(),
			newScreen: func() *Screen {
				s := NewScreen(20, 5)
				s.RenderFromString("Hello")
				return s
			}(),
			expected: "\x1b[2J\x1b[H\x1b[1;1HHello\x1b[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			differ := NewDiffer()
			result := differ.Diff(tt.oldScreen, tt.newScreen)

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestScreenDiffer(t *testing.T) {

	tests := []struct {

		name           string

		initialContent string

		updates        []struct {

			content string

			expected string

		}

		resizeWidth    int

		resizeHeight   int

		expectedResize string

		reset          bool

		expectedReset  string

	}{

		{

			name: "Initial update and no changes",

			initialContent: "Hello\nWorld",

			updates: []struct{

				content string

				expected string

			}{

				{content: "Hello\nWorld", expected: "\x1b[2J\x1b[H\x1b[1;1HHello\x1b[2;1HWorld\x1b[0m"},

				{content: "Hello\nWorld", expected: "\x1b[0m"},

			},

		},

		{

			name: "Single line change",

			initialContent: "Line1\nLine2\nLine3",

			updates: []struct{

				content string

				expected string

			}{

				{content: "Line1\nLine2\nLine3", expected: "\x1b[2J\x1b[H\x1b[1;1HLine1\x1b[2;1HLine2\x1b[3;1HLine3\x1b[0m"},

				{content: "Line1\nChanged\nLine3", expected: "\x1b[2;1H\x1b[0KChanged\x1b[0m"},

			},

		},

		{

			name: "Resize forces redraw",

			initialContent: "Hello",

			updates: []struct{

				content string

				expected string

			}{

				{content: "Hello", expected: "\x1b[2J\x1b[H\x1b[1;1HHello\x1b[0m"},

			},

			resizeWidth: 30,

		resizeHeight: 10,

		expectedResize: "\x1b[2J\x1b[H\x1b[1;1HHello\x1b[0m",

		},

		{

			name: "Reset clears state",

			initialContent: "Hello",

			updates: []struct{

				content string

				expected string

			}{

				{content: "Hello", expected: "\x1b[2J\x1b[H\x1b[1;1HHello\x1b[0m"},

			},

			reset: true,

			expectedReset: "\x1b[2J\x1b[H\x1b[1;1HHello\x1b[0m",

		},

	}



	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			sd := NewScreenDiffer(20, 5)



			for i, update := range tt.updates {

				result := sd.Update(update.content)

				if result != update.expected {

					t.Errorf("Update %d: Expected %q, got %q", i, update.expected, result)

				}

			}



			if tt.resizeWidth > 0 || tt.resizeHeight > 0 {

				sd.Resize(tt.resizeWidth, tt.resizeHeight)

				result := sd.Update(tt.initialContent) // Re-render after resize

				if result != tt.expectedResize {

					t.Errorf("Resize: Expected %q, got %q", tt.expectedResize, result)

				}

			}



			if tt.reset {

				sd.Reset()

				result := sd.Update(tt.initialContent) // Re-render after reset

				if result != tt.expectedReset {

					t.Errorf("Reset: Expected %q, got %q", tt.expectedReset, result)

				}

			}

		})

	}

}



func TestStyleTransitions(t *testing.T) {
	// Setup screen directly with styles for renderLine test
	screen := NewScreen(20, 1)
	bold := NewStyle().Bold(true)
	red := NewStyle().Foreground(Red)

	screen.SetCell(0, 0, 'A', NewStyle())
	screen.SetCell(1, 0, 'B', bold)
	screen.SetCell(2, 0, 'C', bold)
	screen.SetCell(3, 0, 'D', red)
	screen.SetCell(4, 0, 'E', NewStyle())

	differ := &Differ{newScreen: screen}
	line := differ.renderLine(screen, 0)

	expected := "A\x1b[0m\x1b[1mBC\x1b[0m\x1b[31mD\x1b[0mE"
	if line != expected {
		t.Errorf("renderLine Expected %q, got %q", expected, line)
	}

	// Test with ScreenDiffer and full ANSI string input
	sd := NewScreenDiffer(20, 1)
	
	// Initial content with full ANSI styles for correct parsing
	initialANSIContent := "A\x1b[0m\x1b[1mBC\x1b[0m\x1b[31mD\x1b[0mE"
	
	// First update should be a full redraw
	fullRenderedANSI := sd.Update(initialANSIContent)

	// Expected: clear screen, move cursor to 1;1, then the styled line, then final reset.
	expectedFull := "\x1b[2J\x1b[H\x1b[1;1H" + initialANSIContent + "\x1b[0m"

	if fullRenderedANSI != expectedFull {
		t.Errorf("ScreenDiffer.Update Expected full render %q, got %q", expectedFull, fullRenderedANSI)
	}

	// Now update with no changes, should produce only a reset
	noChangeANSI := sd.Update(initialANSIContent)
	if noChangeANSI != "\x1b[0m" {
		t.Errorf("ScreenDiffer.Update Expected no change output %q, got %q", "\x1b[0m", noChangeANSI)
	}

	// Update with a single cell change
	changedANSIContent := "A\x1b[0m\x1b[1mBC\x1b[0m\x1b[32mD\x1b[0mE" // Changed D to green
	singleChangeANSI := sd.Update(changedANSIContent)
	expectedSingleChange := "\x1b[1;1H\x1b[0K" + changedANSIContent + "\x1b[0m"
	if singleChangeANSI != expectedSingleChange {
		t.Errorf("ScreenDiffer.Update Expected single change %q, got %q", expectedSingleChange, singleChangeANSI)
	}
}