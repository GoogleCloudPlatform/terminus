// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package terminus

import (
	"strings"
	"unicode/utf8"
)

// Cell represents a single character cell in the terminal
type Cell struct {
	Rune  rune
	Style Style
}

// Line represents a line of cells
type Line []Cell

// Screen represents the virtual terminal screen
type Screen struct {
	width  int
	height int
	lines  []Line
	cursor struct {
		x int
		y int
	}
}

// NewScreen creates a new virtual screen
func NewScreen(width, height int) *Screen {
	s := &Screen{
		width:  width,
		height: height,
		lines:  make([]Line, height),
	}
	
	// Initialize empty lines
	for i := range s.lines {
		s.lines[i] = make(Line, width)
		for j := range s.lines[i] {
			s.lines[i][j] = Cell{Rune: ' '}
		}
	}
	
	return s
}

// Clear clears the screen
func (s *Screen) Clear() {
	for i := range s.lines {
		for j := range s.lines[i] {
			s.lines[i][j] = Cell{Rune: ' '}
		}
	}
	s.cursor.x = 0
	s.cursor.y = 0
}

// SetCell sets a cell at the given position
func (s *Screen) SetCell(x, y int, r rune, style Style) {
	if x >= 0 && x < s.width && y >= 0 && y < s.height {
		s.lines[y][x] = Cell{Rune: r, Style: style}
	}
}

// GetCell gets the cell at the given position
func (s *Screen) GetCell(x, y int) Cell {
	if x >= 0 && x < s.width && y >= 0 && y < s.height {
		return s.lines[y][x]
	}
	return Cell{Rune: ' '}
}

// RenderFromString renders a string to the screen, handling ANSI codes
func (s *Screen) RenderFromString(content string) {
	s.Clear()
	
	// Parse the string and render to screen
	parser := NewANSIParser(content)
	s.cursor.x = 0
	s.cursor.y = 0
	
	for {
		r, style, ok := parser.Next()
		if !ok {
			break
		}
		
		// Handle special characters
		switch r {
		case '\n':
			s.cursor.x = 0
			s.cursor.y++
			if s.cursor.y >= s.height {
				// Scroll up
				s.scrollUp()
				s.cursor.y = s.height - 1
			}
		case '\r':
			s.cursor.x = 0
		case '\t':
			// Move to next tab stop (every 8 characters)
			nextTab := ((s.cursor.x / 8) + 1) * 8
			if nextTab < s.width {
				s.cursor.x = nextTab
			}
		default:
			// Regular character
			if s.cursor.x < s.width && s.cursor.y < s.height {
				s.SetCell(s.cursor.x, s.cursor.y, r, style)
				s.cursor.x++
				
				// Wrap to next line
				if s.cursor.x >= s.width {
					s.cursor.x = 0
					s.cursor.y++
					if s.cursor.y >= s.height {
						// Scroll up
						s.scrollUp()
						s.cursor.y = s.height - 1
					}
				}
			}
		}
	}
}

// scrollUp scrolls the screen up by one line
func (s *Screen) scrollUp() {
	// Move all lines up
	copy(s.lines, s.lines[1:])
	
	// Clear the last line
	s.lines[s.height-1] = make(Line, s.width)
	for j := range s.lines[s.height-1] {
		s.lines[s.height-1][j] = Cell{Rune: ' '}
	}
}

// ToString converts the screen to a plain string (for testing)
func (s *Screen) ToString() string {
	var builder strings.Builder
	
	for y, line := range s.lines {
		for _, cell := range line {
			builder.WriteRune(cell.Rune)
		}
		if y < s.height-1 {
			builder.WriteRune('\n')
		}
	}
	
	return builder.String()
}

// ANSIParser parses ANSI escape sequences from a string
type ANSIParser struct {
	input   string
	pos     int
	current Style
}

// NewANSIParser creates a new ANSI parser
func NewANSIParser(input string) *ANSIParser {
	return &ANSIParser{
		input:   input,
		current: NewStyle(),
	}
}

