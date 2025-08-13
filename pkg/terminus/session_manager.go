package terminus

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// SessionManager manages active sessions
type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

// CreateSession creates a new session
func (sm *SessionManager) CreateSession(conn *websocket.Conn, component Component) *Session {
	id := uuid.New().String()
	session := NewSession(id, conn, component)
	
	sm.mu.Lock()
	sm.sessions[id] = session
	sm.mu.Unlock()
	
	fmt.Printf("Session created: %s\n", id)
	return session
}

// RemoveSession removes a session
func (sm *SessionManager) RemoveSession(id string) {
	sm.mu.Lock()
	delete(sm.sessions, id)
	sm.mu.Unlock()
	
	fmt.Printf("Session removed: %s\n", id)
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(id string) *Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.sessions[id]
}

// CloseAll closes all sessions
func (sm *SessionManager) CloseAll() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	for id, session := range sm.sessions {
		session.Close()
		delete(sm.sessions, id)
	}
}

// Count returns the number of active sessions
func (sm *SessionManager) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}