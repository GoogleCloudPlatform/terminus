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
	"testing"
)

func TestDiffer(t *testing.T) {
	tests := []struct {
		name     string
		oldScreen *Screen
		newScreen *Screen
		expected int // expected number of ops
		checkOps func(t *testing.T, ops []DiffOp)
	}{
		{
			name:      "First render (nil old screen)",
			oldScreen: nil,
			newScreen: func() *Screen {
				s := NewScreen(10, 3)
				s.RenderFromString("Hello\nWorld")
				return s
			}(),
			expected: 3, // Clear + 2 lines
			checkOps: func(t *testing.T, ops []DiffOp) {
				if ops[0].Type != DiffOpClear {
					t.Error("First op should be clear")
				}
				if ops[1].Type != DiffOpUpdateLine {
					t.Error("Second op should be update line")
				}
				if ops[2].Type != DiffOpUpdateLine {
					t.Error("Third op should be update line")
				}
			},
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
			expected: 0,
			checkOps: func(t *testing.T, ops []DiffOp) {
				// No ops expected
			},
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
			expected: 1,
			checkOps: func(t *testing.T, ops []DiffOp) {
				if ops[0].Type != DiffOpUpdateLine {
					t.Error("Should have update line op")
				}
				data := ops[0].Data.(UpdateLineOp)
				if data.Y != 1 {
					t.Errorf("Expected line 1, got %d", data.Y)
				}
			},
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
			expected: 2,
			checkOps: func(t *testing.T, ops []DiffOp) {
				if len(ops) != 2 {
					t.Fatalf("Expected 2 ops, got %d", len(ops))
				}
				// Should update lines 0 and 2
				data0 := ops[0].Data.(UpdateLineOp)
				data1 := ops[1].Data.(UpdateLineOp)
				if data0.Y != 0 || data1.Y != 2 {
					t.Error("Wrong lines updated")
				}
			},
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
			expected: 2, // Clear + 1 line
			checkOps: func(t *testing.T, ops []DiffOp) {
				if ops[0].Type != DiffOpClear {
					t.Error("First op should be clear for dimension change")
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			differ := NewDiffer()
			ops := differ.Diff(tt.oldScreen, tt.newScreen)
			
			if len(ops) != tt.expected {
				t.Errorf("Expected %d ops, got %d", tt.expected, len(ops))
			}
			
			if tt.checkOps != nil {
				tt.checkOps(t, ops)
			}
		})
	}
}

func TestRenderLine(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *Screen
		lineNum  int
		expected string
	}{
		{
			name: "Plain text line",
			setup: func() *Screen {
				s := NewScreen(10, 1)
				s.RenderFromString("Hello")
				return s
			},
			lineNum:  0,
			expected: "Hello",
		},
		{
			name: "Line with trailing spaces trimmed",
			setup: func() *Screen {
				s := NewScreen(10, 1)
				s.RenderFromString("Hi")
				return s
			},
			lineNum:  0,
			expected: "Hi", // Trailing spaces should be trimmed
		},
		{
			name: "Empty line returns empty string",
			setup: func() *Screen {
				s := NewScreen(10, 1)
				// Don't render anything, leave as spaces
				return s
			},
			lineNum:  0,
			expected: "",
		},
		{
			name: "Line with styled text",
			setup: func() *Screen {
				s := NewScreen(20, 1)
				s.RenderFromString("\x1b[1mBold\x1b[0m Normal")
				return s
			},
			lineNum:  0,
			expected: "\x1b[0;1mBold\x1b[0m Normal",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screen := tt.setup()
			differ := &Differ{newScreen: screen}
			result := differ.renderLine(screen, tt.lineNum)
			
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestScreenDiffer(t *testing.T) {
	tests := []struct {
		name     string
		test     func(t *testing.T)
	}{
		{
			name: "Initial update",
			test: func(t *testing.T) {
				sd := NewScreenDiffer(20, 5)
				ops := sd.Update("Hello\nWorld")
				
				// Should have clear + 2 lines
				if len(ops) < 3 {
					t.Errorf("Expected at least 3 ops, got %d", len(ops))
				}
				
				if ops[0].Type != DiffOpClear {
					t.Error("First update should start with clear")
				}
			},
		},
		{
			name: "Sequential updates",
			test: func(t *testing.T) {
				sd := NewScreenDiffer(20, 5)
				
				// First update
				ops1 := sd.Update("Line1\nLine2")
				if len(ops1) == 0 {
					t.Error("First update should have ops")
				}
				
				// Same content - no changes
				ops2 := sd.Update("Line1\nLine2")
				if len(ops2) != 0 {
					t.Error("Same content should produce no ops")
				}
				
				// Changed content
				ops3 := sd.Update("Line1\nChanged")
				if len(ops3) != 1 {
					t.Errorf("Expected 1 op for single line change, got %d", len(ops3))
				}
			},
		},
		{
			name: "Resize forces redraw",
			test: func(t *testing.T) {
				sd := NewScreenDiffer(20, 5)
				sd.Update("Hello")
				
				// Resize
				sd.Resize(30, 10)
				
				// Next update should force full redraw
				ops := sd.Update("Hello")
				if len(ops) == 0 {
					t.Error("Resize should force redraw")
				}
				if ops[0].Type != DiffOpClear {
					t.Error("Resize should start with clear")
				}
			},
		},
		{
			name: "Reset clears state",
			test: func(t *testing.T) {
				sd := NewScreenDiffer(20, 5)
				sd.Update("Hello")
				
				// Reset
				sd.Reset()
				
				// Next update should be like initial
				ops := sd.Update("Hello")
				if len(ops) == 0 {
					t.Error("Reset should force redraw")
				}
				if ops[0].Type != DiffOpClear {
					t.Error("After reset should start with clear")
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

func TestStyleTransitions(t *testing.T) {
	// Test that style transitions are properly encoded
	screen := NewScreen(20, 1)
	
	// Create content with style changes
	bold := NewStyle().Bold(true)
	red := NewStyle().Foreground(Red)
	
	// Manually set cells with different styles
	screen.SetCell(0, 0, 'A', NewStyle())      // Plain
	screen.SetCell(1, 0, 'B', bold)            // Bold
	screen.SetCell(2, 0, 'C', bold)            // Still bold
	screen.SetCell(3, 0, 'D', red)             // Red (not bold)
	screen.SetCell(4, 0, 'E', NewStyle())      // Back to plain
	
	differ := &Differ{newScreen: screen}
	line := differ.renderLine(screen, 0)
	
	// Should have style transitions
	if line == "ABCDE" {
		t.Error("Line should contain ANSI codes for style changes")
	}
	
	// Should contain reset codes
	if !contains(line, "\x1b[0m") {
		t.Error("Line should contain reset codes")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}