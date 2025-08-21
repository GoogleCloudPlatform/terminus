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
	"strings"
	"unicode"

	"github.com/yourusername/terminusgo/pkg/terminus"
)

// TextInput is a single-line text input widget
type TextInput struct {
	Model
	
	// Input state
	value       string
	placeholder string
	cursor      int
	
	// Display settings
	showCursor   bool
	cursorChar   rune
	maxLength    int
	
	// Styling
	style           terminus.Style
	focusStyle      terminus.Style
	placeholderStyle terminus.Style
	cursorStyle     terminus.Style
	
	// Validation
	validator func(string) bool
	
	// Events
	onSubmit func(string) terminus.Cmd
	onChange func(string) terminus.Cmd
}

// NewTextInput creates a new text input widget
func NewTextInput() *TextInput {
	return &TextInput{
		Model:           NewModel(),
		showCursor:      true,
		cursorChar:      '|',
		maxLength:       100,
		style:           terminus.NewStyle(),
		focusStyle:      terminus.NewStyle().Underline(true),
		placeholderStyle: terminus.NewStyle().Faint(true),
		cursorStyle:     terminus.NewStyle().Reverse(true),
	}
}

// SetValue sets the input value
func (t *TextInput) SetValue(value string) *TextInput {
	t.value = value
	t.cursor = len(t.value) // Move cursor to end of new value
	return t
}

// Value returns the current input value
func (t *TextInput) Value() string {
	return t.value
}

// SetPlaceholder sets the placeholder text
func (t *TextInput) SetPlaceholder(placeholder string) *TextInput {
	t.placeholder = placeholder
	return t
}

// SetMaxLength sets the maximum input length
func (t *TextInput) SetMaxLength(length int) *TextInput {
	t.maxLength = length
	return t
}

// SetValidator sets a validation function
func (t *TextInput) SetValidator(validator func(string) bool) *TextInput {
	t.validator = validator
	return t
}

// SetOnSubmit sets the submit callback (triggered by Enter)
func (t *TextInput) SetOnSubmit(callback func(string) terminus.Cmd) *TextInput {
	t.onSubmit = callback
	return t
}

// SetOnChange sets the change callback (triggered on every keystroke)
func (t *TextInput) SetOnChange(callback func(string) terminus.Cmd) *TextInput {
	t.onChange = callback
	return t
}

// SetStyle sets the default style
func (t *TextInput) SetStyle(style terminus.Style) *TextInput {
	t.style = style
	return t
}

// SetFocusStyle sets the focused state style
func (t *TextInput) SetFocusStyle(style terminus.Style) *TextInput {
	t.focusStyle = style
	return t
}

// SetPlaceholderStyle sets the placeholder text style
func (t *TextInput) SetPlaceholderStyle(style terminus.Style) *TextInput {
	t.placeholderStyle = style
	return t
}

// SetCursorStyle sets the cursor style
func (t *TextInput) SetCursorStyle(style terminus.Style) *TextInput {
	t.cursorStyle = style
	return t
}

// SetCursorChar sets the cursor character
func (t *TextInput) SetCursorChar(char rune) *TextInput {
	t.cursorChar = char
	return t
}

// Init implements the Component interface
func (t *TextInput) Init() terminus.Cmd {
	return nil
}

