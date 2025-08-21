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
	"time"

	"github.com/yourusername/terminusgo/pkg/terminus"
)

func TestSpinner(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "Default state",
			test: func(t *testing.T) {
				spinner := NewSpinner()

				if spinner.IsSpinning() {
					t.Error("New spinner should not be spinning")
				}

				if spinner.Text() != "" {
					t.Error("New spinner should have empty text")
				}

				if spinner.Frame() != 0 {
					t.Error("New spinner should start at frame 0")
				}

				if spinner.ElapsedTime() != 0 {
					t.Error("New spinner should have 0 elapsed time")
				}
			},
		},
		{
			name: "Start and stop",
			test: func(t *testing.T) {
				spinner := NewSpinner()

				// Start spinner
				spinner.Start()
				if !spinner.IsSpinning() {
					t.Error("Spinner should be spinning after Start()")
				}

				if spinner.ElapsedTime() == 0 {
					t.Error("Spinner should have non-zero elapsed time after start")
				}

				// Stop spinner
				spinner.Stop()
				if spinner.IsSpinning() {
					t.Error("Spinner should not be spinning after Stop()")
				}
			},
		},
		{
			name: "Text configuration",
			test: func(t *testing.T) {
				spinner := NewSpinner()
				spinner.SetText("Loading...")

				if spinner.Text() != "Loading..." {
					t.Errorf("Expected text 'Loading...', got '%s'", spinner.Text())
				}

				// View should contain text even when not spinning
				view := spinner.View()
				if view == "" {
					t.Error("View should not be empty when text is set")
				}
			},
		},
		{
			name: "Custom characters",
			test: func(t *testing.T) {
				spinner := NewSpinner()
				customChars := []string{"1", "2", "3", "4"}
				spinner.SetCustomChars(customChars)

				// Test that custom chars are used
				chars := spinner.getChars()
				if len(chars) != 4 {
					t.Errorf("Expected 4 custom chars, got %d", len(chars))
				}

				for i, char := range customChars {
					if chars[i] != char {
						t.Errorf("Expected char %s at index %d, got %s", char, i, chars[i])
					}
				}
			},
		},
		{
			name: "Animation frames",
			test: func(t *testing.T) {
				spinner := NewSpinner()
				spinner.Start()

				initialFrame := spinner.Frame()

				// Simulate tick message
				spinner.Update(SpinnerTickMsg{ID: "spinner"})

				if spinner.Frame() <= initialFrame {
					t.Error("Frame should advance after tick message")
				}
			},
		},
		{
			name: "Frame cycling",
			test: func(t *testing.T) {
				spinner := NewSpinner()
				customChars := []string{"A", "B", "C"}
				spinner.SetCustomChars(customChars)

				// Test frame cycling
				spinner.SetFrame(0)
				if spinner.getCurrentChar() != "A" {
					t.Errorf("Expected 'A' at frame 0, got '%s'", spinner.getCurrentChar())
				}

				spinner.SetFrame(1)
				if spinner.getCurrentChar() != "B" {
					t.Errorf("Expected 'B' at frame 1, got '%s'", spinner.getCurrentChar())
				}

				spinner.SetFrame(3) // Should cycle back to first
				if spinner.getCurrentChar() != "A" {
					t.Errorf("Expected 'A' at frame 3 (cycled), got '%s'", spinner.getCurrentChar())
				}
			},
		},
		{
			name: "Text positions",
			test: func(t *testing.T) {
				spinner := NewSpinner()
				spinner.SetText("Test").Start()
				spinner.SetFrame(0) // Use consistent frame for testing

				// Test different positions
				positions := []TextPosition{TextLeft, TextRight, TextAbove, TextBelow}
				for _, pos := range positions {
					spinner.SetTextPosition(pos)
					view := spinner.View()
					if view == "" {
						t.Errorf("View should not be empty for position %d", pos)
					}
					// Each position should produce different output
				}
			},
		},
		{
			name: "Spinner styles",
			test: func(t *testing.T) {
				spinner := NewSpinner()
				spinner.Start()

				// Test different spinner styles
				styles := []SpinnerStyle{SpinnerDots, SpinnerLine, SpinnerCircle, SpinnerArrow}
				for _, style := range styles {
					spinner.SetSpinnerStyle(style)
					chars := spinner.getChars()
					if len(chars) == 0 {
						t.Errorf("Style %d should have characters", style)
					}
				}
			},
		},
		{
			name: "Speed configuration",
			test: func(t *testing.T) {
				spinner := NewSpinner()
				newSpeed := 50 * time.Millisecond
				spinner.SetSpeed(newSpeed)

				if spinner.speed != newSpeed {
					t.Errorf("Expected speed %v, got %v", newSpeed, spinner.speed)
				}
			},
		},
		{
			name: "Reset functionality",
			test: func(t *testing.T) {
				spinner := NewSpinner()
				spinner.Start()
				spinner.SetFrame(5)

				// Reset should stop and reset frame
				spinner.Reset()

				if spinner.IsSpinning() {
					t.Error("Spinner should not be spinning after Reset()")
				}

				if spinner.Frame() != 0 {
					t.Errorf("Frame should be 0 after Reset(), got %d", spinner.Frame())
				}
			},
		},
		{
			name: "WithLoadingText convenience method",
			test: func(t *testing.T) {
				spinner := NewSpinner()
				spinner.WithLoadingText("Processing...")

				if spinner.Text() != "Processing..." {
					t.Errorf("Expected text 'Processing...', got '%s'", spinner.Text())
				}

				if !spinner.IsSpinning() {
					t.Error("Spinner should be spinning after WithLoadingText()")
				}
			},
		},
		{
			name: "View rendering with different states",
			test: func(t *testing.T) {
				spinner := NewSpinner()

				// Empty state
				view := spinner.View()
				if view != "" {
					t.Error("Empty spinner should have empty view")
				}

				// Text only
				spinner.SetText("Loading...")
				view = spinner.View()
				if view == "" {
					t.Error("Spinner with text should have non-empty view")
				}

				// Spinning with text
				spinner.Start()
				view = spinner.View()
				if view == "" {
					t.Error("Spinning spinner with text should have non-empty view")
				}

				// Spinning without text
				spinner.SetText("")
				view = spinner.View()
				if view == "" {
					t.Error("Spinning spinner should have non-empty view even without text")
				}
			},
		},
		{
			name: "Update message handling",
			test: func(t *testing.T) {
				spinner := NewSpinner()
				spinner.Start()

				initialFrame := spinner.Frame()

				// Update with tick message
				newSpinner, cmd := spinner.Update(SpinnerTickMsg{ID: "spinner"})
				spinner = newSpinner.(*Spinner)

				if spinner.Frame() <= initialFrame {
					t.Error("Frame should advance after tick message")
				}

				if cmd == nil {
					t.Error("Update should return a command to continue animation")
				}

				// Update with non-tick message should not affect frame
				initialFrame = spinner.Frame()
				newSpinner, _ = spinner.Update(terminus.KeyMsg{Type: terminus.KeyEnter})
				spinner = newSpinner.(*Spinner)

				if spinner.Frame() != initialFrame {
					t.Error("Frame should not change for non-tick messages")
				}
			},
		},
		{
			name: "Stopped spinner ignores ticks",
			test: func(t *testing.T) {
				spinner := NewSpinner()
				// Don't start the spinner

				initialFrame := spinner.Frame()

				// Tick message should be ignored when not spinning
				newSpinner, cmd := spinner.Update(SpinnerTickMsg{ID: "spinner"})
				spinner = newSpinner.(*Spinner)

				if spinner.Frame() != initialFrame {
					t.Error("Stopped spinner should ignore tick messages")
				}

				if cmd != nil {
					t.Error("Stopped spinner should not return commands for tick messages")
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

func TestSpinnerPresets(t *testing.T) {
	tests := []struct {
		name    string
		spinner *Spinner
	}{
		{"LoadingSpinner", NewLoadingSpinner()},
		{"ProcessingSpinner", NewProcessingSpinner()},
		{"SavingSpinner", NewSavingSpinner()},
		{"MinimalSpinner", NewMinimalSpinner()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.spinner == nil {
				t.Error("Preset spinner should not be nil")
			}

			// All presets should have valid configuration
			chars := tt.spinner.getChars()
			if len(chars) == 0 {
				t.Error("Preset spinner should have characters")
			}

			if tt.spinner.speed <= 0 {
				t.Error("Preset spinner should have positive speed")
			}
		})
	}
}

func TestSpinnerChaining(t *testing.T) {
	// Test that all setter methods return *Spinner for method chaining
	spinner := NewSpinner().
		SetSpinnerStyle(SpinnerLine).
		SetCustomChars([]string{"1", "2", "3"}).
		SetText("Loading...").
		SetTextPosition(TextRight).
		SetSpeed(50 * time.Millisecond).
		SetStyle(terminus.NewStyle()).
		SetTextStyle(terminus.NewStyle()).
		SetSpinnerColor(terminus.NewStyle()).
		Start().
		SetFrame(1)

	if !spinner.IsSpinning() {
		t.Error("Method chaining should work correctly")
	}

	if spinner.Frame() != 1 {
		t.Error("Method chaining should work correctly")
	}
}