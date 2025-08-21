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

func TestKeyMsgString(t *testing.T) {
	tests := []struct {
		name     string
		keyMsg   KeyMsg
		expected string
	}{
		{
			name:     "Character input",
			keyMsg:   KeyMsg{Type: KeyRunes, Runes: []rune{'a'}},
			expected: "a",
		},
		{
			name:     "Multiple characters",
			keyMsg:   KeyMsg{Type: KeyRunes, Runes: []rune{'h', 'e', 'l', 'l', 'o'}},
			expected: "hello",
		},
		{
			name:     "Empty runes",
			keyMsg:   KeyMsg{Type: KeyRunes, Runes: []rune{}},
			expected: "",
		},
		{
			name:     "Enter key",
			keyMsg:   KeyMsg{Type: KeyEnter},
			expected: "enter",
		},
		{
			name:     "Space key",
			keyMsg:   KeyMsg{Type: KeySpace},
			expected: "space",
		},
		{
			name:     "Backspace key",
			keyMsg:   KeyMsg{Type: KeyBackspace},
			expected: "backspace",
		},
		{
			name:     "Tab key",
			keyMsg:   KeyMsg{Type: KeyTab},
			expected: "tab",
		},
		{
			name:     "Escape key",
			keyMsg:   KeyMsg{Type: KeyEsc},
			expected: "esc",
		},
		{
			name:     "Arrow up",
			keyMsg:   KeyMsg{Type: KeyUp},
			expected: "up",
		},
		{
			name:     "Arrow down",
			keyMsg:   KeyMsg{Type: KeyDown},
			expected: "down",
		},
		{
			name:     "Arrow left",
			keyMsg:   KeyMsg{Type: KeyLeft},
			expected: "left",
		},
		{
			name:     "Arrow right",
			keyMsg:   KeyMsg{Type: KeyRight},
			expected: "right",
		},
		{
			name:     "Ctrl+C",
			keyMsg:   KeyMsg{Type: KeyCtrlC},
			expected: "ctrl+c",
		},
		{
			name:     "F1 key",
			keyMsg:   KeyMsg{Type: KeyF1},
			expected: "f1",
		},
		{
			name:     "F12 key",
			keyMsg:   KeyMsg{Type: KeyF12},
			expected: "f12",
		},
		{
			name:     "Unknown key",
			keyMsg:   KeyMsg{Type: KeyType(999)},
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.keyMsg.String()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestKeyMsgModifiers(t *testing.T) {
	msg := KeyMsg{
		Type:  KeyRunes,
		Runes: []rune{'a'},
		Alt:   true,
		Ctrl:  true,
		Shift: true,
	}

	if !msg.Alt || !msg.Ctrl || !msg.Shift {
		t.Error("Modifiers should be set correctly")
	}
}

func TestWindowSizeMsg(t *testing.T) {
	msg := WindowSizeMsg{
		Width:  80,
		Height: 24,
	}

	if msg.Width != 80 {
		t.Errorf("Expected width 80, got %d", msg.Width)
	}
	if msg.Height != 24 {
		t.Errorf("Expected height 24, got %d", msg.Height)
	}
}