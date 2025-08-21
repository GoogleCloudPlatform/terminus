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
	"os"
	"os/signal"
	"syscall"

	"github.com/yourusername/terminusgo/pkg/terminus"
)

//go:embed all:static/*
var staticFiles embed.FS

// HelloComponent is a simple example component
type HelloComponent struct {
	message string
	counter int
}

func NewHelloComponent() *HelloComponent {
	return &HelloComponent{
		message: "Welcome to TerminusGo!",
		counter: 0,
	}
}

func (h *HelloComponent) Init() terminus.Cmd {
	return nil
}

func (h *HelloComponent) Update(msg terminus.Msg) (terminus.Component, terminus.Cmd) {
	switch msg := msg.(type) {
	case terminus.KeyMsg:
		switch msg.Type {
		case terminus.KeyCtrlC:
			return h, terminus.Quit
		case terminus.KeyRunes:
			if len(msg.Runes) > 0 && msg.Runes[0] == 'q' {
				return h, terminus.Quit
			}
			h.counter++
			h.message = fmt.Sprintf("You pressed: %s (count: %d)", msg.String(), h.counter)
		case terminus.KeyEnter:
			h.counter = 0
			h.message = "Counter reset!"
		default:
			h.message = fmt.Sprintf("Key pressed: %s", msg.String())
		}
	case terminus.WindowSizeMsg:
		h.message = fmt.Sprintf("Window resized to %dx%d", msg.Width, msg.Height)
	}

	return h, nil
}

func (h *HelloComponent) View() string {
	// Create styled output
	titleStyle := terminus.NewStyle().Bold(true).Foreground(terminus.Cyan)
	messageStyle := terminus.NewStyle().Foreground(terminus.Green)
	instructionStyle := terminus.NewStyle().Faint(true)
	counterStyle := terminus.NewStyle().Bold(true).Foreground(terminus.Yellow)

	title := titleStyle.Render("TerminusGo Example")
	message := messageStyle.Render(h.message)
	instructions := instructionStyle.Render(`Press any key to see it displayed here.
Press Enter to reset the counter.
Press 'q' or Ctrl+C to quit.`)
	counter := counterStyle.Render(fmt.Sprintf("Counter: %d", h.counter))

	return fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s", title, message, instructions, counter)
}

func main() {
	// Component factory
	factory := func() terminus.Component {
		return NewHelloComponent()
	}

	// Create program with static files
	program := terminus.NewProgram(
		factory,
		terminus.WithStaticFiles(staticFiles, "static"),
		terminus.WithAddress(":8888"),
	)

	// Start the program
	if err := program.Start(); err != nil {
		log.Fatalf("Failed to start program: %v", err)
	}

	fmt.Println("TerminusGo is running on http://localhost:8888")
	fmt.Println("Press Ctrl+C to stop...")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")

	// Stop the program
	if err := program.Stop(); err != nil {
		log.Fatalf("Failed to stop program: %v", err)
	}

	fmt.Println("Goodbye!")
}
