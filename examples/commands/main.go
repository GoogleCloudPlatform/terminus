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
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/yourusername/terminusgo/pkg/terminus"
	"github.com/yourusername/terminusgo/pkg/terminus/widget"
)

//go:embed all:static/*
var staticFiles embed.FS

// CommandDemo demonstrates advanced command features
type CommandDemo struct {
	// UI components
	log          []string
	statusText   string
	statusStyle  terminus.Style
	spinner      *widget.Spinner
	searchInput  *widget.TextInput
	httpStatus   string

	// State
	tickCount    int
	searchQuery  string
	isLoading    bool
	activeTimers map[string]bool
}

func NewCommandDemo() *CommandDemo {
	demo := &CommandDemo{
		log:          make([]string, 0),
		statusStyle:  terminus.NewStyle().Foreground(terminus.Green),
		spinner:      widget.NewSpinner().SetText("Idle").SetSpinnerStyle(widget.SpinnerDots),
		activeTimers: make(map[string]bool),
	}

	// Initialize search input with debouncing
	demo.searchInput = widget.NewTextInput().
		SetPlaceholder("Type to search (debounced)...").
		SetMaxLength(50).
		SetOnChange(func(value string) terminus.Cmd {
			demo.searchQuery = value
			if value == "" {
				return nil
			}
			// Debounce the search
			return terminus.Debounce("search", 500*time.Millisecond, func() terminus.Msg {
				return SearchMsg{Query: value}
			})
		})

	demo.searchInput.SetSize(40, 1)
	return demo
}

// Message types
type TickMsg struct {
	Time time.Time
}

type SearchMsg struct {
	Query string
}

type TimerStartedMsg struct {
	ID string
}

type TimerStoppedMsg struct {
	ID string
}

func (d *CommandDemo) Init() terminus.Cmd {
	d.addLog("Command Demo initialized")
	// Start a periodic tick every 2 seconds
	return d.startTicker()
}

func (d *CommandDemo) Update(msg terminus.Msg) (terminus.Component, terminus.Cmd) {
	switch msg := msg.(type) {
	case terminus.KeyMsg:
		switch msg.Type {
		case terminus.KeyCtrlC:
			return d, terminus.Quit

		case terminus.KeyRunes:
			if len(msg.Runes) > 0 {
				switch msg.Runes[0] {
				case 'q', 'Q':
					return d, terminus.Quit

				case '1':
					// Demonstrate Sequence command
					d.addLog("Starting sequential operations...")
					return d, terminus.Sequence(
						d.delayedCmd("Step 1", 500*time.Millisecond),
						d.delayedCmd("Step 2", 500*time.Millisecond),
						d.delayedCmd("Step 3", 500*time.Millisecond),
					)

				case '2':
					// Demonstrate Parallel command
					d.addLog("Starting parallel operations...")
					return d, terminus.Parallel(
						d.delayedCmd("Parallel 1", 1*time.Second),
						d.delayedCmd("Parallel 2", 1*time.Second),
						d.delayedCmd("Parallel 3", 1*time.Second),
					)

				case '3':
					// Demonstrate HTTP request
					d.addLog("Making HTTP request...")
					d.httpStatus = "Loading..."
					d.spinner.Start()
					return d, terminus.Get("https://api.github.com/zen")

				case '4':
					// Demonstrate cancellable timer
					if !d.activeTimers["demo"] {
						d.addLog("Starting cancellable timer (5s)...")
						d.activeTimers["demo"] = true
						return d, terminus.WithCancel("demo-timer", func(ctx context.Context) terminus.Msg {
							select {
							case <-time.After(5 * time.Second):
								return TimerStoppedMsg{ID: "demo"}
							case <-ctx.Done():
								return TimerStoppedMsg{ID: "demo"}
							}
						})
					} else {
						d.addLog("Cancelling timer...")
						terminus.Cancel("demo-timer")
						d.activeTimers["demo"] = false
					}

				case '5':
					// Demonstrate throttled command
					d.addLog("Throttled command called (max once per second)")
					return d, terminus.Throttle("throttle-demo", 1*time.Second, func() terminus.Msg {
						d.addLog("Throttled command executed!")
						return nil
					})

				case 'c', 'C':
					// Clear log
					d.log = []string{}
					d.addLog("Log cleared")
				}
			}

		default:
			// Forward to search input
			if d.searchInput.Focused() {
				newInput, cmd := d.searchInput.Update(msg)
				d.searchInput = newInput.(*widget.TextInput)
				return d, cmd
			}
		}

	case TickMsg:
		d.tickCount++
		d.statusText = fmt.Sprintf("Tick #%d at %s", d.tickCount, msg.Time.Format("15:04:05"))
		// Continue ticking
		return d, d.startTicker()

	case SearchMsg:
		d.addLog(fmt.Sprintf("Search executed for: '%s'", msg.Query))
		// Simulate search API call
		return d, d.simulateSearch(msg.Query)

	case terminus.HTTPRequestMsg:
		d.spinner.Stop()
		if msg.Error != nil {
			d.httpStatus = fmt.Sprintf("Error: %v", msg.Error)
			d.addLog("HTTP request failed")
		} else {
			d.httpStatus = fmt.Sprintf("Response: %s", strings.TrimSpace(msg.String()))
			d.addLog(fmt.Sprintf("HTTP request completed: %d", msg.StatusCode()))
		}

	case TimerStartedMsg:
		d.addLog(fmt.Sprintf("Timer '%s' started", msg.ID))

	case TimerStoppedMsg:
		d.addLog(fmt.Sprintf("Timer '%s' stopped", msg.ID))
		d.activeTimers[msg.ID] = false

	case widget.SpinnerTickMsg:
		// Forward to spinner
		newSpinner, cmd := d.spinner.Update(msg)
		d.spinner = newSpinner.(*widget.Spinner)
		return d, cmd
	}

	return d, nil
}

