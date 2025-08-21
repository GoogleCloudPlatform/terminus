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
	"strings"
	"testing"
)

func TestStyleBuilder(t *testing.T) {
	tests := []struct {
		name     string
		style    Style
		text     string
		contains []string
		exact    bool
		expected string
	}{
		{
			name:     "Plain text",
			style:    New(),
			text:     "Hello",
			expected: "Hello",
			exact:    true,
		},
		{
			name:     "Bold text",
			style:    New().Bold(true),
			text:     "Bold",
			contains: []string{"\x1b[", "1", "Bold", "\x1b[0m"},
		},
		{
			name:     "Italic text",
			style:    New().Italic(true),
			text:     "Italic",
			contains: []string{"\x1b[", "3", "Italic", "\x1b[0m"},
		},
		{
			name:     "Underline text",
			style:    New().Underline(true),
			text:     "Underline",
			contains: []string{"\x1b[", "4", "Underline", "\x1b[0m"},
		},
		{
			name:     "Multiple attributes",
			style:    New().Bold(true).Italic(true).Underline(true),
			text:     "Multi",
			contains: []string{"\x1b[", "1", "3", "4", "Multi", "\x1b[0m"},
		},
		{
			name:     "Foreground color",
			style:    New().Foreground(Red),
			text:     "Red",
			contains: []string{"\x1b[", "31", "Red", "\x1b[0m"},
		},
		{
			name:     "Background color",
			style:    New().Background(Blue),
			text:     "Blue",
			contains: []string{"\x1b[", "44", "Blue", "\x1b[0m"},
		},
		{
			name:     "Both colors",
			style:    New().Foreground(White).Background(Black),
			text:     "Text",
			contains: []string{"\x1b[", "37", "40", "Text", "\x1b[0m"},
		},
		{
			name:     "All attributes",
			style:    New().Bold(true).Italic(true).Underline(true).CrossOut(true).Foreground(Green).Background(Yellow),
			text:     "All",
			contains: []string{"\x1b[", "1", "3", "4", "9", "32", "43", "All", "\x1b[0m"},
		},
		{
			name:     "Empty text",
			style:    New().Bold(true),
			text:     "",
			expected: "",
			exact:    true,
		},
		{
			name:     "Faint text",
			style:    New().Faint(true),
			text:     "Faint",
			contains: []string{"\x1b[", "2", "Faint", "\x1b[0m"},
		},
		{
			name:     "Reverse video",
			style:    New().Reverse(true),
			text:     "Reverse",
			contains: []string{"\x1b[", "7", "Reverse", "\x1b[0m"},
		},
		{
			name:     "Blink text",
			style:    New().Blink(true),
			text:     "Blink",
			contains: []string{"\x1b[", "5", "Blink", "\x1b[0m"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.style.Render(tt.text)
			
			if tt.exact {
				if result != tt.expected {
					t.Errorf("Expected exact: %q, got %q", tt.expected, result)
				}
			} else {
				for _, substr := range tt.contains {
					if !strings.Contains(result, substr) {
						t.Errorf("Expected result to contain %q, but it doesn't: %q", substr, result)
					}
				}
			}
		})
	}
}

func TestStyleString(t *testing.T) {
	tests := []struct {
		name     string
		style    Style
		expected string
	}{
		{
			name:     "Empty style",
			style:    New(),
			expected: "Style{}",
		},
		{
			name:     "Bold only",
			style:    New().Bold(true),
			expected: "Style{bold}",
		},
		{
			name:     "Multiple attributes",
			style:    New().Bold(true).Italic(true),
			expected: "Style{bold, italic}",
		},
		{
			name:     "With colors",
			style:    New().Foreground(Red).Background(Blue),
			expected: "Style{fg:red, bg:blue}",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.style.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestStyleChaining(t *testing.T) {
	// Test that style methods can be chained
	style := New().
		Bold(true).
		Italic(true).
		Underline(true).
		CrossOut(true).
		Faint(true).
		Reverse(true).
		Blink(true).
		Foreground(Red).
		Background(Blue)
	
	// Test that all attributes are set
	result := style.String()
	expected := []string{"bold", "faint", "italic", "underline", "crossout", "reverse", "blink", "fg:red", "bg:blue"}
	
	for _, attr := range expected {
		if !strings.Contains(result, attr) {
			t.Errorf("Expected style to contain %q, but it doesn't: %s", attr, result)
		}
	}
}

func TestStyleImmutability(t *testing.T) {
	// Test that styles are immutable
	original := New()
	bold := original.Bold(true)
	
	// Original should not be modified
	if original.String() != "Style{}" {
		t.Error("Original style was modified")
	}
	
	if bold.String() != "Style{bold}" {
		t.Error("Bold style not correctly set")
	}
}