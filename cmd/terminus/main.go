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

	log.SetFlags(0) // Disable timestamping for log messages
	log.SetOutput(os.Stderr) // Log to stderr
	
	fmt.Println("Terminus CLI client")

	// Define WebSocket server URL
	u := url.URL{Scheme: "ws", Host: *serverAddr, Path: "/ws"}
	fmt.Printf("Connecting to %s\n", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer c.Close()

	fmt.Println("Connected to WebSocket server.")

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
				// Wrap raw input in JSON protocol
				var msg ClientMessage
				char := buf[:n]
				
				// Basic mapping for control characters
				switch {
				case char[0] == '\r': // Enter
					msg = ClientMessage{
						Type: "key",
						Data: map[string]interface{}{"keyType": "enter"},
					}
				case char[0] == '\x7f': // Backspace
					msg = ClientMessage{
						Type: "key",
						Data: map[string]interface{}{"keyType": "backspace"},
					}
				case char[0] == '\t': // Tab
					msg = ClientMessage{
						Type: "key",
						Data: map[string]interface{}{"keyType": "tab"},
					}
				case char[0] == '\x1b': // Escape
					msg = ClientMessage{
						Type: "key",
						Data: map[string]interface{}{"keyType": "escape"},
					}
				case char[0] == '\x03': // Ctrl+C
					msg = ClientMessage{
						Type: "key",
						Data: map[string]interface{}{"keyType": "ctrl+c"},
					}
				case char[0] == ' ': // Space
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
		defer close(done)
		for {
			messageType, message, err := c.ReadMessage()
			if err != nil {
				log.Printf("Error reading from WebSocket: %v", err)
				return
			}
			
			if messageType == websocket.TextMessage {
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
