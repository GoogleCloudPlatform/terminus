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

package layout

import (
	"strings"
)

// BoxStyle represents different box drawing styles
type BoxStyle int

const (
	BoxStyleSingle BoxStyle = iota
	BoxStyleDouble
	BoxStyleRounded
	BoxStyleBold
	BoxStyleASCII
)

// boxChars defines the characters for different box styles
var boxChars = map[BoxStyle]struct {
	TopLeft     string
	TopRight    string
	BottomLeft  string
	BottomRight string
	Horizontal  string
	Vertical    string
	Cross       string
	TeeTop      string
	TeeBottom   string
	TeeLeft     string
	TeeRight    string
}{
	BoxStyleSingle: {
		TopLeft:     "┌",
		TopRight:    "┐",
		BottomLeft:  "└",
		BottomRight: "┘",
		Horizontal:  "─",
		Vertical:    "│",
		Cross:       "┼",
		TeeTop:      "┬",
		TeeBottom:   "┴",
		TeeLeft:     "├",
		TeeRight:    "┤",
	},
	BoxStyleDouble: {
		TopLeft:     "╔",
		TopRight:    "╗",
		BottomLeft:  "╚",
		BottomRight: "╝",
		Horizontal:  "═",
		Vertical:    "║",
		Cross:       "╬",
		TeeTop:      "╦",
		TeeBottom:   "╩",
		TeeLeft:     "╠",
		TeeRight:    "╣",
	},
	BoxStyleRounded: {
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "╰",
		BottomRight: "╯",
		Horizontal:  "─",
		Vertical:    "│",
		Cross:       "┼",
		TeeTop:      "┬",
		TeeBottom:   "┴",
		TeeLeft:     "├",
		TeeRight:    "┤",
	},
	BoxStyleBold: {
		TopLeft:     "┏",
		TopRight:    "┓",
		BottomLeft:  "┗",
		BottomRight: "┛",
		Horizontal:  "━",
		Vertical:    "┃",
		Cross:       "╋",
		TeeTop:      "┳",
		TeeBottom:   "┻",
		TeeLeft:     "┣",
		TeeRight:    "┫",
	},
	BoxStyleASCII: {
		TopLeft:     "+",
		TopRight:    "+",
		BottomLeft:  "+",
		BottomRight: "+",
		Horizontal:  "-",
		Vertical:    "|",
		Cross:       "+",
		TeeTop:      "+",
		TeeBottom:   "+",
		TeeLeft:     "+",
		TeeRight:    "+",
	},
}

// Box represents a box with content
type Box struct {
	content     string
	width       int
	height      int
	style       BoxStyle
	title       string
	padding     Padding
	borderColor string
}

// Padding represents spacing inside a box
type Padding struct {
	Top    int
	Right  int
	Bottom int
	Left   int
}

// NewBox creates a new box with content
func NewBox(content string) *Box {
	lines := strings.Split(content, "\n")
	maxWidth := 0
	for _, line := range lines {
		// Use visible length to exclude ANSI escape sequences
		lineWidth := visibleLength(line)
		if lineWidth > maxWidth {
			maxWidth = lineWidth
		}
	}

	return &Box{
		content: content,
		width:   maxWidth,
		height:  len(lines),
		style:   BoxStyleSingle,
		padding: Padding{0, 0, 0, 0},
	}
}

// WithStyle sets the box style
func (b *Box) WithStyle(style BoxStyle) *Box {
	b.style = style
	return b
}

// WithTitle sets the box title
func (b *Box) WithTitle(title string) *Box {
	b.title = title
	return b
}

// WithPadding sets the box padding
func (b *Box) WithPadding(top, right, bottom, left int) *Box {
	b.padding = Padding{top, right, bottom, left}
	return b
}

// WithUniformPadding sets uniform padding on all sides
func (b *Box) WithUniformPadding(padding int) *Box {
	b.padding = Padding{padding, padding, padding, padding}
	return b
}

// WithWidth sets a fixed width for the box
func (b *Box) WithWidth(width int) *Box {
	b.width = width
	return b
}

// WithHeight sets a fixed height for the box
func (b *Box) WithHeight(height int) *Box {
	b.height = height
	return b
}

// Render renders the box as a string
func (b *Box) Render() string {
	chars := boxChars[b.style]

	// Calculate inner dimensions
	innerWidth := b.width + b.padding.Left + b.padding.Right

	var result strings.Builder

	// Top border
	result.WriteString(chars.TopLeft)
	if b.title != "" && visibleLength(b.title) < innerWidth-2 {
		titleLen := visibleLength(b.title)
		titlePadding := (innerWidth - titleLen - 2) / 2
		result.WriteString(strings.Repeat(chars.Horizontal, titlePadding))
		result.WriteString(" " + b.title + " ")
		result.WriteString(strings.Repeat(chars.Horizontal, innerWidth-titleLen-2-titlePadding))
	} else {
		result.WriteString(strings.Repeat(chars.Horizontal, innerWidth))
	}
	result.WriteString(chars.TopRight)
	result.WriteString("\n")

	// Content lines
	lines := strings.Split(b.content, "\n")

	// Top padding
	for i := 0; i < b.padding.Top; i++ {
		result.WriteString(chars.Vertical)
		result.WriteString(strings.Repeat(" ", innerWidth))
		result.WriteString(chars.Vertical)
		result.WriteString("\n")
	}

	// Content with padding
	for i := 0; i < b.height; i++ {
		result.WriteString(chars.Vertical)
		result.WriteString(strings.Repeat(" ", b.padding.Left))

		if i < len(lines) {
			line := lines[i]
			lineLen := visibleLength(line)
			// Don't truncate styled strings at byte level, it could break ANSI sequences
			if lineLen > b.width {
				// For now, just use the line as-is
				// TODO: Implement proper ANSI-aware truncation
				result.WriteString(line)
			} else {
				result.WriteString(line)
				result.WriteString(strings.Repeat(" ", b.width-lineLen))
			}
		} else {
			result.WriteString(strings.Repeat(" ", b.width))
		}

		result.WriteString(strings.Repeat(" ", b.padding.Right))
		result.WriteString(chars.Vertical)
		result.WriteString("\n")
	}

	// Bottom padding
	for i := 0; i < b.padding.Bottom; i++ {
		result.WriteString(chars.Vertical)
		result.WriteString(strings.Repeat(" ", innerWidth))
		result.WriteString(chars.Vertical)
		result.WriteString("\n")
	}

	// Bottom border
	result.WriteString(chars.BottomLeft)
	result.WriteString(strings.Repeat(chars.Horizontal, innerWidth))
	result.WriteString(chars.BottomRight)

	return result.String()
}

// DrawBox is a convenience function to draw a box around content
func DrawBox(content string, style BoxStyle) string {
	return NewBox(content).WithStyle(style).Render()
}

// DrawBoxWithTitle draws a box with a title
func DrawBoxWithTitle(content, title string, style BoxStyle) string {
	return NewBox(content).WithStyle(style).WithTitle(title).Render()
}

// HorizontalLine draws a horizontal line
func HorizontalLine(width int, style BoxStyle) string {
	chars := boxChars[style]
	return strings.Repeat(chars.Horizontal, width)
}

// VerticalLine draws a vertical line
func VerticalLine(height int, style BoxStyle) string {
	chars := boxChars[style]
	var result strings.Builder
	for i := 0; i < height; i++ {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(chars.Vertical)
	}
	return result.String()
}
