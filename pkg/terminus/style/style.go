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

// Render applies the style to the given text and returns styled string
func (s Style) Render(text string) string {
	if text == "" {
		return ""
	}
	
	// Build style codes
	var codes []string
	
	// Reset all styles first
	startCodes := []string{"0"}
	
	// Text attributes
	if s.bold {
		startCodes = append(startCodes, "1")
	}
	if s.faint {
		startCodes = append(startCodes, "2")
	}
	if s.italic {
		startCodes = append(startCodes, "3")
	}
	if s.underline {
		startCodes = append(startCodes, "4")
	}
	if s.blink {
		startCodes = append(startCodes, "5")
	}
	if s.reverse {
		startCodes = append(startCodes, "7")
	}
	if s.crossOut {
		startCodes = append(startCodes, "9")
	}
	
	// Colors
	if s.foreground != nil {
		startCodes = append(startCodes, s.foreground.Foreground())
	}
	if s.background != nil {
		startCodes = append(startCodes, s.background.Background())
	}
	
	// Apply styles
	if len(startCodes) > 1 || startCodes[0] != "0" {
		codes = append(codes, strings.Join(startCodes, ";"))
		return fmt.Sprintf("\x1b[%sm%s\x1b[0m", strings.Join(codes, ";"), text)
	}
	
	return text
}

// String returns the style as a string representation
func (s Style) String() string {
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
	if s.foreground != nil {
		attrs = append(attrs, fmt.Sprintf("fg:%s", s.foreground.String()))
	}
	if s.background != nil {
		attrs = append(attrs, fmt.Sprintf("bg:%s", s.background.String()))
	}
	
	if len(attrs) == 0 {
		return "Style{}"
	}
	
	return fmt.Sprintf("Style{%s}", strings.Join(attrs, ", "))
}