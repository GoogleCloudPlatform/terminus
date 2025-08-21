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
	"testing"
)

// mockComponent is a test implementation of the Component interface
type mockComponent struct {
	initCalled   bool
	updateCalled bool
	viewCalled   bool
	state        string
}

func (m *mockComponent) Init() Cmd {
	m.initCalled = true
	return nil
}

func (m *mockComponent) Update(msg Msg) (Component, Cmd) {
	m.updateCalled = true
	switch msg := msg.(type) {
	case KeyMsg:
		m.state = msg.String()
	}
	return m, nil
}

func (m *mockComponent) View() string {
	m.viewCalled = true
	return m.state
}

func TestComponentInterface(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "Component implements interface",
			test: func(t *testing.T) {
				var c Component = &mockComponent{}
				if c == nil {
					t.Fatal("Component should not be nil")
				}
			},
		},
		{
			name: "Init is called",
			test: func(t *testing.T) {
				c := &mockComponent{}
				cmd := c.Init()
				if !c.initCalled {
					t.Error("Init should have been called")
				}
				if cmd != nil {
					t.Error("Init should return nil command for mock")
				}
			},
		},
		{
			name: "Update is called with message",
			test: func(t *testing.T) {
				c := &mockComponent{}
				msg := KeyMsg{Type: KeyRunes, Runes: []rune{'a'}}
				newC, cmd := c.Update(msg)
				
				if !c.updateCalled {
					t.Error("Update should have been called")
				}
				if newC != c {
					t.Error("Update should return the same component for mock")
				}
				if cmd != nil {
					t.Error("Update should return nil command for mock")
				}
				if c.state != "a" {
					t.Errorf("Expected state to be 'a', got '%s'", c.state)
				}
			},
		},
		{
			name: "View is called",
			test: func(t *testing.T) {
				c := &mockComponent{state: "test view"}
				view := c.View()
				
				if !c.viewCalled {
					t.Error("View should have been called")
				}
				if view != "test view" {
					t.Errorf("Expected view to be 'test view', got '%s'", view)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}