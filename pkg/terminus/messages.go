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

// KeyType represents different types of keyboard input
type KeyType int

const (
	// KeyRunes represents character input
	KeyRunes KeyType = iota
	// KeyEnter represents the Enter key
	KeyEnter
	// KeySpace represents the Space key
	KeySpace
	// KeyBackspace represents the Backspace key
	KeyBackspace
	// KeyDelete represents the Delete key
	KeyDelete
	// KeyTab represents the Tab key
	KeyTab
	// KeyEsc represents the Escape key
	KeyEsc
	// KeyUp represents the Up arrow key
	KeyUp
	// KeyDown represents the Down arrow key
	KeyDown
	// KeyLeft represents the Left arrow key
	KeyLeft
	// KeyRight represents the Right arrow key
	KeyRight
	// KeyHome represents the Home key
	KeyHome
	// KeyEnd represents the End key
	KeyEnd
	// KeyPgUp represents the Page Up key
	KeyPgUp
	// KeyPgDown represents the Page Down key
	KeyPgDown
	// KeyF1 through KeyF12 represent function keys
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
	// KeyCtrlC represents Ctrl+C
	KeyCtrlC
	// KeyCtrlD represents Ctrl+D
	KeyCtrlD
	// KeyCtrlR represents Ctrl+R
	KeyCtrlR
	// KeyCtrlS represents Ctrl+S
	KeyCtrlS
	// KeyCtrlZ represents Ctrl+Z
	KeyCtrlZ
)

// KeyMsg represents a keyboard input message
type KeyMsg struct {
	Type  KeyType
	Runes []rune // For KeyRunes type
	Alt   bool   // Alt modifier
	Ctrl  bool   // Ctrl modifier
	Shift bool   // Shift modifier
}

// String returns a human-readable representation of the key message
func (k KeyMsg) String() string {
	switch k.Type {
	case KeyRunes:
		if len(k.Runes) > 0 {
			return string(k.Runes)
		}
		return ""
	case KeyEnter:
		return "enter"
	case KeySpace:
		return "space"
	case KeyBackspace:
		return "backspace"
	case KeyDelete:
		return "delete"
	case KeyTab:
		return "tab"
	case KeyEsc:
		return "esc"
	case KeyUp:
		return "up"
	case KeyDown:
		return "down"
	case KeyLeft:
		return "left"
	case KeyRight:
		return "right"
	case KeyHome:
		return "home"
	case KeyEnd:
		return "end"
	case KeyPgUp:
		return "pgup"
	case KeyPgDown:
		return "pgdown"
	case KeyF1:
		return "f1"
	case KeyF2:
		return "f2"
	case KeyF3:
		return "f3"
	case KeyF4:
		return "f4"
	case KeyF5:
		return "f5"
	case KeyF6:
		return "f6"
	case KeyF7:
		return "f7"
	case KeyF8:
		return "f8"
	case KeyF9:
		return "f9"
	case KeyF10:
		return "f10"
	case KeyF11:
		return "f11"
	case KeyF12:
		return "f12"
	case KeyCtrlC:
		return "ctrl+c"
	case KeyCtrlD:
		return "ctrl+d"
	case KeyCtrlR:
		return "ctrl+r"
	case KeyCtrlS:
		return "ctrl+s"
	case KeyCtrlZ:
		return "ctrl+z"
	default:
		return "unknown"
	}
}

// QuitMsg is a message type for signaling application quit
type QuitMsg struct{}

// WindowSizeMsg is sent when the terminal window is resized
type WindowSizeMsg struct {
	Width  int
	Height int
}