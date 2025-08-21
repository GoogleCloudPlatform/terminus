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
	"testing"
)

func TestNewBox(t *testing.T) {
	content := "Hello\nWorld"
	box := NewBox(content)

	if box.content != content {
		t.Errorf("Expected content %q, got %q", content, box.content)
	}

	if box.width != 5 {
		t.Errorf("Expected width 5, got %d", box.width)
	}

	if box.height != 2 {
		t.Errorf("Expected height 2, got %d", box.height)
	}
}

func TestBoxRender(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		setup    func(*Box) *Box
		contains []string
	}{
		{
			name:    "Simple box",
			content: "Hello",
			setup:   func(b *Box) *Box { return b },
			contains: []string{
				"┌─────┐",
				"│Hello│",
				"└─────┘",
			},
		},
		{
			name:    "Box with title",
			content: "Content",
			setup: func(b *Box) *Box {
				return b.WithTitle("Title").WithWidth(12)
			},
			contains: []string{
				" Title ",
				"│Content",
			},
		},
		{
			name:    "Box with padding",
			content: "Text",
			setup: func(b *Box) *Box {
				return b.WithUniformPadding(1)
			},
			contains: []string{
				"┌──────┐",
				"│      │",
				"│ Text │",
				"│      │",
				"└──────┘",
			},
		},
		{
			name:    "Double style box",
			content: "Test",
			setup: func(b *Box) *Box {
				return b.WithStyle(BoxStyleDouble)
			},
			contains: []string{
				"╔════╗",
				"║Test║",
				"╚════╝",
			},
		},
		{
			name:    "ASCII style box",
			content: "ASCII",
			setup: func(b *Box) *Box {
				return b.WithStyle(BoxStyleASCII)
			},
			contains: []string{
				"+-----+",
				"|ASCII|",
				"+-----+",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			box := NewBox(tt.content)
			box = tt.setup(box)
			result := box.Render()

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q\nGot:\n%s", expected, result)
				}
			}
		})
	}
}

func TestBoxWithFixedDimensions(t *testing.T) {
	box := NewBox("Hi").WithWidth(10).WithHeight(3)
	result := box.Render()

	lines := strings.Split(result, "\n")
	if len(lines) != 5 { // 3 content lines + 2 border lines
		t.Errorf("Expected 5 lines, got %d", len(lines))
	}

	// Check that content lines are padded to width
	contentLine := lines[1]
	if !strings.Contains(contentLine, "Hi        ") {
		t.Errorf("Expected padded content, got %q", contentLine)
	}
}

func TestDrawBox(t *testing.T) {
	result := DrawBox("Quick test", BoxStyleSingle)
	if !strings.Contains(result, "│Quick test│") {
		t.Error("DrawBox should create a simple box")
	}
}

func TestDrawBoxWithTitle(t *testing.T) {
	result := DrawBoxWithTitle("Longer Content", "Title", BoxStyleSingle)
	if !strings.Contains(result, " Title ") {
		t.Errorf("DrawBoxWithTitle should include the title, got:\n%s", result)
	}
}

func TestHorizontalLine(t *testing.T) {
	line := HorizontalLine(5, BoxStyleSingle)
	if line != "─────" {
		t.Errorf("Expected 5 horizontal line characters, got %q", line)
	}

	doubleLine := HorizontalLine(3, BoxStyleDouble)
	if doubleLine != "═══" {
		t.Errorf("Expected double line characters, got %q", doubleLine)
	}
}

func TestVerticalLine(t *testing.T) {
	line := VerticalLine(3, BoxStyleSingle)
	expected := "│\n│\n│"
	if line != expected {
		t.Errorf("Expected vertical line %q, got %q", expected, line)
	}
}
