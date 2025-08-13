package terminus

import (
	"encoding/json"
	"testing"
)

func TestClientToTerminusMessage(t *testing.T) {
	session := &Session{}
	
	tests := []struct {
		name     string
		input    ClientMessage
		expected Msg
	}{
		{
			name: "Character key",
			input: ClientMessage{
				Type: "key",
				Data: map[string]interface{}{
					"keyType": "runes",
					"runes":   []interface{}{"a"},
				},
			},
			expected: KeyMsg{Type: KeyRunes, Runes: []rune{'a'}},
		},
		{
			name: "Multiple characters",
			input: ClientMessage{
				Type: "key",
				Data: map[string]interface{}{
					"keyType": "runes",
					"runes":   []interface{}{"h", "e", "l", "l", "o"},
				},
			},
			expected: KeyMsg{Type: KeyRunes, Runes: []rune{'h', 'e', 'l', 'l', 'o'}},
		},
		{
			name: "Enter key",
			input: ClientMessage{
				Type: "key",
				Data: map[string]interface{}{
					"keyType": "enter",
				},
			},
			expected: KeyMsg{Type: KeyEnter},
		},
		{
			name: "Space key",
			input: ClientMessage{
				Type: "key",
				Data: map[string]interface{}{
					"keyType": "space",
				},
			},
			expected: KeyMsg{Type: KeySpace},
		},
		{
			name: "Backspace key",
			input: ClientMessage{
				Type: "key",
				Data: map[string]interface{}{
					"keyType": "backspace",
				},
			},
			expected: KeyMsg{Type: KeyBackspace},
		},
		{
			name: "Tab key",
			input: ClientMessage{
				Type: "key",
				Data: map[string]interface{}{
					"keyType": "tab",
				},
			},
			expected: KeyMsg{Type: KeyTab},
		},
		{
			name: "Escape key",
			input: ClientMessage{
				Type: "key",
				Data: map[string]interface{}{
					"keyType": "escape",
				},
			},
			expected: KeyMsg{Type: KeyEsc},
		},
		{
			name: "Arrow up",
			input: ClientMessage{
				Type: "key",
				Data: map[string]interface{}{
					"keyType": "up",
				},
			},
			expected: KeyMsg{Type: KeyUp},
		},
		{
			name: "Arrow down",
			input: ClientMessage{
				Type: "key",
				Data: map[string]interface{}{
					"keyType": "down",
				},
			},
			expected: KeyMsg{Type: KeyDown},
		},
		{
			name: "Arrow left",
			input: ClientMessage{
				Type: "key",
				Data: map[string]interface{}{
					"keyType": "left",
				},
			},
			expected: KeyMsg{Type: KeyLeft},
		},
		{
			name: "Arrow right",
			input: ClientMessage{
				Type: "key",
				Data: map[string]interface{}{
					"keyType": "right",
				},
			},
			expected: KeyMsg{Type: KeyRight},
		},
		{
			name: "Ctrl+C",
			input: ClientMessage{
				Type: "key",
				Data: map[string]interface{}{
					"keyType": "ctrl+c",
				},
			},
			expected: KeyMsg{Type: KeyCtrlC},
		},
		{
			name: "Window resize",
			input: ClientMessage{
				Type: "resize",
				Data: map[string]interface{}{
					"width":  80.0,
					"height": 24.0,
				},
			},
			expected: WindowSizeMsg{Width: 80, Height: 24},
		},
		{
			name: "Unknown message type",
			input: ClientMessage{
				Type: "unknown",
				Data: nil,
			},
			expected: nil,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := session.clientToTerminusMessage(tt.input)
			
			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %+v", result)
				}
				return
			}
			
			switch expected := tt.expected.(type) {
			case KeyMsg:
				keyMsg, ok := result.(KeyMsg)
				if !ok {
					t.Fatalf("Expected KeyMsg, got %T", result)
				}
				
				if keyMsg.Type != expected.Type {
					t.Errorf("Expected key type %v, got %v", expected.Type, keyMsg.Type)
				}
				
				if len(keyMsg.Runes) != len(expected.Runes) {
					t.Errorf("Expected %d runes, got %d", len(expected.Runes), len(keyMsg.Runes))
				} else {
					for i, r := range expected.Runes {
						if keyMsg.Runes[i] != r {
							t.Errorf("Expected rune %c at index %d, got %c", r, i, keyMsg.Runes[i])
						}
					}
				}
				
			case WindowSizeMsg:
				sizeMsg, ok := result.(WindowSizeMsg)
				if !ok {
					t.Fatalf("Expected WindowSizeMsg, got %T", result)
				}
				
				if sizeMsg.Width != expected.Width {
					t.Errorf("Expected width %d, got %d", expected.Width, sizeMsg.Width)
				}
				
				if sizeMsg.Height != expected.Height {
					t.Errorf("Expected height %d, got %d", expected.Height, sizeMsg.Height)
				}
			}
		})
	}
}

func TestServerMessage(t *testing.T) {
	tests := []struct {
		name     string
		message  ServerMessage
		expected string
	}{
		{
			name: "Render message",
			message: ServerMessage{
				Type: "render",
				Data: map[string]interface{}{
					"content": "Hello, World!",
				},
			},
			expected: `{"type":"render","data":{"content":"Hello, World!"}}`,
		},
		{
			name: "Clear message",
			message: ServerMessage{
				Type: "clear",
				Data: map[string]interface{}{},
			},
			expected: `{"type":"clear","data":{}}`,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.message)
			if err != nil {
				t.Fatalf("Failed to marshal message: %v", err)
			}
			
			// Parse both to compare as JSON objects (handles key ordering)
			var expected, actual map[string]interface{}
			
			if err := json.Unmarshal([]byte(tt.expected), &expected); err != nil {
				t.Fatalf("Failed to unmarshal expected: %v", err)
			}
			
			if err := json.Unmarshal(data, &actual); err != nil {
				t.Fatalf("Failed to unmarshal actual: %v", err)
			}
			
			// Compare types
			if expected["type"] != actual["type"] {
				t.Errorf("Expected type %v, got %v", expected["type"], actual["type"])
			}
		})
	}
}