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
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Session represents a single connected client
type Session struct {
	id        string
	conn      *websocket.Conn
	component Component
	engine    *Engine

	// Message channels
	incoming chan []byte
	outgoing chan []byte

	// Rendering
	screenDiffer *ScreenDiffer

	// State
	mu        sync.RWMutex
	closed    bool
	closeOnce sync.Once
	width     int
	height    int
}

// NewSession creates a new session
func NewSession(id string, conn *websocket.Conn, component Component) *Session {
	s := &Session{
		id:           id,
		conn:         conn,
		component:    component,
		incoming:     make(chan []byte, 100),
		outgoing:     make(chan []byte, 100),
		width:        80, // Default dimensions
		height:       24,
		screenDiffer: NewScreenDiffer(80, 24),
	}

	// Create engine with callbacks
	s.engine = NewEngine(component)
	s.engine.SetRenderCallback(s.handleRender)
	s.engine.SetQuitCallback(s.handleQuit)

	return s
}

// ID returns the session ID
func (s *Session) ID() string {
	return s.id
}

// Run starts the session
func (s *Session) Run(ctx context.Context) {
	defer s.Close()

	// Start engine
	if err := s.engine.Start(); err != nil {
		fmt.Printf("Failed to start engine for session %s: %v\n", s.id, err)
		return
	}
	defer s.engine.Stop()

	// Start goroutines
	var wg sync.WaitGroup

	// WebSocket reader
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.readPump()
	}()

	// WebSocket writer
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.writePump(ctx)
	}()

	// Message processor
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.processMessages(ctx)
	}()

	// Wait for context cancellation or session close
	<-ctx.Done()
	s.Close()
	wg.Wait()
}

// Close closes the session
func (s *Session) Close() {
	s.closeOnce.Do(func() {
		// Stop engine first so it won't render into closed channels.
		if s.engine != nil {
			s.engine.Stop()
		}

		s.mu.Lock()
		s.closed = true
		s.mu.Unlock()

		close(s.incoming)
		close(s.outgoing)
		if s.conn != nil {
			s.conn.Close()
		}
	})
}

// readPump reads messages from the WebSocket connection
func (s *Session) readPump() {
	defer s.Close()

	s.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	s.conn.SetPongHandler(func(string) error {
		s.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := s.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("WebSocket error for session %s: %v\n", s.id, err)
			}
			break
		}

		s.mu.RLock()
		closed := s.closed
		s.mu.RUnlock()

		if closed {
			break
		}

		select {
		case s.incoming <- message:
		default:
			fmt.Printf("Incoming message buffer full for session %s\n", s.id)
		}
	}
}

// writePump writes messages to the WebSocket connection
func (s *Session) writePump(ctx context.Context) {
	ticker := time.NewTicker(54 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-s.outgoing:
			s.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				s.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := s.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			s.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := s.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

// processMessages processes incoming messages
func (s *Session) processMessages(ctx context.Context) {
	for {
		select {
		case message, ok := <-s.incoming:
			if !ok {
				return
			}

			// Parse message
			var msg ClientMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				fmt.Printf("Failed to parse message from session %s: %v\n", s.id, err)
				continue
			}

			// Convert to terminus message
			terminusMsg := s.clientToTerminusMessage(msg)
			if terminusMsg != nil {
				s.engine.SendMessage(terminusMsg)
			}

		case <-ctx.Done():
			return
		}
	}
}

// handleRender is called when the engine renders a new view
func (s *Session) handleRender(view string) {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return
	}
	width := s.width
	height := s.height
	s.mu.RUnlock()

	// Ensure screen differ has correct dimensions
	s.screenDiffer.Resize(width, height)

	// Compute diff operations (ANSI string)
	ansiOutput := s.screenDiffer.Update(view)

	// Send raw ANSI string to client
	if ansiOutput != "" {
		select {
		case s.outgoing <- []byte(ansiOutput):
		default:
			fmt.Printf("Outgoing message buffer full for session %s\n", s.id)
		}
	}
}

// handleQuit is called when the engine quits
func (s *Session) handleQuit() {
	s.Close()
}

// clientToTerminusMessage converts client messages to terminus messages
func (s *Session) clientToTerminusMessage(msg ClientMessage) Msg {
	switch msg.Type {
	case "key":
		if keyData, ok := msg.Data.(map[string]interface{}); ok {
			keyType, _ := keyData["keyType"].(string)

			// Parse modifiers
			var alt, ctrl, shift bool
			if mods, ok := keyData["modifiers"].(map[string]interface{}); ok {
				alt, _ = mods["alt"].(bool)
				ctrl, _ = mods["ctrl"].(bool)
				shift, _ = mods["shift"].(bool)
			}

			// Helper to create KeyMsg with modifiers
			newKeyMsg := func(t KeyType, runes ...rune) KeyMsg {
				return KeyMsg{
					Type:  t,
					Runes: runes,
					Alt:   alt,
					Ctrl:  ctrl,
					Shift: shift,
				}
			}

			// Handle different key types
			switch keyType {
			case "runes":
				if runesData, ok := keyData["runes"].([]interface{}); ok {
					runes := make([]rune, 0, len(runesData))
					for _, r := range runesData {
						if str, ok := r.(string); ok && len(str) > 0 {
							// Only take the first character from each string
							// Client sends individual characters as separate strings
							runes = append(runes, []rune(str)[0])
						}
					}
					return newKeyMsg(KeyRunes, runes...)
				}
			case "enter":
				return newKeyMsg(KeyEnter)
			case "space":
				return newKeyMsg(KeySpace)
			case "backspace":
				return newKeyMsg(KeyBackspace)
			case "tab":
				return newKeyMsg(KeyTab)
			case "escape":
				return newKeyMsg(KeyEsc)
			case "up":
				return newKeyMsg(KeyUp)
			case "down":
				return newKeyMsg(KeyDown)
			case "left":
				return newKeyMsg(KeyLeft)
			case "right":
				return newKeyMsg(KeyRight)
			case "ctrl+c":
				return newKeyMsg(KeyCtrlC)
			case "ctrl+r":
				return newKeyMsg(KeyCtrlR)
			case "ctrl+s":
				return newKeyMsg(KeyCtrlS)
			}
		}

	case "resize":
		if resizeData, ok := msg.Data.(map[string]interface{}); ok {
			width, widthOk := resizeData["width"].(float64)
			height, heightOk := resizeData["height"].(float64)

			if !widthOk || !heightOk || width <= 0 || height <= 0 {
				fmt.Printf("Invalid resize dimensions received from session %s: width=%.0f, height=%.0f\n", s.id, width, height)
				return nil
			}

			// Update session dimensions
			s.mu.Lock()
			s.width = int(width)
			s.height = int(height)
			s.mu.Unlock()

			// Update screen differ
			s.screenDiffer.Resize(int(width), int(height))

			return WindowSizeMsg{
				Width:  int(width),
				Height: int(height),
			}
		}
	}

	return nil
}

// ClientMessage represents a message from the client
type ClientMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}