func (d *CommandDemo) View() string {
	var result strings.Builder

	// Title
	titleStyle := terminus.NewStyle().Bold(true).Foreground(terminus.Cyan)
	result.WriteString(titleStyle.Render("Advanced Commands Demo"))
	result.WriteString("\n\n")

	// Instructions
	instructionStyle := terminus.NewStyle().Faint(true)
	result.WriteString(instructionStyle.Render("Commands:"))
	result.WriteString("\n")
	result.WriteString(instructionStyle.Render("1 - Run sequential commands"))
	result.WriteString("\n")
	result.WriteString(instructionStyle.Render("2 - Run parallel commands"))
	result.WriteString("\n")
	result.WriteString(instructionStyle.Render("3 - Make HTTP request"))
	result.WriteString("\n")
	result.WriteString(instructionStyle.Render("4 - Start/cancel timer"))
	result.WriteString("\n")
	result.WriteString(instructionStyle.Render("5 - Throttled command"))
	result.WriteString("\n")
	result.WriteString(instructionStyle.Render("c - Clear log"))
	result.WriteString("\n")
	result.WriteString(instructionStyle.Render("q - Quit"))
	result.WriteString("\n\n")

	// Search input
	result.WriteString("Search (debounced):\n")
	result.WriteString(d.searchInput.View())
	result.WriteString("\n\n")

	// Status
	if d.statusText != "" {
		result.WriteString(d.statusStyle.Render("Status: " + d.statusText))
		result.WriteString("\n")
	}

	// HTTP Status
	if d.httpStatus != "" {
		result.WriteString("HTTP: ")
		result.WriteString(d.httpStatus)
		result.WriteString(" ")
		result.WriteString(d.spinner.View())
		result.WriteString("\n")
	}

	result.WriteString("\n")

	// Log
	logStyle := terminus.NewStyle().Foreground(terminus.White)
	result.WriteString(logStyle.Render("Activity Log:"))
	result.WriteString("\n")
	result.WriteString(strings.Repeat("-", 50))
	result.WriteString("\n")

	// Show last 10 log entries
	start := 0
	if len(d.log) > 10 {
		start = len(d.log) - 10
	}

	for i := start; i < len(d.log); i++ {
		result.WriteString(fmt.Sprintf("[%02d] %s\n", i+1, d.log[i]))
	}

	return result.String()
}

func (d *CommandDemo) addLog(message string) {
	timestamp := time.Now().Format("15:04:05")
	d.log = append(d.log, fmt.Sprintf("%s - %s", timestamp, message))
}

func (d *CommandDemo) startTicker() terminus.Cmd {
	return terminus.Tick(2*time.Second, func(t time.Time) terminus.Msg {
		return TickMsg{Time: t}
	})
}

func (d *CommandDemo) delayedCmd(name string, delay time.Duration) terminus.Cmd {
	return func() terminus.Msg {
		d.addLog(fmt.Sprintf("%s started", name))
		time.Sleep(delay)
		d.addLog(fmt.Sprintf("%s completed", name))
		return nil
	}
}

func (d *CommandDemo) simulateSearch(query string) terminus.Cmd {
	return terminus.Tick(300*time.Millisecond, func(t time.Time) terminus.Msg {
		d.addLog(fmt.Sprintf("Search results for '%s' retrieved", query))
		return nil
	})
}

func main() {
	// Component factory
	factory := func() terminus.Component {
		return NewCommandDemo()
	}

	// Create program
	program := terminus.NewProgram(
		factory,
		terminus.WithStaticFiles(staticFiles, "static"),
		terminus.WithAddress(":8890"),
	)

	// Start the program
	if err := program.Start(); err != nil {
		log.Fatalf("Failed to start program: %v", err)
	}

	fmt.Println("Command Demo is running on http://localhost:8890")
	fmt.Println("Press Ctrl+C to stop...")

	// Wait for the program to run
	program.Wait()

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