// Update implements the Component interface
func (t *TextInput) Update(msg terminus.Msg) (terminus.Component, terminus.Cmd) {
	if !t.Focused() {
		return t, nil
	}
	
	var cmd terminus.Cmd
	
	switch msg := msg.(type) {
	case terminus.KeyMsg:
		switch msg.Type {
		case terminus.KeyEnter:
			if t.onSubmit != nil {
				cmd = t.onSubmit(t.value)
			}
			
		case terminus.KeyBackspace:
			if t.cursor > 0 && len(t.value) > 0 {
				// Remove character before cursor
				t.value = t.value[:t.cursor-1] + t.value[t.cursor:]
				t.cursor--
				if t.onChange != nil {
					cmd = t.onChange(t.value)
				}
			}
			
		case terminus.KeyDelete:
			if t.cursor < len(t.value) {
				// Remove character at cursor
				t.value = t.value[:t.cursor] + t.value[t.cursor+1:]
				if t.onChange != nil {
					cmd = t.onChange(t.value)
				}
			}
			
		case terminus.KeyLeft:
			if t.cursor > 0 {
				t.cursor--
			}
			
		case terminus.KeyRight:
			if t.cursor < len(t.value) {
				t.cursor++
			}
			
		case terminus.KeyHome:
			t.cursor = 0
			
		case terminus.KeyEnd:
			t.cursor = len(t.value)
			
		case terminus.KeySpace:
			// Handle space key
			if len(t.value) < t.maxLength {
				testValue := t.value[:t.cursor] + " " + t.value[t.cursor:]
				if t.validator == nil || t.validator(testValue) {
					t.value = testValue
					t.cursor++
					if t.onChange != nil {
						cmd = t.onChange(t.value)
					}
				}
			}
			
		case terminus.KeyRunes:
			// Insert characters at cursor position
			for _, r := range msg.Runes {
				if unicode.IsPrint(r) && len(t.value) < t.maxLength {
					// Validate input if validator is set
					testValue := t.value[:t.cursor] + string(r) + t.value[t.cursor:]
					if t.validator == nil || t.validator(testValue) {
						t.value = testValue
						t.cursor++
					}
				}
			}
			if t.onChange != nil {
				cmd = t.onChange(t.value)
			}
		}
	}
	
	return t, cmd
}

// View implements the Component interface
func (t *TextInput) View() string {
	// Determine what to display
	displayValue := t.value
	showPlaceholder := len(t.value) == 0
	
	if showPlaceholder {
		displayValue = t.placeholder
	}
	
	// Calculate display bounds based on width
	start := 0
	end := len(displayValue)
	
	// If content is longer than width, scroll to show cursor
	if len(displayValue) > t.width {
		if t.cursor >= t.width {
			start = t.cursor - t.width + 1
		}
		end = start + t.width
		if end > len(displayValue) {
			end = len(displayValue)
		}
	}
	
	// Extract visible portion
	visibleValue := ""
	if end > start {
		visibleValue = displayValue[start:end]
	}
	
	// Pad to full width
	visibleValue = padRight(visibleValue, t.width)
	
	// Build the final rendered output
	if showPlaceholder {
		return t.placeholderStyle.Render(visibleValue)
	}
	
	// Determine base style
	baseStyle := t.style
	if t.Focused() {
		baseStyle = t.focusStyle
	}
	
	// Handle cursor display
	if t.Focused() && t.showCursor {
		cursorPos := t.cursor - start
		if cursorPos >= 0 && cursorPos <= t.width {
			// Style the parts separately
			var result string
			
			// Part before cursor
			if cursorPos > 0 {
				result += baseStyle.Render(visibleValue[:cursorPos])
			}
			
			// Cursor character
			if cursorPos < len(visibleValue) {
				char := []rune(visibleValue)[cursorPos]
				if char == ' ' {
					char = t.cursorChar
				}
				result += t.cursorStyle.Render(string(char))
				
				// Part after cursor
				if cursorPos+1 < len(visibleValue) {
					result += baseStyle.Render(visibleValue[cursorPos+1:])
				}
			} else {
				// Cursor at end
				result += t.cursorStyle.Render(string(t.cursorChar))
			}
			
			return result
		}
	}
	
	// No cursor, just apply base style
	return baseStyle.Render(visibleValue)
}

// padRight pads a string to the specified width with spaces
func padRight(str string, width int) string {
	if len(str) >= width {
		return str[:width]
	}
	return str + strings.Repeat(" ", width-len(str))
}

// Clear clears the input value
func (t *TextInput) Clear() {
	t.value = ""
	t.cursor = 0
}

// MoveCursorToEnd moves the cursor to the end of the input
func (t *TextInput) MoveCursorToEnd() {
	t.cursor = len(t.value)
}

// MoveCursorToStart moves the cursor to the start of the input
func (t *TextInput) MoveCursorToStart() {
	t.cursor = 0
}

// SetCursor sets the cursor position
func (t *TextInput) SetCursor(pos int) {
	if pos < 0 {
		pos = 0
	}
	if pos > len(t.value) {
		pos = len(t.value)
	}
	t.cursor = pos
}