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

// DiffOp represents a diff operation
type DiffOp struct {
	Type DiffOpType
	Data interface{}
}

// DiffOpType represents the type of diff operation
type DiffOpType string

const (
	DiffOpClear      DiffOpType = "clear"
	DiffOpSetCell    DiffOpType = "setCell"
	DiffOpUpdateLine DiffOpType = "updateLine"
	DiffOpScrollUp   DiffOpType = "scrollUp"
	DiffOpScrollDown DiffOpType = "scrollDown"
)

// SetCellOp represents a single cell update
type SetCellOp struct {
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Rune  string `json:"rune"`
	Style string `json:"style,omitempty"`
}

// UpdateLineOp represents a full line update
type UpdateLineOp struct {
	Y       int    `json:"y"`
	Content string `json:"content"`
}

// Differ computes differences between two screens
type Differ struct {
	oldScreen *Screen
	newScreen *Screen
}

// NewDiffer creates a new differ
func NewDiffer() *Differ {
	return &Differ{}
}

// Diff computes the differences between two screens
func (d *Differ) Diff(oldScreen, newScreen *Screen) []DiffOp {
	d.oldScreen = oldScreen
	d.newScreen = newScreen
	
	// If dimensions changed, clear and redraw
	if oldScreen == nil || 
		oldScreen.width != newScreen.width || 
		oldScreen.height != newScreen.height {
		return d.fullRedraw()
	}
	
	// Compute line-by-line differences
	return d.computeLineDiffs()
}

// fullRedraw creates diff ops for a full screen redraw
func (d *Differ) fullRedraw() []DiffOp {
	ops := []DiffOp{
		{Type: DiffOpClear},
	}
	
	// Add all non-empty lines
	for y := 0; y < d.newScreen.height; y++ {
		lineContent := d.renderLine(d.newScreen, y)
		if lineContent != "" {
			ops = append(ops, DiffOp{
				Type: DiffOpUpdateLine,
				Data: UpdateLineOp{
					Y:       y,
					Content: lineContent,
				},
			})
		}
	}
	
	return ops
}

// computeLineDiffs computes line-by-line differences
func (d *Differ) computeLineDiffs() []DiffOp {
	ops := []DiffOp{}
	
	for y := 0; y < d.newScreen.height; y++ {
		// Compare lines
		if !d.linesEqual(y) {
			// Line changed, send update
			lineContent := d.renderLine(d.newScreen, y)
			ops = append(ops, DiffOp{
				Type: DiffOpUpdateLine,
				Data: UpdateLineOp{
					Y:       y,
					Content: lineContent,
				},
			})
		}
	}
	
	return ops
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
		if oldLine[x].Rune != newLine[x].Rune {
			return false
		}
		// For now, ignore style differences in comparison
		// TODO: Compare styles when client supports it
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
	
	// Reset style at end if needed
	if !isDefaultStyle(currentStyle) {
		result += "\x1b[0m"
	}
	
	return result
}

// stylesEqual compares two styles for equality
func stylesEqual(a, b Style) bool {
	// This is a simplified comparison
	// In a real implementation, we'd compare all style attributes
	return a.String() == b.String()
}

// isDefaultStyle checks if a style is the default (no attributes)
func isDefaultStyle(s Style) bool {
	return s.String() == "Style{}"
}

// renderStyleTransition renders ANSI codes to transition from one style to another
func renderStyleTransition(from, to Style) string {
	// For simplicity, always reset and apply new style
	// A more sophisticated implementation would compute minimal transitions
	if isDefaultStyle(to) {
		return "\x1b[0m"
	}
	
	// Reset and apply new style
	// This is inefficient but simple
	result := "\x1b[0m"
	
	// Apply new style by rendering a dummy string and extracting codes
	styled := to.Render("X")
	if len(styled) > 1 {
		// Find the ANSI codes
		if styled[0] == '\x1b' {
			// Extract everything up to 'm'
			for i, r := range styled {
				if r == 'm' {
					result = styled[:i+1]
					break
				}
			}
		}
	}
	
	return result
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

// Update computes diff operations for a new screen state
func (sd *ScreenDiffer) Update(content string) []DiffOp {
	// Create new screen and render content
	newScreen := NewScreen(sd.width, sd.height)
	newScreen.RenderFromString(content)
	
	// Compute diff
	ops := sd.differ.Diff(sd.oldScreen, newScreen)
	
	// Update old screen
	sd.oldScreen = newScreen
	
	return ops
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