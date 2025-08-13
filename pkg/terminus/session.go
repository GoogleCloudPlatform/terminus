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
	mu       sync.RWMutex
	closed   bool
	closeOnce sync.Once
	width    int
	height   int
}

// NewSession creates a new session
func NewSession(id string, conn *websocket.Conn, component Component) *Session {
	s := &Session{
		id:           id,
		conn:         conn,
		component:    component,
		incoming:     make(chan []byte, 100),
		outgoing:     make(chan []byte, 100),
		width:        80,  // Default dimensions
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
	width := s.width
	height := s.height
	s.mu.RUnlock()
	
	// Ensure screen differ has correct dimensions
	s.screenDiffer.Resize(width, height)
	
	// Compute diff operations
	ops := s.screenDiffer.Update(view)
	
	// Convert diff ops to render commands
	for _, op := range ops {
		var msg ServerMessage
		
		switch op.Type {
		case DiffOpClear:
			msg = ServerMessage{
				Type: "clear",
				Data: map[string]interface{}{},
			}
			
		case DiffOpUpdateLine:
			lineOp := op.Data.(UpdateLineOp)
			msg = ServerMessage{
				Type: "updateLine",
				Data: map[string]interface{}{
					"y":       lineOp.Y,
					"content": lineOp.Content,
				},
			}
			
		case DiffOpSetCell:
			cellOp := op.Data.(SetCellOp)
			msg = ServerMessage{
				Type: "setCell",
				Data: map[string]interface{}{
					"x":     cellOp.X,
					"y":     cellOp.Y,
					"rune":  cellOp.Rune,
					"style": cellOp.Style,
				},
			}
			
		default:
			continue
		}
		
		data, err := json.Marshal(msg)
		if err != nil {
			fmt.Printf("Failed to marshal render message for session %s: %v\n", s.id, err)
			continue
		}
		
		select {
		case s.outgoing <- data:
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
					return KeyMsg{Type: KeyRunes, Runes: runes}
				}
			case "enter":
				return KeyMsg{Type: KeyEnter}
			case "space":
				return KeyMsg{Type: KeySpace}
			case "backspace":
				return KeyMsg{Type: KeyBackspace}
			case "tab":
				return KeyMsg{Type: KeyTab}
			case "escape":
				return KeyMsg{Type: KeyEsc}
			case "up":
				return KeyMsg{Type: KeyUp}
			case "down":
				return KeyMsg{Type: KeyDown}
			case "left":
				return KeyMsg{Type: KeyLeft}
			case "right":
				return KeyMsg{Type: KeyRight}
			case "ctrl+c":
				return KeyMsg{Type: KeyCtrlC}
			}
		}
		
	case "resize":
		if resizeData, ok := msg.Data.(map[string]interface{}); ok {
			width, _ := resizeData["width"].(float64)
			height, _ := resizeData["height"].(float64)
			
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

// ServerMessage represents a message to the client
type ServerMessage struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}