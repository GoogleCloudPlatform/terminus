package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/yourusername/terminusgo/pkg/terminus"
	"github.com/yourusername/terminusgo/pkg/terminus/widget"
)

//go:embed all:static/*
var staticFiles embed.FS

// TextInputExample demonstrates the TextInput widget functionality
type TextInputExample struct {
	container  *widget.Container
	nameInput  *widget.TextInput
	emailInput *widget.TextInput
	phoneInput *widget.TextInput
	result     string
	submitted  bool
}

func NewTextInputExample() *TextInputExample {
	ex := &TextInputExample{
		container: widget.NewContainer(),
	}

	// Create text inputs with different configurations
	ex.nameInput = widget.NewTextInput().
		SetPlaceholder("Enter your name...").
		SetMaxLength(50).
		SetStyle(terminus.NewStyle().Foreground(terminus.White)).
		SetFocusStyle(terminus.NewStyle().Foreground(terminus.Cyan).Underline(true)).
		SetPlaceholderStyle(terminus.NewStyle().Faint(true)).
		SetOnChange(func(value string) terminus.Cmd {
			// Real-time validation could go here
			return nil
		}).
		SetOnSubmit(func(value string) terminus.Cmd {
			if strings.TrimSpace(value) == "" {
				ex.result = "Name cannot be empty!"
			} else {
				ex.result = fmt.Sprintf("Hello, %s!", value)
			}
			return nil
		})

	ex.emailInput = widget.NewTextInput().
		SetPlaceholder("Enter your email...").
		SetMaxLength(100).
		SetValidator(func(s string) bool {
			// Simple email validation - just check for @ symbol
			return strings.Count(s, "@") <= 1
		}).
		SetStyle(terminus.NewStyle().Foreground(terminus.White)).
		SetFocusStyle(terminus.NewStyle().Foreground(terminus.Green).Underline(true)).
		SetPlaceholderStyle(terminus.NewStyle().Faint(true)).
		SetOnSubmit(func(value string) terminus.Cmd {
			if !strings.Contains(value, "@") {
				ex.result = "Please enter a valid email address!"
			} else {
				ex.result = fmt.Sprintf("Email set to: %s", value)
			}
			return nil
		})

	ex.phoneInput = widget.NewTextInput().
		SetPlaceholder("Enter phone number (digits only)...").
		SetMaxLength(15).
		SetValidator(func(s string) bool {
			// Only allow digits, spaces, and common phone characters
			for _, r := range s {
				if !((r >= '0' && r <= '9') || r == ' ' || r == '-' || r == '(' || r == ')' || r == '+') {
					return false
				}
			}
			return true
		}).
		SetStyle(terminus.NewStyle().Foreground(terminus.White)).
		SetFocusStyle(terminus.NewStyle().Foreground(terminus.Yellow).Underline(true)).
		SetPlaceholderStyle(terminus.NewStyle().Faint(true)).
		SetOnSubmit(func(value string) terminus.Cmd {
			ex.result = fmt.Sprintf("Phone number: %s", value)
			return nil
		})

	// Set sizes for all inputs
	ex.nameInput.SetSize(40, 1)
	ex.emailInput.SetSize(40, 1)
	ex.phoneInput.SetSize(40, 1)

	// Add inputs to container
	ex.container.AddChild(ex.nameInput)
	ex.container.AddChild(ex.emailInput)
	ex.container.AddChild(ex.phoneInput)

	return ex
}

func (ex *TextInputExample) Init() terminus.Cmd {
	return ex.container.Init()
}