// Next returns the next rune and its style
func (p *ANSIParser) Next() (rune, Style, bool) {
	if p.pos >= len(p.input) {
		return 0, Style{}, false
	}
	
	// Check for ANSI escape sequence
	if p.pos+1 < len(p.input) && p.input[p.pos] == '\x1b' && p.input[p.pos+1] == '[' {
		// Parse ANSI sequence
		p.pos += 2 // Skip ESC[
		
		// Find the end of the sequence
		start := p.pos
		for p.pos < len(p.input) {
			c := p.input[p.pos]
			if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
				// Found terminator
				p.pos++
				break
			}
			p.pos++
		}
		
		// Parse the sequence
		if p.pos > start {
			codes := p.input[start : p.pos-1]
			terminator := p.input[p.pos-1]
			
			if terminator == 'm' {
				// SGR (Select Graphic Rendition) sequence
				p.parseSGR(codes)
			}
		}
		
		// Continue to next character
		return p.Next()
	}
	
	// Regular character
	r, size := utf8.DecodeRuneInString(p.input[p.pos:])
	p.pos += size
	
	return r, p.current, true
}

// parseSGR parses SGR (Select Graphic Rendition) codes
func (p *ANSIParser) parseSGR(codes string) {
	if codes == "" || codes == "0" {
		// Reset all attributes
		p.current = NewStyle()
		return
	}
	
	// Split codes by semicolon
	parts := strings.Split(codes, ";")
	
	for i := 0; i < len(parts); i++ {
		code := parts[i]
		
		switch code {
		case "0":
			// Reset
			p.current = NewStyle()
		case "1":
			// Bold
			p.current = p.current.Bold(true)
		case "2":
			// Faint
			p.current = p.current.Faint(true)
		case "3":
			// Italic
			p.current = p.current.Italic(true)
		case "4":
			// Underline
			p.current = p.current.Underline(true)
		case "5":
			// Blink
			p.current = p.current.Blink(true)
		case "7":
			// Reverse
			p.current = p.current.Reverse(true)
		case "9":
			// Crossed out
			p.current = p.current.CrossOut(true)
			
		// Foreground colors
		case "30":
			p.current = p.current.Foreground(Black)
		case "31":
			p.current = p.current.Foreground(Red)
		case "32":
			p.current = p.current.Foreground(Green)
		case "33":
			p.current = p.current.Foreground(Yellow)
		case "34":
			p.current = p.current.Foreground(Blue)
		case "35":
			p.current = p.current.Foreground(Magenta)
		case "36":
			p.current = p.current.Foreground(Cyan)
		case "37":
			p.current = p.current.Foreground(White)
			
		// Background colors
		case "40":
			p.current = p.current.Background(Black)
		case "41":
			p.current = p.current.Background(Red)
		case "42":
			p.current = p.current.Background(Green)
		case "43":
			p.current = p.current.Background(Yellow)
		case "44":
			p.current = p.current.Background(Blue)
		case "45":
			p.current = p.current.Background(Magenta)
		case "46":
			p.current = p.current.Background(Cyan)
		case "47":
			p.current = p.current.Background(White)
			
		// Bright foreground colors
		case "90":
			p.current = p.current.Foreground(BrightBlack)
		case "91":
			p.current = p.current.Foreground(BrightRed)
		case "92":
			p.current = p.current.Foreground(BrightGreen)
		case "93":
			p.current = p.current.Foreground(BrightYellow)
		case "94":
			p.current = p.current.Foreground(BrightBlue)
		case "95":
			p.current = p.current.Foreground(BrightMagenta)
		case "96":
			p.current = p.current.Foreground(BrightCyan)
		case "97":
			p.current = p.current.Foreground(BrightWhite)
			
		// Bright background colors
		case "100":
			p.current = p.current.Background(BrightBlack)
		case "101":
			p.current = p.current.Background(BrightRed)
		case "102":
			p.current = p.current.Background(BrightGreen)
		case "103":
			p.current = p.current.Background(BrightYellow)
		case "104":
			p.current = p.current.Background(BrightBlue)
		case "105":
			p.current = p.current.Background(BrightMagenta)
		case "106":
			p.current = p.current.Background(BrightCyan)
		case "107":
			p.current = p.current.Background(BrightWhite)
			
		// 256 color and RGB not implemented yet for simplicity
		}
	}
}