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

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/term"
)

func main() {
	serverAddr := flag.String("addr", "localhost:8890", "Terminus server address")
	flag.Parse()

	// Setup file logging for debug
	f, err := os.OpenFile("terminus-cli.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	log.Println("--- Terminus CLI Client Started ---")
	
	fmt.Println("Terminus CLI client")

	// Put terminal into raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatalf("Failed to put terminal into raw mode: %v", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// Define WebSocket server URL
	u := url.URL{Scheme: "ws", Host: *serverAddr, Path: "/ws"}
	log.Printf("Connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		// We can't log to stderr easily in raw mode without messing up screen, 
		// but Restore is deferred. We'll rely on the log file.
		log.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer c.Close()

	log.Println("Connected to WebSocket server.")

	done := make(chan struct{}) // Channel to signal when to exit
	
	// Channel for input data
	inputChan := make(chan []byte)
	
	// Goroutine to read stdin
	go func() {
		defer close(inputChan)
		buf := make([]byte, 1024)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				if err == io.EOF {
					log.Println("Stdin closed.")
				} else {
					log.Printf("Error reading stdin: %v", err)
				}
				return
			}
			if n > 0 {
				// Copy data to avoid race conditions with buffer reuse
				chunk := make([]byte, n)
				copy(chunk, buf[:n])
				inputChan <- chunk
			}
		}
	}()

	// Goroutine to parse input and send messages
	go func() {
		defer close(done)
		
		const (
			StateNormal = iota
			StateEsc
			StateCSI
		)
		
		state := StateNormal
		var buffer []byte
		
		// Timer for escape sequence timeout
		escapeTimer := time.NewTimer(50 * time.Millisecond)
		if !escapeTimer.Stop() {
			select {
			case <-escapeTimer.C:
			default:
			}
		}
		
		// Helper to send a message
		sendMsg := func(msg ClientMessage) {
			jsonBytes, err := json.Marshal(msg)
			if err != nil {
				log.Printf("Error marshaling key message: %v", err)
				return
			}
			c.WriteMessage(websocket.TextMessage, jsonBytes)
		}
		
		// Process a completed CSI sequence
		handleCSI := func(seq []byte) {
			if len(seq) < 3 { return } // Should be at least ESC [ X
			
			final := seq[len(seq)-1]
			switch final {
			case 'Z': // Shift+Tab: ESC [ Z
				log.Println("Detected Shift+Tab")
				sendMsg(ClientMessage{
					Type: "key",
					Data: map[string]interface{}{
						"keyType": "tab",
						"modifiers": map[string]bool{"shift": true},
					},
				})
			case 'A': // Up
				sendMsg(ClientMessage{Type: "key", Data: map[string]interface{}{"keyType": "up"}})
			case 'B': // Down
				sendMsg(ClientMessage{Type: "key", Data: map[string]interface{}{"keyType": "down"}})
			case 'C': // Right
				sendMsg(ClientMessage{Type: "key", Data: map[string]interface{}{"keyType": "right"}})
			case 'D': // Left
				sendMsg(ClientMessage{Type: "key", Data: map[string]interface{}{"keyType": "left"}})
			default:
				// Unknown sequence, treat as runes? Or ignore?
				log.Printf("Unknown CSI sequence: %q", seq)
			}
		}
		
		// Helper to flush buffer as raw runes
		flushBuffer := func() {
			for _, b := range buffer {
				// If it was an Escape that timed out
				if b == '\x1b' {
					sendMsg(ClientMessage{Type: "key", Data: map[string]interface{}{"keyType": "escape"}})
				} else {
					// Treat as rune
					sendMsg(ClientMessage{
						Type: "key",
						Data: map[string]interface{}{
							"keyType": "runes",
							"runes":   []string{string(b)},
						},
					})
				}
			}
			buffer = nil
			state = StateNormal
		}

		for {
			select {
			case chunk, ok := <-inputChan:
				if !ok {
					return
				}
				
				log.Printf("Read %d bytes: %v", len(chunk), chunk)
				
				for _, b := range chunk {
					switch state {
					case StateNormal:
						if b == '\x1b' {
							state = StateEsc
							buffer = append(buffer, b)
							escapeTimer.Reset(50 * time.Millisecond)
						} else {
							// Handle Normal Key
							switch b {
							case '\r', '\n':
								sendMsg(ClientMessage{Type: "key", Data: map[string]interface{}{"keyType": "enter"}})
							case '\x7f', '\b':
								sendMsg(ClientMessage{Type: "key", Data: map[string]interface{}{"keyType": "backspace"}})
							case '\t':
								sendMsg(ClientMessage{Type: "key", Data: map[string]interface{}{"keyType": "tab"}})
							case '\x03': // Ctrl+C
								log.Println("Detected Ctrl+C")
								sendMsg(ClientMessage{
									Type: "key", 
									Data: map[string]interface{}{"keyType": "ctrl+c", "modifiers": map[string]bool{"ctrl": true}},
								})
								return // Exit
							case '\x12': // Ctrl+R
								log.Println("Detected Ctrl+R")
								sendMsg(ClientMessage{
									Type: "key",
									Data: map[string]interface{}{"keyType": "ctrl+r", "modifiers": map[string]bool{"ctrl": true}},
								})
							case '\x13': // Ctrl+S
								log.Println("Detected Ctrl+S")
								sendMsg(ClientMessage{
									Type: "key",
									Data: map[string]interface{}{"keyType": "ctrl+s", "modifiers": map[string]bool{"ctrl": true}},
								})
							case ' ':
								sendMsg(ClientMessage{Type: "key", Data: map[string]interface{}{"keyType": "space"}})
							default:
								sendMsg(ClientMessage{
									Type: "key",
									Data: map[string]interface{}{
										"keyType": "runes",
										"runes":   []string{string(b)},
									},
								})
							}
						}
						
					case StateEsc:
						buffer = append(buffer, b)
						if b == '[' {
							state = StateCSI
						} else {
							// Not CSI (e.g. Alt+Key or SS3 \x1bO)
							// For now, just flush (treat as Esc + Key)
							// Ideally we handle \x1bO for F-keys here too
							flushBuffer()
							if !escapeTimer.Stop() {
								select { case <-escapeTimer.C: default: }
							}
						}
						
					case StateCSI:
						buffer = append(buffer, b)
						// Check for final byte (0x40-0x7E)
						if b >= 0x40 && b <= 0x7E {
							handleCSI(buffer)
							buffer = nil
							state = StateNormal
							if !escapeTimer.Stop() {
								select { case <-escapeTimer.C: default: }
							}
						}
					}
				}
				
			case <-escapeTimer.C:
				// Timeout waiting for sequence completion
				log.Println("Escape sequence timeout, flushing buffer")
				flushBuffer()
			}
		}
	}()

	// Goroutine to read from WebSocket and write to stdout
	go func() {
		// We don't close 'done' here because the server closing 
		// shouldn't necessarily kill the client input loop immediately, 
		// but effectively it ends the session.
		for {
			messageType, message, err := c.ReadMessage()
			if err != nil {
				log.Printf("Error reading from WebSocket: %v", err)
				close(done) // Signal exit
				return
			}
			
			if messageType == websocket.TextMessage {
				// Direct pass-through of ANSI to stdout
				os.Stdout.Write(message)
			}
		}
	}()

	// Handle interrupt signals for graceful shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Handle SIGWINCH for terminal resizing
	sigWinCh := make(chan os.Signal, 1)
	signal.Notify(sigWinCh, syscall.SIGWINCH)

	// Initial resize message to server
	go sendResizeMessage(c)

	// Goroutine to handle resize events
	go func() {
		var resizeTimer *time.Timer
		for range sigWinCh {
			if resizeTimer != nil {
				resizeTimer.Stop()
			}
			resizeTimer = time.AfterFunc(100*time.Millisecond, func() {
				sendResizeMessage(c)
			})
		}
	}()

	select {
	case <-done:
		log.Println("Goroutines finished, exiting.")
	case <-interrupt:
		log.Println("Interrupt signal received, closing WebSocket...")
		// Cleanly close the WebSocket connection by sending a close message
		err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Printf("Error sending close message: %v", err)
			return
		}
		// Wait for the server to close the connection, or timeout
		select {
		case <-done:
		case <-time.After(time.Second):
		}
	}
}

// ClientMessage represents a message from the client (copied from pkg/terminus/session.go for CLI use)
type ClientMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func sendResizeMessage(conn *websocket.Conn) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Printf("Error getting terminal size: %v", err)
		return
	}

	resizeMsg := ClientMessage{
		Type: "resize",
		Data: map[string]interface{}{
			"width":  width,
			"height": height,
		},
	}

	jsonMsg, err := json.Marshal(resizeMsg)
	if err != nil {
		log.Printf("Error marshalling resize message: %v", err)
		return
	}

	err = conn.WriteMessage(websocket.TextMessage, jsonMsg)
	if err != nil {
		log.Printf("Error sending resize message: %v", err)
	}
}
