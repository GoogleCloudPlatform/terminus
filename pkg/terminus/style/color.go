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

// Color represents a terminal color
type Color struct {
	value     string
	colorType colorType
}

type colorType int

const (
	namedColor colorType = iota
	ansi256Color
	rgbColor
)

// Predefined colors
var (
	Black   = Color{value: "30", colorType: namedColor}
	Red     = Color{value: "31", colorType: namedColor}
	Green   = Color{value: "32", colorType: namedColor}
	Yellow  = Color{value: "33", colorType: namedColor}
	Blue    = Color{value: "34", colorType: namedColor}
	Magenta = Color{value: "35", colorType: namedColor}
	Cyan    = Color{value: "36", colorType: namedColor}
	White   = Color{value: "37", colorType: namedColor}
	
	// Bright colors
	BrightBlack   = Color{value: "90", colorType: namedColor}
	BrightRed     = Color{value: "91", colorType: namedColor}
	BrightGreen   = Color{value: "92", colorType: namedColor}
	BrightYellow  = Color{value: "93", colorType: namedColor}
	BrightBlue    = Color{value: "94", colorType: namedColor}
	BrightMagenta = Color{value: "95", colorType: namedColor}
	BrightCyan    = Color{value: "96", colorType: namedColor}
	BrightWhite   = Color{value: "97", colorType: namedColor}
)

// namedColors maps color names to Color values
var namedColors = map[string]Color{
	"black":         Black,
	"red":           Red,
	"green":         Green,
	"yellow":        Yellow,
	"blue":          Blue,
	"magenta":       Magenta,
	"cyan":          Cyan,
	"white":         White,
	"brightblack":   BrightBlack,
	"brightred":     BrightRed,
	"brightgreen":   BrightGreen,
	"brightyellow":  BrightYellow,
	"brightblue":    BrightBlue,
	"brightmagenta": BrightMagenta,
	"brightcyan":    BrightCyan,
	"brightwhite":   BrightWhite,
	"gray":          BrightBlack,
	"grey":          BrightBlack,
}

// ColorFromString creates a color from a string (name, hex, or number)
func ColorFromString(s string) Color {
	s = strings.ToLower(strings.TrimSpace(s))
	
	// Check named colors
	if c, ok := namedColors[s]; ok {
		return c
	}
	
	// Check hex colors (#RRGGBB or #RGB)
	if strings.HasPrefix(s, "#") {
		return parseHexColor(s)
	}
	
	// Check for ANSI 256 color (0-255)
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err == nil && n >= 0 && n <= 255 {
		return ANSI256(n)
	}
	
	// Default to white
	return White
}

// parseHexColor parses a hex color string
func parseHexColor(s string) Color {
	s = strings.TrimPrefix(s, "#")
	
	var r, g, b int
	var err error
	
	switch len(s) {
	case 3:
		// #RGB format
		_, err = fmt.Sscanf(s, "%1x%1x%1x", &r, &g, &b)
		if err != nil {
			return White
		}
		// Expand to full values
		r = r*16 + r
		g = g*16 + g
		b = b*16 + b
	case 6:
		// #RRGGBB format
		_, err = fmt.Sscanf(s, "%2x%2x%2x", &r, &g, &b)
		if err != nil {
			return White
		}
	default:
		return White
	}
	
	return RGB(r, g, b)
}

// ANSI256 creates a color from an ANSI 256 color index
func ANSI256(n int) Color {
	if n < 0 || n > 255 {
		return White
	}
	return Color{
		value:     fmt.Sprintf("%d", n),
		colorType: ansi256Color,
	}
}

// RGB creates a color from RGB values (0-255)
func RGB(r, g, b int) Color {
	// Clamp values
	r = clamp(r, 0, 255)
	g = clamp(g, 0, 255)
	b = clamp(b, 0, 255)
	
	return Color{
		value:     fmt.Sprintf("%d;%d;%d", r, g, b),
		colorType: rgbColor,
	}
}

// Foreground returns the ANSI escape code for foreground color
func (c Color) Foreground() string {
	switch c.colorType {
	case namedColor:
		return c.value
	case ansi256Color:
		return fmt.Sprintf("38;5;%s", c.value)
	case rgbColor:
		return fmt.Sprintf("38;2;%s", c.value)
	default:
		return "37" // Default white
	}
}

// Background returns the ANSI escape code for background color
func (c Color) Background() string {
	switch c.colorType {
	case namedColor:
		// Convert foreground to background codes
		var n int
		fmt.Sscanf(c.value, "%d", &n)
		if n >= 30 && n <= 37 {
			return fmt.Sprintf("%d", n+10)
		} else if n >= 90 && n <= 97 {
			return fmt.Sprintf("%d", n+10)
		}
		return "47" // Default white background
	case ansi256Color:
		return fmt.Sprintf("48;5;%s", c.value)
	case rgbColor:
		return fmt.Sprintf("48;2;%s", c.value)
	default:
		return "47" // Default white background
	}
}

// String returns a string representation of the color
func (c Color) String() string {
	switch c.colorType {
	case namedColor:
		// Find the name
		for name, color := range namedColors {
			if color.value == c.value {
				return name
			}
		}
		return c.value
	case ansi256Color:
		return fmt.Sprintf("ansi256(%s)", c.value)
	case rgbColor:
		return fmt.Sprintf("rgb(%s)", c.value)
	default:
		return "unknown"
	}
}

// clamp restricts a value to a range
func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}