func (ex *TextInputExample) Update(msg terminus.Msg) (terminus.Component, terminus.Cmd) {
	switch msg := msg.(type) {
	case terminus.KeyMsg:
		switch msg.Type {
		case terminus.KeyCtrlC:
			return ex, terminus.Quit
		case terminus.KeyRunes:
			if len(msg.Runes) > 0 && msg.Runes[0] == 'q' {
				return ex, terminus.Quit
			}
		case terminus.KeyCtrlS:
			// Submit all forms
			name := ex.nameInput.Value()
			email := ex.emailInput.Value()
			phone := ex.phoneInput.Value()

			if strings.TrimSpace(name) == "" {
				ex.result = "Please fill in your name!"
			} else if strings.TrimSpace(email) == "" {
				ex.result = "Please fill in your email!"
			} else if !strings.Contains(email, "@") {
				ex.result = "Please enter a valid email address!"
			} else {
				ex.result = fmt.Sprintf("Form submitted!\nName: %s\nEmail: %s\nPhone: %s", name, email, phone)
				ex.submitted = true
			}
			return ex, nil
		case terminus.KeyCtrlR:
			// Reset all forms
			ex.nameInput.Clear()
			ex.emailInput.Clear()
			ex.phoneInput.Clear()
			ex.result = "Form cleared!"
			ex.submitted = false
			return ex, nil
		}
	}

	// Forward to container for focus management and input handling
	newContainer, cmd := ex.container.Update(msg)
	ex.container = newContainer.(*widget.Container)

	return ex, cmd
}

func (ex *TextInputExample) View() string {
	// Create styled output
	titleStyle := terminus.NewStyle().Bold(true).Foreground(terminus.Cyan)
	labelStyle := terminus.NewStyle().Bold(true).Foreground(terminus.White)
	instructionStyle := terminus.NewStyle().Faint(true)
	resultStyle := terminus.NewStyle().Foreground(terminus.Green)
	errorStyle := terminus.NewStyle().Foreground(terminus.Red)

	var result strings.Builder

	// Title
	result.WriteString(titleStyle.Render("TextInput Widget Demo"))
	result.WriteString("\n\n")

	// Instructions
	result.WriteString(instructionStyle.Render("Use Tab/Shift+Tab to navigate between fields"))
	result.WriteString("\n")
	result.WriteString(instructionStyle.Render("Press Enter on a field to submit it individually"))
	result.WriteString("\n")
	result.WriteString(instructionStyle.Render("Press Ctrl+S to submit all fields"))
	result.WriteString("\n")
	result.WriteString(instructionStyle.Render("Press Ctrl+R to reset all fields"))
	result.WriteString("\n")
	result.WriteString(instructionStyle.Render("Press 'q' or Ctrl+C to quit"))
	result.WriteString("\n\n")

	// Form fields
	result.WriteString(labelStyle.Render("Name:"))
	result.WriteString("\n")
	result.WriteString(ex.nameInput.View())
	result.WriteString("\n\n")

	result.WriteString(labelStyle.Render("Email:"))
	result.WriteString("\n")
	result.WriteString(ex.emailInput.View())
	result.WriteString("\n\n")

	result.WriteString(labelStyle.Render("Phone:"))
	result.WriteString("\n")
	result.WriteString(ex.phoneInput.View())
	result.WriteString("\n\n")

	// Result/Status
	if ex.result != "" {
		if ex.submitted || strings.Contains(ex.result, "Hello") || strings.Contains(ex.result, "Email set") || strings.Contains(ex.result, "Phone number") || strings.Contains(ex.result, "Form submitted") || strings.Contains(ex.result, "Form cleared") {
			result.WriteString(resultStyle.Render(ex.result))
		} else {
			result.WriteString(errorStyle.Render(ex.result))
		}
		result.WriteString("\n")
	}

	// Current values display
	result.WriteString("\n")
	result.WriteString(instructionStyle.Render("Current Values:"))
	result.WriteString("\n")
	result.WriteString(instructionStyle.Render(fmt.Sprintf("Name: '%s' | Email: '%s' | Phone: '%s'",
		ex.nameInput.Value(), ex.emailInput.Value(), ex.phoneInput.Value())))

	return result.String()
}

func main() {
	// Component factory
	factory := func() terminus.Component {
		return NewTextInputExample()
	}

	// Create program with static files
	program := terminus.NewProgram(
		factory,
		terminus.WithStaticFiles(staticFiles, "static"),
		terminus.WithAddress(":8890"),
	)

	// Start the program
	if err := program.Start(); err != nil {
		log.Fatalf("Failed to start program: %v", err)
	}

	fmt.Println("TextInput Demo is running on http://localhost:8890")
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
