package style

import (
	"testing"
)

func TestColorFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Color
	}{
		// Named colors
		{name: "Named red", input: "red", expected: Red},
		{name: "Named blue", input: "blue", expected: Blue},
		{name: "Named green", input: "green", expected: Green},
		{name: "Named uppercase", input: "RED", expected: Red},
		{name: "Named with spaces", input: "  red  ", expected: Red},
		{name: "Named gray", input: "gray", expected: BrightBlack},
		{name: "Named grey", input: "grey", expected: BrightBlack},
		{name: "Named bright", input: "brightred", expected: BrightRed},
		
		// Hex colors
		{name: "Hex 6 digit", input: "#FF0000", expected: RGB(255, 0, 0)},
		{name: "Hex 3 digit", input: "#F00", expected: RGB(255, 0, 0)},
		{name: "Hex lowercase", input: "#ff0000", expected: RGB(255, 0, 0)},
		{name: "Hex green", input: "#00FF00", expected: RGB(0, 255, 0)},
		{name: "Hex blue", input: "#0000FF", expected: RGB(0, 0, 255)},
		{name: "Hex mixed", input: "#123456", expected: RGB(18, 52, 86)},
		{name: "Hex 3 mixed", input: "#369", expected: RGB(51, 102, 153)},
		
		// ANSI 256
		{name: "ANSI 0", input: "0", expected: ANSI256(0)},
		{name: "ANSI 255", input: "255", expected: ANSI256(255)},
		{name: "ANSI 128", input: "128", expected: ANSI256(128)},
		
		// Invalid
		{name: "Invalid name", input: "notacolor", expected: White},
		{name: "Invalid hex", input: "#GGGGGG", expected: White},
		{name: "Invalid number", input: "256", expected: White},
		{name: "Empty string", input: "", expected: White},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ColorFromString(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestRGB(t *testing.T) {
	tests := []struct {
		name string
		r, g, b int
		expected string
	}{
		{name: "Black", r: 0, g: 0, b: 0, expected: "0;0;0"},
		{name: "White", r: 255, g: 255, b: 255, expected: "255;255;255"},
		{name: "Red", r: 255, g: 0, b: 0, expected: "255;0;0"},
		{name: "Clamped high", r: 300, g: 300, b: 300, expected: "255;255;255"},
		{name: "Clamped low", r: -10, g: -10, b: -10, expected: "0;0;0"},
		{name: "Mixed", r: 100, g: 150, b: 200, expected: "100;150;200"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color := RGB(tt.r, tt.g, tt.b)
			if color.value != tt.expected {
				t.Errorf("Expected value %s, got %s", tt.expected, color.value)
			}
			if color.colorType != rgbColor {
				t.Error("Expected rgbColor type")
			}
		})
	}
}

func TestANSI256(t *testing.T) {
	tests := []struct {
		name     string
		n        int
		expected string
		valid    bool
	}{
		{name: "Min value", n: 0, expected: "0", valid: true},
		{name: "Max value", n: 255, expected: "255", valid: true},
		{name: "Mid value", n: 128, expected: "128", valid: true},
		{name: "Out of range high", n: 256, valid: false},
		{name: "Out of range low", n: -1, valid: false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color := ANSI256(tt.n)
			if tt.valid {
				if color.value != tt.expected {
					t.Errorf("Expected value %s, got %s", tt.expected, color.value)
				}
				if color.colorType != ansi256Color {
					t.Error("Expected ansi256Color type")
				}
			} else {
				if color != White {
					t.Error("Expected White for invalid input")
				}
			}
		})
	}
}

func TestColorForeground(t *testing.T) {
	tests := []struct {
		name     string
		color    Color
		expected string
	}{
		{name: "Named red", color: Red, expected: "31"},
		{name: "Named blue", color: Blue, expected: "34"},
		{name: "Bright green", color: BrightGreen, expected: "92"},
		{name: "ANSI 256", color: ANSI256(100), expected: "38;5;100"},
		{name: "RGB", color: RGB(100, 150, 200), expected: "38;2;100;150;200"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.color.Foreground()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestColorBackground(t *testing.T) {
	tests := []struct {
		name     string
		color    Color
		expected string
	}{
		{name: "Named red", color: Red, expected: "41"},
		{name: "Named blue", color: Blue, expected: "44"},
		{name: "Bright green", color: BrightGreen, expected: "102"},
		{name: "ANSI 256", color: ANSI256(100), expected: "48;5;100"},
		{name: "RGB", color: RGB(100, 150, 200), expected: "48;2;100;150;200"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.color.Background()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestColorString(t *testing.T) {
	tests := []struct {
		name     string
		color    Color
		expected string
	}{
		{name: "Named red", color: Red, expected: "red"},
		{name: "Named blue", color: Blue, expected: "blue"},
		{name: "ANSI 256", color: ANSI256(100), expected: "ansi256(100)"},
		{name: "RGB", color: RGB(100, 150, 200), expected: "rgb(100;150;200)"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.color.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		v, min, max, expected int
	}{
		{10, 0, 100, 10},
		{-10, 0, 100, 0},
		{110, 0, 100, 100},
		{50, 0, 100, 50},
	}
	
	for _, tt := range tests {
		result := clamp(tt.v, tt.min, tt.max)
		if result != tt.expected {
			t.Errorf("clamp(%d, %d, %d) = %d, expected %d", tt.v, tt.min, tt.max, result, tt.expected)
		}
	}
}