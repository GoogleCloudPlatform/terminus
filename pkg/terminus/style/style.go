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

package style

import (
	"fmt"
	"strings"
)

// Style represents text styling attributes
type Style struct {
	bold      bool
	faint     bool
	italic    bool
	underline bool
	crossOut  bool
	reverse   bool
	blink     bool
	
	foreground *Color
	background *Color
}

// New creates a new style with default settings
func New() Style {
	return Style{}
}

// Bold sets the bold attribute
func (s Style) Bold(v bool) Style {
	s.bold = v
	return s
}

// Faint sets the faint/dim attribute
func (s Style) Faint(v bool) Style {
	s.faint = v
	return s
}

// Italic sets the italic attribute
func (s Style) Italic(v bool) Style {
	s.italic = v
	return s
}

// Underline sets the underline attribute
func (s Style) Underline(v bool) Style {
	s.underline = v
	return s
}

// CrossOut sets the strikethrough attribute
func (s Style) CrossOut(v bool) Style {
	s.crossOut = v
	return s
}

// Reverse sets the reverse video attribute
func (s Style) Reverse(v bool) Style {
	s.reverse = v
	return s
}

// Blink sets the blink attribute
func (s Style) Blink(v bool) Style {
	s.blink = v
	return s
}

// Foreground sets the foreground color
func (s Style) Foreground(c Color) Style {
	s.foreground = &c
	return s
}

// Background sets the background color
func (s Style) Background(c Color) Style {
	s.background = &c
	return s
}


// ToANSI generates the ANSI escape code sequence for the style, without the leading ESC[ and trailing m.
// E.g., "1;31" for bold red foreground.
func (s Style) ToANSI() string {
	var codes []string

	if s.bold {
		codes = append(codes, "1")
	} else if s.faint {
		codes = append(codes, "2")
	}
	// No explicit code for normal intensity (22) needed when building from scratch, as it's implied by absence

	if s.italic {
		codes = append(codes, "3")
	}
	if s.underline {
		codes = append(codes, "4")
	}
	// Blink
	if s.blink {
		codes = append(codes, "5")
	}
	// Reverse video
	if s.reverse {
		codes = append(codes, "7")
	}
	// Hidden
	// if s.hidden {
	// 	codes = append(codes, "8")
	// }
	// Cross-out
	if s.crossOut {
		codes = append(codes, "9")
	}

	// Colors
	if s.foreground != nil && !s.foreground.IsZero() {
		codes = append(codes, s.foreground.AnsiCode(false))
	}
	if s.background != nil && !s.background.IsZero() {
		codes = append(codes, s.background.AnsiCode(true))
	}

	return strings.Join(codes, ";")
}

// Render applies the style to the given text and returns styled string
func (s Style) Render(text string) string {
	if text == "" {
		return ""
	}

	ansiCode := s.ToANSI()
	if ansiCode == "" {
		return text
	}

	return fmt.Sprintf("\x1b[%sm%s\x1b[0m", ansiCode, text)
}

// String returns the style as a string representation
func (s Style) String() string {
	// This method is primarily for comparison in diffing logic
	// For rendering to ANSI, use ToANSI()
	var attrs []string

	if s.bold {
		attrs = append(attrs, "bold")
	}
	if s.faint {
		attrs = append(attrs, "faint")
	}
	if s.italic {
		attrs = append(attrs, "italic")
	}
	if s.underline {
		attrs = append(attrs, "underline")
	}
	if s.crossOut {
		attrs = append(attrs, "crossout")
	}
	if s.reverse {
		attrs = append(attrs, "reverse")
	}
	if s.blink {
		attrs = append(attrs, "blink")
	}
	if s.foreground != nil && !s.foreground.IsZero() {
		attrs = append(attrs, fmt.Sprintf("fg:%s", s.foreground.String()))
	}
	if s.background != nil && !s.background.IsZero() {
		attrs = append(attrs, fmt.Sprintf("bg:%s", s.background.String()))
	}

	if len(attrs) == 0 {
		return "Style{}"
	}

	return fmt.Sprintf("Style{%s}", strings.Join(attrs, ", "))
}