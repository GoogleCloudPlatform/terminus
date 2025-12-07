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
	
	// Goroutine to read stdin and send to WebSocket
	go func() {
		defer close(done)
		buf := make([]byte, 1)
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
				char := buf[0]
				log.Printf("Read byte: %d (%q)", char, char)

				// Wrap raw input in JSON protocol
				var msg ClientMessage
				
				// Basic mapping for control characters
				switch {
				case char == '\r' || char == '\n': // Enter (Handle both CR and LF)
					log.Println("Detected Enter")
					msg = ClientMessage{
						Type: "key",
						Data: map[string]interface{}{"keyType": "enter"},
					}
				case char == '\x7f' || char == '\b': // Backspace (Handle DEL and BS)
					msg = ClientMessage{
						Type: "key",
						Data: map[string]interface{}{"keyType": "backspace"},
					}
				case char == '\t': // Tab
					msg = ClientMessage{
						Type: "key",
						Data: map[string]interface{}{"keyType": "tab"},
					}
				case char == '\x1b': // Escape
					msg = ClientMessage{
						Type: "key",
						Data: map[string]interface{}{"keyType": "escape"},
					}
				case char == '\x03': // Ctrl+C
					log.Println("Detected Ctrl+C, exiting...")
					msg = ClientMessage{
						Type: "key",
						Data: map[string]interface{}{"keyType": "ctrl+c"},
					}
					// Send the message then exit
					jsonBytes, _ := json.Marshal(msg)
					c.WriteMessage(websocket.TextMessage, jsonBytes)
					return 
				case char == ' ': // Space
					msg = ClientMessage{
						Type: "key",
						Data: map[string]interface{}{"keyType": "space"},
					}
				default:
					// Treat as regular rune(s)
					msg = ClientMessage{
						Type: "key",
						Data: map[string]interface{}{
							"keyType": "runes",
							"runes":   []string{string(char)},
						},
					}
				}

				jsonBytes, err := json.Marshal(msg)
				if err != nil {
					log.Printf("Error marshaling key message: %v", err)
					continue
				}

				err = c.WriteMessage(websocket.TextMessage, jsonBytes)
				if err != nil {
					log.Printf("Error writing to WebSocket: %v", err)
					return
				}
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
