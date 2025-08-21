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

package widget

import (
	"testing"

	"github.com/yourusername/terminusgo/pkg/terminus"
)

// mockWidget for testing
type mockWidget struct {
	Model
	name string
}

func newMockWidget(name string) *mockWidget {
	return &mockWidget{
		Model: NewModel(),
		name:  name,
	}
}

func (m *mockWidget) Init() terminus.Cmd {
	return nil
}

func (m *mockWidget) Update(msg terminus.Msg) (terminus.Component, terminus.Cmd) {
	return m, nil
}

func (m *mockWidget) View() string {
	return m.name
}

func TestModel(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "Default state",
			test: func(t *testing.T) {
				m := NewModel()
				
				if m.Focused() {
					t.Error("Model should not be focused by default")
				}
				
				width, height := m.GetSize()
				if width != 10 || height != 1 {
					t.Errorf("Expected size (10,1), got (%d,%d)", width, height)
				}
				
				x, y := m.GetPosition()
				if x != 0 || y != 0 {
					t.Errorf("Expected position (0,0), got (%d,%d)", x, y)
				}
			},
		},
		{
			name: "Focus management",
			test: func(t *testing.T) {
				m := NewModel()
				
				m.Focus()
				if !m.Focused() {
					t.Error("Model should be focused after Focus()")
				}
				
				m.Blur()
				if m.Focused() {
					t.Error("Model should not be focused after Blur()")
				}
			},
		},
		{
			name: "Size management",
			test: func(t *testing.T) {
				m := NewModel()
				
				m.SetSize(20, 5)
				width, height := m.GetSize()
				if width != 20 || height != 5 {
					t.Errorf("Expected size (20,5), got (%d,%d)", width, height)
				}
			},
		},
		{
			name: "Position management",
			test: func(t *testing.T) {
				m := NewModel()
				
				m.SetPosition(10, 15)
				x, y := m.GetPosition()
				if x != 10 || y != 15 {
					t.Errorf("Expected position (10,15), got (%d,%d)", x, y)
				}
			},
		},
		{
			name: "Disabled state",
			test: func(t *testing.T) {
				m := NewModel()
				
				if m.Disabled() {
					t.Error("Model should not be disabled by default")
				}
				
				m.SetDisabled(true)
				if !m.Disabled() {
					t.Error("Model should be disabled after SetDisabled(true)")
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

func TestFocusManager(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "Empty focus manager",
			test: func(t *testing.T) {
				fm := NewFocusManager()
				
				if fm.Current() != nil {
					t.Error("Empty focus manager should have no current widget")
				}
				
				// Should not panic
				fm.Next()
				fm.Previous()
			},
		},
		{
			name: "Single widget",
			test: func(t *testing.T) {
				w1 := newMockWidget("widget1")
				fm := NewFocusManager(w1)
				
				if fm.Current() != w1 {
					t.Error("Single widget should be current")
				}
				
				if !w1.Focused() {
					t.Error("Single widget should be focused")
				}
				
				fm.Next()
				if fm.Current() != w1 || !w1.Focused() {
					t.Error("Single widget should remain focused after Next()")
				}
			},
		},
		{
			name: "Multiple widgets navigation",
			test: func(t *testing.T) {
				w1 := newMockWidget("widget1")
				w2 := newMockWidget("widget2")
				w3 := newMockWidget("widget3")
				fm := NewFocusManager(w1, w2, w3)
				
				// Should start with first widget focused
				if fm.Current() != w1 || !w1.Focused() {
					t.Error("First widget should be focused initially")
				}
				
				// Move to next
				fm.Next()
				if fm.Current() != w2 || !w2.Focused() || w1.Focused() {
					t.Error("Second widget should be focused after Next()")
				}
				
				// Move to next again
				fm.Next()
				if fm.Current() != w3 || !w3.Focused() || w2.Focused() {
					t.Error("Third widget should be focused after Next()")
				}
				
				// Wrap around
				fm.Next()
				if fm.Current() != w1 || !w1.Focused() || w3.Focused() {
					t.Error("Should wrap around to first widget")
				}
				
				// Go backwards
				fm.Previous()
				if fm.Current() != w3 || !w3.Focused() || w1.Focused() {
					t.Error("Should go to last widget with Previous()")
				}
			},
		},
		{
			name: "Add widget dynamically",
			test: func(t *testing.T) {
				fm := NewFocusManager()
				w1 := newMockWidget("widget1")
				
				fm.AddWidget(w1)
				
				if fm.Current() != w1 || !w1.Focused() {
					t.Error("Added widget should be focused when first")
				}
				
				w2 := newMockWidget("widget2")
				fm.AddWidget(w2)
				
				// w1 should still be focused
				if fm.Current() != w1 || !w1.Focused() || w2.Focused() {
					t.Error("Original widget should remain focused when adding second")
				}
				
				fm.Next()
				if fm.Current() != w2 || !w2.Focused() || w1.Focused() {
					t.Error("Should be able to navigate to added widget")
				}
			},
		},
		{
			name: "Tab key handling",
			test: func(t *testing.T) {
				w1 := newMockWidget("widget1")
				w2 := newMockWidget("widget2")
				fm := NewFocusManager(w1, w2)
				
				// Tab should move to next widget
				handled := fm.HandleKey(terminus.KeyMsg{Type: terminus.KeyTab})
				if !handled {
					t.Error("Tab key should be handled")
				}
				if fm.Current() != w2 {
					t.Error("Tab should move to next widget")
				}
				
				// Shift+Tab should move to previous widget
				handled = fm.HandleKey(terminus.KeyMsg{Type: terminus.KeyTab, Shift: true})
				if !handled {
					t.Error("Shift+Tab key should be handled")
				}
				if fm.Current() != w1 {
					t.Error("Shift+Tab should move to previous widget")
				}
				
				// Other keys should not be handled
				handled = fm.HandleKey(terminus.KeyMsg{Type: terminus.KeyEnter})
				if handled {
					t.Error("Non-tab keys should not be handled")
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

func TestContainer(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "Empty container",
			test: func(t *testing.T) {
				c := NewContainer()
				
				if len(c.Children()) != 0 {
					t.Error("Empty container should have no children")
				}
				
				view := c.View()
				if view != "" {
					t.Error("Empty container should render empty string")
				}
			},
		},
		{
			name: "Container with children",
			test: func(t *testing.T) {
				c := NewContainer()
				w1 := newMockWidget("widget1")
				w2 := newMockWidget("widget2")
				
				c.AddChild(w1)
				c.AddChild(w2)
				
				children := c.Children()
				if len(children) != 2 {
					t.Errorf("Expected 2 children, got %d", len(children))
				}
				
				view := c.View()
				expected := "widget1\nwidget2"
				if view != expected {
					t.Errorf("Expected view %q, got %q", expected, view)
				}
			},
		},
		{
			name: "Container focus management",
			test: func(t *testing.T) {
				c := NewContainer()
				w1 := newMockWidget("widget1")
				w2 := newMockWidget("widget2")
				
				c.AddChild(w1)
				c.AddChild(w2)
				
				// First widget should be focused
				if !w1.Focused() {
					t.Error("First widget should be focused")
				}
				
				// Tab should move focus
				c.Update(terminus.KeyMsg{Type: terminus.KeyTab})
				if !w2.Focused() || w1.Focused() {
					t.Error("Tab should move focus to second widget")
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