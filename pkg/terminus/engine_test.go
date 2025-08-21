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
	"sync"
	"testing"
	"time"
)

// testComponent is a simple component for testing
type testComponent struct {
	mu       sync.Mutex
	state    string
	initCmd  Cmd
	updateCmd Cmd
	updates  int
}

func (t *testComponent) Init() Cmd {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.state = "initialized"
	return t.initCmd
}

func (t *testComponent) Update(msg Msg) (Component, Cmd) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	t.updates++
	
	switch m := msg.(type) {
	case testMsg:
		t.state = m.value
	case KeyMsg:
		t.state = "key: " + m.String()
	}
	
	return t, t.updateCmd
}

func (t *testComponent) View() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.state
}

func (t *testComponent) getState() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.state
}

func (t *testComponent) getUpdates() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.updates
}

type testMsg struct {
	value string
}

func TestEngineLifecycle(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "Engine starts and stops cleanly",
			test: func(t *testing.T) {
				comp := &testComponent{}
				engine := NewEngine(comp)
				
				renderMu := sync.Mutex{}
				renderCalled := false
				engine.SetRenderCallback(func(view string) {
					renderMu.Lock()
					renderCalled = true
					renderMu.Unlock()
				})
				
				err := engine.Start()
				if err != nil {
					t.Fatalf("Failed to start engine: %v", err)
				}
				
				// Give engine time to initialize
				time.Sleep(10 * time.Millisecond)
				
				renderMu.Lock()
				called := renderCalled
				renderMu.Unlock()
				
				if !called {
					t.Error("Render callback should have been called")
				}
				
				if comp.getState() != "initialized" {
					t.Errorf("Expected state 'initialized', got '%s'", comp.getState())
				}
				
				engine.Stop()
			},
		},
		{
			name: "Init command is executed",
			test: func(t *testing.T) {
				msgReceived := make(chan bool, 1)
				
				comp := &testComponent{
					initCmd: func() Msg {
						select {
						case msgReceived <- true:
						default:
						}
						return testMsg{value: "from init"}
					},
				}
				
				engine := NewEngine(comp)
				engine.Start()
				
				select {
				case <-msgReceived:
					// Success
				case <-time.After(100 * time.Millisecond):
					t.Error("Init command was not executed")
				}
				
				// Wait for message to be processed
				time.Sleep(20 * time.Millisecond)
				
				if comp.getState() != "from init" {
					t.Errorf("Expected state 'from init', got '%s'", comp.getState())
				}
				
				engine.Stop()
			},
		},
		{
			name: "Messages are processed",
			test: func(t *testing.T) {
				comp := &testComponent{}
				engine := NewEngine(comp)
				
				renderMu := sync.Mutex{}
				renderCount := 0
				engine.SetRenderCallback(func(view string) {
					renderMu.Lock()
					renderCount++
					renderMu.Unlock()
				})
				
				engine.Start()
				
				// Send a message
				engine.SendMessage(testMsg{value: "test message"})
				
				// Wait for processing
				time.Sleep(20 * time.Millisecond)
				
				if comp.getState() != "test message" {
					t.Errorf("Expected state 'test message', got '%s'", comp.getState())
				}
				
				if comp.getUpdates() != 1 {
					t.Errorf("Expected 1 update, got %d", comp.getUpdates())
				}
				
				// Should have rendered at least twice (init + update)
				renderMu.Lock()
				count := renderCount
				renderMu.Unlock()
				
				if count < 2 {
					t.Errorf("Expected at least 2 renders, got %d", count)
				}
				
				engine.Stop()
			},
		},
		{
			name: "Update commands are executed",
			test: func(t *testing.T) {
				cmdExecuted := make(chan bool, 1)
				
				comp := &testComponent{
					updateCmd: func() Msg {
						select {
						case cmdExecuted <- true:
						default:
						}
						return testMsg{value: "from command"}
					},
				}
				
				engine := NewEngine(comp)
				engine.Start()
				
				// Send a message to trigger update
				engine.SendMessage(testMsg{value: "trigger"})
				
				select {
				case <-cmdExecuted:
					// Success
				case <-time.After(100 * time.Millisecond):
					t.Error("Update command was not executed")
				}
				
				// Wait for command result to be processed
				time.Sleep(20 * time.Millisecond)
				
				if comp.getState() != "from command" {
					t.Errorf("Expected state 'from command', got '%s'", comp.getState())
				}
				
				engine.Stop()
			},
		},
		{
			name: "Quit message triggers callback",
			test: func(t *testing.T) {
				comp := &testComponent{}
				engine := NewEngine(comp)
				
				quitCalled := false
				engine.SetQuitCallback(func() {
					quitCalled = true
				})
				
				engine.Start()
				
				// Send quit message
				engine.SendMessage(Quit())
				
				// Wait for processing
				time.Sleep(20 * time.Millisecond)
				
				if !quitCalled {
					t.Error("Quit callback should have been called")
				}
				
				engine.Stop()
			},
		},
		{
			name: "Multiple messages processed in order",
			test: func(t *testing.T) {
				comp := &testComponent{}
				engine := NewEngine(comp)
				
				engine.Start()
				
				// Send multiple messages
				engine.SendMessage(testMsg{value: "first"})
				engine.SendMessage(testMsg{value: "second"})
				engine.SendMessage(testMsg{value: "third"})
				
				// Wait for processing
				time.Sleep(30 * time.Millisecond)
				
				if comp.getState() != "third" {
					t.Errorf("Expected state 'third', got '%s'", comp.getState())
				}
				
				if comp.getUpdates() != 3 {
					t.Errorf("Expected 3 updates, got %d", comp.getUpdates())
				}
				
				engine.Stop()
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}