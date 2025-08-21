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

package widget

import (
	"testing"

	"github.com/yourusername/terminusgo/pkg/terminus"
)

func TestTextInput(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "Default state",
			test: func(t *testing.T) {
				ti := NewTextInput()
				
				if ti.Value() != "" {
					t.Error("TextInput should have empty value by default")
				}
				
				if ti.cursor != 0 {
					t.Error("TextInput cursor should be at 0 by default")
				}
				
				if !ti.showCursor {
					t.Error("TextInput should show cursor by default")
				}
				
				if ti.maxLength != 100 {
					t.Error("TextInput should have maxLength of 100 by default")
				}
			},
		},
		{
			name: "Set value",
			test: func(t *testing.T) {
				ti := NewTextInput()
				ti.SetValue("hello")
				
				if ti.Value() != "hello" {
					t.Errorf("Expected value 'hello', got '%s'", ti.Value())
				}
				
				if ti.cursor != 5 {
					t.Errorf("Expected cursor at 5, got %d", ti.cursor)
				}
			},
		},
		{
			name: "Set value adjusts cursor",
			test: func(t *testing.T) {
				ti := NewTextInput()
				ti.SetValue("hello world")
				ti.cursor = 15 // Beyond the string length
				ti.SetValue("hi")
				
				if ti.cursor != 2 {
					t.Errorf("Expected cursor adjusted to 2, got %d", ti.cursor)
				}
			},
		},
		{
			name: "Insert characters",
			test: func(t *testing.T) {
				ti := NewTextInput()
				ti.Focus()
				
				// Insert 'h'
				ti.Update(terminus.KeyMsg{Type: terminus.KeyRunes, Runes: []rune{'h'}})
				if ti.Value() != "h" {
					t.Errorf("Expected 'h', got '%s'", ti.Value())
				}
				if ti.cursor != 1 {
					t.Errorf("Expected cursor at 1, got %d", ti.cursor)
				}
				
				// Insert 'e'
				ti.Update(terminus.KeyMsg{Type: terminus.KeyRunes, Runes: []rune{'e'}})
				if ti.Value() != "he" {
					t.Errorf("Expected 'he', got '%s'", ti.Value())
				}
				if ti.cursor != 2 {
					t.Errorf("Expected cursor at 2, got %d", ti.cursor)
				}
			},
		},
		{
			name: "Insert at cursor position",
			test: func(t *testing.T) {
				ti := NewTextInput()
				ti.Focus()
				ti.SetValue("hllo")
				ti.cursor = 1
				
				// Insert 'e' between 'h' and 'l'
				ti.Update(terminus.KeyMsg{Type: terminus.KeyRunes, Runes: []rune{'e'}})
				if ti.Value() != "hello" {
					t.Errorf("Expected 'hello', got '%s'", ti.Value())
				}
				if ti.cursor != 2 {
					t.Errorf("Expected cursor at 2, got %d", ti.cursor)
				}
			},
		},
		{
			name: "Backspace",
			test: func(t *testing.T) {
				ti := NewTextInput()
				ti.Focus()
				ti.SetValue("hello")
				ti.cursor = 5
				
				// Backspace should remove 'o'
				ti.Update(terminus.KeyMsg{Type: terminus.KeyBackspace})
				if ti.Value() != "hell" {
					t.Errorf("Expected 'hell', got '%s'", ti.Value())
				}
				if ti.cursor != 4 {
					t.Errorf("Expected cursor at 4, got %d", ti.cursor)
				}
			},
		},
		{
			name: "Backspace at beginning",
			test: func(t *testing.T) {
				ti := NewTextInput()
				ti.Focus()
				ti.SetValue("hello")
				ti.cursor = 0
				
				// Backspace at beginning should do nothing
				ti.Update(terminus.KeyMsg{Type: terminus.KeyBackspace})
				if ti.Value() != "hello" {
					t.Errorf("Expected 'hello', got '%s'", ti.Value())
				}
				if ti.cursor != 0 {
					t.Errorf("Expected cursor at 0, got %d", ti.cursor)
				}
			},
		},
		{
			name: "Delete",
			test: func(t *testing.T) {
				ti := NewTextInput()
				ti.Focus()
				ti.SetValue("hello")
				ti.cursor = 1
				
				// Delete should remove 'e'
				ti.Update(terminus.KeyMsg{Type: terminus.KeyDelete})
				if ti.Value() != "hllo" {
					t.Errorf("Expected 'hllo', got '%s'", ti.Value())
				}
				if ti.cursor != 1 {
					t.Errorf("Expected cursor at 1, got %d", ti.cursor)
				}
			},
		},
		{
			name: "Delete at end",
			test: func(t *testing.T) {
				ti := NewTextInput()
				ti.Focus()
				ti.SetValue("hello")
				ti.cursor = 5
				
				// Delete at end should do nothing
				ti.Update(terminus.KeyMsg{Type: terminus.KeyDelete})
				if ti.Value() != "hello" {
					t.Errorf("Expected 'hello', got '%s'", ti.Value())
				}
				if ti.cursor != 5 {
					t.Errorf("Expected cursor at 5, got %d", ti.cursor)
				}
			},
		},
		{
			name: "Cursor movement",
			test: func(t *testing.T) {
				ti := NewTextInput()
				ti.Focus()
				ti.SetValue("hello")
				ti.cursor = 2
				
				// Move left
				ti.Update(terminus.KeyMsg{Type: terminus.KeyLeft})
				if ti.cursor != 1 {
					t.Errorf("Expected cursor at 1, got %d", ti.cursor)
				}
				
				// Move right
				ti.Update(terminus.KeyMsg{Type: terminus.KeyRight})
				if ti.cursor != 2 {
					t.Errorf("Expected cursor at 2, got %d", ti.cursor)
				}
				
				// Move to beginning
				ti.Update(terminus.KeyMsg{Type: terminus.KeyHome})
				if ti.cursor != 0 {
					t.Errorf("Expected cursor at 0, got %d", ti.cursor)
				}
				
				// Move to end
				ti.Update(terminus.KeyMsg{Type: terminus.KeyEnd})
				if ti.cursor != 5 {
					t.Errorf("Expected cursor at 5, got %d", ti.cursor)
				}
			},
		},
		{
			name: "Cursor boundaries",
			test: func(t *testing.T) {
				ti := NewTextInput()
				ti.Focus()
				ti.SetValue("hi")
				ti.cursor = 0
				
				// Left at beginning should stay at 0
				ti.Update(terminus.KeyMsg{Type: terminus.KeyLeft})
				if ti.cursor != 0 {
					t.Errorf("Expected cursor at 0, got %d", ti.cursor)
				}
				
				ti.cursor = 2
				// Right at end should stay at end
				ti.Update(terminus.KeyMsg{Type: terminus.KeyRight})
				if ti.cursor != 2 {
					t.Errorf("Expected cursor at 2, got %d", ti.cursor)
				}
			},
		},
		{
			name: "Max length validation",
			test: func(t *testing.T) {
				ti := NewTextInput()
				ti.Focus()
				ti.SetMaxLength(3)
				ti.SetValue("hi") // Cursor will be at position 2
				
				// Should accept one more character (at the end)
				ti.Update(terminus.KeyMsg{Type: terminus.KeyRunes, Runes: []rune{'!'}})
				if ti.Value() != "hi!" {
					t.Errorf("Expected 'hi!', got '%s'", ti.Value())
				}
				
				// Should reject additional characters
				ti.Update(terminus.KeyMsg{Type: terminus.KeyRunes, Runes: []rune{'?'}})
				if ti.Value() != "hi!" {
					t.Errorf("Expected 'hi!' (unchanged), got '%s'", ti.Value())
				}
			},
		},
		{
			name: "Custom validator",
			test: func(t *testing.T) {
				ti := NewTextInput()
				ti.Focus()
				
				// Only allow digits
				ti.SetValidator(func(s string) bool {
					for _, r := range s {
						if r < '0' || r > '9' {
							return false
						}
					}
					return true
				})
				
				// Should accept digits
				ti.Update(terminus.KeyMsg{Type: terminus.KeyRunes, Runes: []rune{'1', '2', '3'}})
				if ti.Value() != "123" {
					t.Errorf("Expected '123', got '%s'", ti.Value())
				}
				
				// Should reject letters
				ti.Update(terminus.KeyMsg{Type: terminus.KeyRunes, Runes: []rune{'a'}})
				if ti.Value() != "123" {
					t.Errorf("Expected '123' (unchanged), got '%s'", ti.Value())
				}
			},
		},
		{
			name: "Events",
			test: func(t *testing.T) {
				ti := NewTextInput()
				ti.Focus()
				
				var submitValue string
				var changeValue string
				
				ti.SetOnSubmit(func(value string) terminus.Cmd {
					submitValue = value
					return nil
				})
				
				ti.SetOnChange(func(value string) terminus.Cmd {
					changeValue = value
					return nil
				})
				
				// Type a character
				ti.Update(terminus.KeyMsg{Type: terminus.KeyRunes, Runes: []rune{'h'}})
				if changeValue != "h" {
					t.Errorf("Expected onChange to be called with 'h', got '%s'", changeValue)
				}
				
				// Press Enter
				ti.Update(terminus.KeyMsg{Type: terminus.KeyEnter})
				if submitValue != "h" {
					t.Errorf("Expected onSubmit to be called with 'h', got '%s'", submitValue)
				}
			},
		},
		{
			name: "Unfocused ignores input",
			test: func(t *testing.T) {
				ti := NewTextInput()
				// Don't focus the input
				
				originalValue := ti.Value()
				ti.Update(terminus.KeyMsg{Type: terminus.KeyRunes, Runes: []rune{'h'}})
				
				if ti.Value() != originalValue {
					t.Error("Unfocused TextInput should ignore input")
				}
			},
		},
		{
			name: "View with placeholder",
			test: func(t *testing.T) {
				ti := NewTextInput()
				ti.SetPlaceholder("Enter text...")
				ti.SetSize(20, 1)
				
				view := ti.View()
				// Should show placeholder when empty
				if view == "" {
					t.Error("View should not be empty with placeholder")
				}
			},
		},
		{
			name: "View with content",
			test: func(t *testing.T) {
				ti := NewTextInput()
				ti.SetValue("hello")
				ti.SetSize(20, 1)
				
				view := ti.View()
				if view == "" {
					t.Error("View should not be empty with content")
				}
			},
		},
		{
			name: "Clear method",
			test: func(t *testing.T) {
				ti := NewTextInput()
				ti.SetValue("hello world")
				ti.cursor = 5
				
				ti.Clear()
				
				if ti.Value() != "" {
					t.Errorf("Expected empty value after Clear(), got '%s'", ti.Value())
				}
				if ti.cursor != 0 {
					t.Errorf("Expected cursor at 0 after Clear(), got %d", ti.cursor)
				}
			},
		},
		{
			name: "Cursor positioning methods",
			test: func(t *testing.T) {
				ti := NewTextInput()
				ti.SetValue("hello")
				
				ti.MoveCursorToStart()
				if ti.cursor != 0 {
					t.Errorf("Expected cursor at 0, got %d", ti.cursor)
				}
				
				ti.MoveCursorToEnd()
				if ti.cursor != 5 {
					t.Errorf("Expected cursor at 5, got %d", ti.cursor)
				}
				
				ti.SetCursor(2)
				if ti.cursor != 2 {
					t.Errorf("Expected cursor at 2, got %d", ti.cursor)
				}
				
				// Test bounds checking
				ti.SetCursor(-1)
				if ti.cursor != 0 {
					t.Errorf("Expected cursor clamped to 0, got %d", ti.cursor)
				}
				
				ti.SetCursor(10)
				if ti.cursor != 5 {
					t.Errorf("Expected cursor clamped to 5, got %d", ti.cursor)
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

func TestTextInputChaining(t *testing.T) {
	// Test that all setter methods return *TextInput for method chaining
	ti := NewTextInput().
		SetValue("test").
		SetPlaceholder("Enter text...").
		SetMaxLength(50).
		SetValidator(func(s string) bool { return true }).
		SetOnSubmit(func(s string) terminus.Cmd { return nil }).
		SetOnChange(func(s string) terminus.Cmd { return nil }).
		SetStyle(terminus.NewStyle()).
		SetFocusStyle(terminus.NewStyle()).
		SetPlaceholderStyle(terminus.NewStyle()).
		SetCursorStyle(terminus.NewStyle()).
		SetCursorChar('_')
	
	if ti.Value() != "test" {
		t.Error("Method chaining should work correctly")
	}
}