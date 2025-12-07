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
	"fmt"
	"strings"
)

// Differ computes differences between two screens
type Differ struct {
	oldScreen *Screen
	newScreen *Screen
}

// NewDiffer creates a new differ
func NewDiffer() *Differ {
	return &Differ{}
}

// Diff computes the differences between two screens and returns an ANSI string
func (d *Differ) Diff(oldScreen, newScreen *Screen) string {
	var sb strings.Builder
	d.oldScreen = oldScreen
	d.newScreen = newScreen

	// If dimensions changed, clear and redraw
	if oldScreen == nil || oldScreen.width != newScreen.width || oldScreen.height != newScreen.height {
		return d.fullRedraw()
	}

	// Compute line-by-line differences
	sb.WriteString(d.computeLineDiffs())

	// Reset style at the end to ensure terminal is in a known state
	sb.WriteString("\x1b[0m")

	return sb.String()
}

// fullRedraw creates an ANSI string for a full screen redraw
func (d *Differ) fullRedraw() string {
	var sb strings.Builder
	sb.WriteString("\x1b[2J") // Clear screen
	sb.WriteString("\x1b[H")  // Move cursor to home

	// Add all non-empty lines
	for y := 0; y < d.newScreen.height; y++ {
		lineContent := d.renderLine(d.newScreen, y)
		if lineContent != "" {
			sb.WriteString(fmt.Sprintf("\x1b[%d;1H", y+1)) // Move cursor to start of line
			sb.WriteString(lineContent)
		}
	}
	sb.WriteString("\x1b[0m") // Ensure reset at the end of full redraw
	return sb.String()
}

// computeLineDiffs computes line-by-line differences and returns an ANSI string
func (d *Differ) computeLineDiffs() string {
	var sb strings.Builder

	for y := 0; y < d.newScreen.height; y++ {
		// Compare lines
		if !d.linesEqual(y) {
			// Line changed, send update
			lineContent := d.renderLine(d.newScreen, y)
			sb.WriteString(fmt.Sprintf("\x1b[%d;1H", y+1)) // Move cursor to start of line
			sb.WriteString("\x1b[0K") // Clear to end of line
			sb.WriteString(lineContent)
		}
	}
	return sb.String()
}

// linesEqual checks if two lines are equal
func (d *Differ) linesEqual(y int) bool {
	if y >= d.oldScreen.height || y >= d.newScreen.height {
		return false
	}

	oldLine := d.oldScreen.lines[y]
	newLine := d.newScreen.lines[y]

	if len(oldLine) != len(newLine) {
		return false
	}

	for x := 0; x < len(oldLine); x++ {
		// Compare rune and style
		if oldLine[x].Rune != newLine[x].Rune || !stylesEqual(oldLine[x].Style, newLine[x].Style) {
			return false
		}
	}

	return true
}

// renderLine renders a line to a string with ANSI codes
func (d *Differ) renderLine(screen *Screen, y int) string {
	if y >= screen.height {
		return ""
	}

	line := screen.lines[y]
	result := ""
	currentStyle := NewStyle()

	// Find the last non-space character
	lastNonSpace := -1
	for i := len(line) - 1; i >= 0; i-- {
		if line[i].Rune != ' ' {
			lastNonSpace = i
			break
		}
	}

	// If entire line is spaces, return empty
	if lastNonSpace == -1 {
		return ""
	}

	// Render up to last non-space
	for x := 0; x <= lastNonSpace; x++ {
		cell := line[x]

		// Check if style changed
		if !stylesEqual(currentStyle, cell.Style) {
			// Emit style change
			result += renderStyleTransition(currentStyle, cell.Style)
			currentStyle = cell.Style
		}

		// Emit character
		result += string(cell.Rune)
	}



	return result
}

// stylesEqual compares two styles for equality
func stylesEqual(a, b Style) bool {
	return a.ToANSI() == b.ToANSI()
}

// isDefaultStyle checks if a style is the default (no attributes)
func isDefaultStyle(s Style) bool {
	return s.ToANSI() == "" // Default style should have no ANSI codes
}

// renderStyleTransition renders ANSI codes to transition from one style to another
func renderStyleTransition(from, to Style) string {
	if from.ToANSI() == to.ToANSI() {
		return "" // No change in style, no transition needed
	}

	// If the target style is default, just reset
	if isDefaultStyle(to) {
		return "\x1b[0m"
	}

	// Otherwise, reset and apply the new style
	return fmt.Sprintf("\x1b[0m\x1b[%sm", to.ToANSI())
}

// ScreenDiffer manages stateful diffing between screen updates
type ScreenDiffer struct {
	width     int
	height    int
	oldScreen *Screen
	differ    *Differ
}

// NewScreenDiffer creates a new screen differ
func NewScreenDiffer(width, height int) *ScreenDiffer {
	return &ScreenDiffer{
		width:  width,
		height: height,
		differ: NewDiffer(),
	}
}

// Update computes diff operations for a new screen state and returns an ANSI string
func (sd *ScreenDiffer) Update(content string) string {
	// Create new screen and render content
	newScreen := NewScreen(sd.width, sd.height)
	newScreen.RenderFromString(content)

	// Compute diff
	ansiOutput := sd.differ.Diff(sd.oldScreen, newScreen)

	// Update old screen
	sd.oldScreen = newScreen

	return ansiOutput
}

// Resize updates the screen dimensions
func (sd *ScreenDiffer) Resize(width, height int) {
	sd.width = width
	sd.height = height
	sd.oldScreen = nil // Force full redraw on next update
}

// Reset clears the differ state
func (sd *ScreenDiffer) Reset() {
	sd.oldScreen = nil
}
