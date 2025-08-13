package terminus

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Program is the main entry point for a TerminusGo application
type Program struct {
	// Configuration
	addr                   string
	rootComponentFactory   func() Component
	staticFS               embed.FS
	staticPath             string
	
	// Runtime state
	server         *http.Server
	sessionManager *SessionManager
	upgrader       websocket.Upgrader
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
}

// ProgramOption is a function that configures a Program
type ProgramOption func(*Program)

// WithStaticFiles configures the program to serve static files from an embedded filesystem
func WithStaticFiles(fs embed.FS, path string) ProgramOption {
	return func(p *Program) {
		p.staticFS = fs
		p.staticPath = path
	}
}

// WithAddress configures the server address
func WithAddress(addr string) ProgramOption {
	return func(p *Program) {
		p.addr = addr
	}
}

// NewProgram creates a new TerminusGo program
func NewProgram(rootComponentFactory func() Component, opts ...ProgramOption) *Program {
	ctx, cancel := context.WithCancel(context.Background())
	
	p := &Program{
		addr:                 ":8080",
		rootComponentFactory: rootComponentFactory,
		sessionManager:       NewSessionManager(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// In production, implement proper origin checking
				return true
			},
		},
		ctx:    ctx,
		cancel: cancel,
	}
	
	// Apply options
	for _, opt := range opts {
		opt(p)
	}
	
	return p
}

// Start starts the TerminusGo program
func (p *Program) Start() error {
	mux := http.NewServeMux()
	
	// Serve static files if configured
	if p.staticPath != "" {
		// Create a sub-filesystem from the static path
		subFS, err := fs.Sub(p.staticFS, p.staticPath)
		if err != nil {
			return fmt.Errorf("failed to create sub filesystem: %w", err)
		}
		fileServer := http.FileServer(http.FS(subFS))
		mux.Handle("/", fileServer)
	} else {
		// Serve default HTML if no static files configured
		mux.HandleFunc("/", p.handleIndex)
	}
	
	// WebSocket endpoint
	mux.HandleFunc("/ws", p.handleWebSocket)
	
	p.server = &http.Server{
		Addr:    p.addr,
		Handler: mux,
	}
	
	// Start server in goroutine
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()
	
	return nil
}

// Stop gracefully shuts down the program
func (p *Program) Stop() error {
	p.cancel()
	
	// Shutdown HTTP server
	if p.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5)
		defer cancel()
		if err := p.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}
	}
	
	// Close all sessions
	p.sessionManager.CloseAll()
	
	// Wait for all goroutines
	p.wg.Wait()
	
	return nil
}

// Wait blocks until the program is stopped
func (p *Program) Wait() {
	p.wg.Wait()
}

// handleIndex serves the default HTML page
func (p *Program) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, defaultHTML)
}

// handleWebSocket upgrades HTTP connections to WebSocket
func (p *Program) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := p.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("WebSocket upgrade failed: %v\n", err)
		return
	}
	
	// Create new session
	session := p.sessionManager.CreateSession(conn, p.rootComponentFactory())
	
	// Start session
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		session.Run(p.ctx)
		p.sessionManager.RemoveSession(session.ID())
	}()
}

// defaultHTML is the minimal HTML served when no static files are configured
const defaultHTML = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>TerminusGo</title>
    <style>
        body {
            margin: 0;
            padding: 0;
            background: #1e1e1e;
            color: #d4d4d4;
            font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
            font-size: 14px;
            line-height: 1.5;
        }
        #terminal {
            padding: 20px;
            white-space: pre-wrap;
            word-wrap: break-word;
        }
    </style>
</head>
<body>
    <div id="terminal">Connecting...</div>
    <script src="/terminus-client.js"></script>
</body>
</html>`