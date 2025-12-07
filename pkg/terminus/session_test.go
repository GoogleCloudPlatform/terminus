package terminus

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestClientToTerminusMessage(t *testing.T) {
	// Setup a dummy Program to handle WebSocket connections
	factory := func() Component { return &mockProgramComponent{} }
	program := NewProgram(factory)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		program.handleWebSocket(w, r)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Establish a WebSocket connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Get the session created by the program
	// This is a bit of a hack for testing, normally you wouldn't reach into the manager directly
	var testSession *Session
	// Wait for the session to be created
	for i := 0; i < 10; i++ {
		if program.sessionManager.Count() > 0 {
			for _, sess := range program.sessionManager.sessions { // Assuming sessions map is accessible
				testSession = sess
				break
			}
		}
		if testSession != nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if testSession == nil {
		t.Fatalf("Failed to retrieve test session")
	}

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
				Data: map[string]interface{}{"keyType": "enter"},
			},
			expected: KeyMsg{Type: KeyEnter},
		},
		{
			name: "Space key",
			input: ClientMessage{
				Type: "key",
				Data: map[string]interface{}{"keyType": "space"},
			},
			expected: KeyMsg{Type: KeySpace},
		},
		{
			name: "Backspace key",
			input: ClientMessage{
				Type: "key",
				Data: map[string]interface{}{"keyType": "backspace"},
			},
			expected: KeyMsg{Type: KeyBackspace},
		},
		{
			name: "Tab key",
			input: ClientMessage{
				Type: "key",
				Data: map[string]interface{}{"keyType": "tab"},
			},
			expected: KeyMsg{Type: KeyTab},
		},
		{
			name: "Escape key",
			input: ClientMessage{
				Type: "key",
				Data: map[string]interface{}{"keyType": "escape"},
			},
			expected: KeyMsg{Type: KeyEsc},
		},
		{
			name: "Arrow up",
			input: ClientMessage{
				Type: "key",
				Data: map[string]interface{}{"keyType": "up"},
			},
			expected: KeyMsg{Type: KeyUp},
		},
		{
			name: "Arrow down",
			input: ClientMessage{
				Type: "key",
				Data: map[string]interface{}{"keyType": "down"},
			},
			expected: KeyMsg{Type: KeyDown},
		},
		{
			name: "Arrow left",
			input: ClientMessage{
				Type: "key",
				Data: map[string]interface{}{"keyType": "left"},
			},
			expected: KeyMsg{Type: KeyLeft},
		},
		{
			name: "Arrow right",
			input: ClientMessage{
				Type: "key",
				Data: map[string]interface{}{"keyType": "right"},
			},
			expected: KeyMsg{Type: KeyRight},
		},
		{
			name: "Ctrl+C",
			input: ClientMessage{
				Type: "key",
				Data: map[string]interface{}{"keyType": "ctrl+c"},
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
			result := testSession.clientToTerminusMessage(tt.input) // Use testSession here

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
