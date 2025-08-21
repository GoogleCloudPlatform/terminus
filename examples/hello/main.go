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
	"embed"
	"fmt"
	"log"

	"github.com/yourusername/terminusgo/pkg/terminus"
	"github.com/yourusername/terminusgo/pkg/terminus/style"
)

//go:embed all:static/*
var staticFiles embed.FS

// HelloModel represents the state of our Hello World application
type HelloModel struct {
	// The greeting message to display
	greeting string
	// User's name input
	name string
	// Whether we're collecting the user's name
	collectingName bool
	// Whether to show the personalized greeting
	showPersonalized bool
}

// HelloComponent is our main component that implements terminus.Component
type HelloComponent struct {
	model HelloModel
}

// NewHelloComponent creates a new instance of the Hello component
func NewHelloComponent() *HelloComponent {
	return &HelloComponent{
		model: HelloModel{
			greeting:         "Hello, World!",
			name:             "",
			collectingName:   false,
			showPersonalized: false,
		},
	}
}

// Init is called when the component starts
// It can return an initial command to execute
func (h *HelloComponent) Init() terminus.Cmd {
	// No initial command needed for this example
	return nil
}

// Update handles incoming messages and updates the component's state
// This is the heart of the MVU (Model-View-Update) pattern
func (h *HelloComponent) Update(msg terminus.Msg) (terminus.Component, terminus.Cmd) {
	// Handle different types of messages
	switch msg := msg.(type) {
	case terminus.KeyMsg:
		// Handle keyboard input
		switch msg.Type {
		case terminus.KeyEnter:
			if h.model.collectingName && h.model.name != "" {
				// User pressed Enter after typing their name
				h.model.showPersonalized = true
				h.model.collectingName = false
			} else if !h.model.collectingName && !h.model.showPersonalized {
				// User pressed Enter to start entering their name
				h.model.collectingName = true
			}
			return h, nil

		case terminus.KeyBackspace:
			// Remove last character from name
			if h.model.collectingName && len(h.model.name) > 0 {
				h.model.name = h.model.name[:len(h.model.name)-1]
			}
			return h, nil

		case terminus.KeyRunes:
			// Add typed characters to name
			if h.model.collectingName {
				h.model.name += string(msg.Runes)
			}
			return h, nil

		case terminus.KeyEsc:
			// Reset to initial state
			h.model.collectingName = false
			h.model.showPersonalized = false
			h.model.name = ""
			return h, nil
		}

		// Handle specific key strings
		switch msg.String() {
		case "q", "ctrl+c":
			// Quit the application
			return h, terminus.Quit
		case "r":
			// Reset the application
			if !h.model.collectingName {
				h.model.collectingName = false
				h.model.showPersonalized = false
				h.model.name = ""
			}
			return h, nil
		}
	}

	// Return the component unchanged if the message wasn't handled
	return h, nil
}

// View renders the current state of the component as a string
// This is what the user sees in their terminal
func (h *HelloComponent) View() string {
	// Create style instances for different text styles
	titleStyle := style.New().Bold(true).Foreground(style.Cyan)
	promptStyle := style.New().Foreground(style.Yellow)
	inputStyle := style.New().Foreground(style.Green)
	instructionStyle := style.New().Faint(true)
	greetingStyle := style.New().Bold(true).Foreground(style.Magenta)

	// Build the view based on current state
	view := titleStyle.Render("=== TerminusGo Hello World Example ===") + "\n\n"

	if h.model.showPersonalized {
		// Show personalized greeting
		greeting := fmt.Sprintf("Hello, %s! Welcome to TerminusGo!", h.model.name)
		view += greetingStyle.Render(greeting) + "\n\n"
		view += instructionStyle.Render("Press 'r' to reset, 'q' to quit") + "\n"
	} else if h.model.collectingName {
		// Collecting user's name
		view += promptStyle.Render("What's your name? ") 
		view += inputStyle.Render(h.model.name)
		// Show a blinking cursor
		view += style.New().Blink(true).Render("│") + "\n\n"
		view += instructionStyle.Render("Press Enter to submit, Esc to cancel") + "\n"
	} else {
		// Initial state
		view += h.model.greeting + "\n\n"
		view += promptStyle.Render("Press Enter to personalize this greeting!") + "\n\n"
		view += instructionStyle.Render("Press 'q' to quit") + "\n"
	}

	// Add some spacing and a footer
	view += "\n"
	view += instructionStyle.Render("This example demonstrates basic MVU patterns in TerminusGo") + "\n"

	return view
}

func main() {
	// Create and configure the TerminusGo program
	// The factory function creates a new instance of the component for each session
	program := terminus.NewProgram(
		func() terminus.Component {
			return NewHelloComponent()
		},
		terminus.WithStaticFiles(staticFiles, "static"),
		terminus.WithAddress(":8890"),
	)

	// Start the server
	if err := program.Start(); err != nil {
		log.Fatalf("Failed to start program: %v", err)
	}
	
	fmt.Println("TerminusGo Hello World example is running on http://localhost:8890")
	fmt.Println("Press Ctrl+C to stop...")
	
	// Wait for the program to finish
	program.Wait()
}