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

	"github.com/google/generative-ai-go/genai"
	"github.com/yourusername/terminusgo/pkg/terminus"
	"github.com/yourusername/terminusgo/pkg/terminus/style"
	"github.com/yourusername/terminusgo/pkg/terminus/widget"
	"google.golang.org/api/option"
)

//go:embed all:static/*
var staticFiles embed.FS

// Message represents a chat message
type Message struct {
	Role      string // "user" or "assistant"
	Content   string
	Timestamp time.Time
}

// GeminiChatModel represents the application state
type GeminiChatModel struct {
	messages      []Message
	input         *widget.TextInput
	client        *genai.Client
	chat          *genai.ChatSession
	apiKey        string
	isConnected   bool
	isWaiting     bool
	error         string
	showTimestamp bool
	scrollOffset  int
	viewHeight    int
}

// GeminiChatComponent is the main component
type GeminiChatComponent struct {
	model GeminiChatModel
}

// NewGeminiChatComponent creates a new Gemini chat component
func NewGeminiChatComponent() *GeminiChatComponent {
	input := widget.NewTextInput().
		SetPlaceholder("Type your message...").
		SetStyle(terminus.NewStyle()).
		SetFocusStyle(terminus.NewStyle()).
		SetPlaceholderStyle(terminus.NewStyle().Faint(true)).
		SetCursorChar('█')
	
	// Set a reasonable width for the input
	input.SetSize(80, 1)

	return &GeminiChatComponent{
		model: GeminiChatModel{
			messages:      []Message{},
			input:         input,
			apiKey:        os.Getenv("GEMINI_API_KEY"),
			isConnected:   false,
			isWaiting:     false,
			showTimestamp: true,
			scrollOffset:  0,
			viewHeight:    20, // Default view height
		},
	}
}

// Init initializes the component
func (g *GeminiChatComponent) Init() terminus.Cmd {
	g.model.input.Focus()

	// Initialize Gemini client
	if g.model.apiKey == "" {
		g.model.error = "GEMINI_API_KEY environment variable not set"
		g.addSystemMessage("Error: Please set GEMINI_API_KEY environment variable")
		return nil
	}

	return g.connectToGemini()
}

// connectToGemini creates the Gemini client and chat session
func (g *GeminiChatComponent) connectToGemini() terminus.Cmd {
	return func() terminus.Msg {
		ctx := context.Background()
		client, err := genai.NewClient(ctx, option.WithAPIKey(g.model.apiKey))
		if err != nil {
			return GeminiErrorMsg{Error: err}
		}

		// Get the model
		model := client.GenerativeModel("gemini-2.5-flash-preview-05-20")

		// Start a chat session
		chat := model.StartChat()

		return GeminiConnectedMsg{Client: client, Chat: chat}
	}
}

// Update handles messages
func (g *GeminiChatComponent) Update(msg terminus.Msg) (terminus.Component, terminus.Cmd) {
	switch msg := msg.(type) {
	case terminus.KeyMsg:
		switch msg.Type {
		case terminus.KeyEnter:
			if g.model.input.Value() != "" && !g.model.isWaiting {
				userMessage := g.model.input.Value()
				g.model.input.SetValue("")

				// Add user message
				g.addMessage("user", userMessage)

				// Send to Gemini
				if g.model.isConnected {
					g.model.isWaiting = true
					return g, g.sendToGemini(userMessage)
				}
			}
			return g, nil

		case terminus.KeyEsc:
			return g, terminus.Quit

		case terminus.KeyCtrlC:
			return g, terminus.Quit
			
		case terminus.KeyUp:
			// Scroll up
			if g.model.scrollOffset > 0 {
				g.model.scrollOffset--
			}
			return g, nil
			
		case terminus.KeyDown:
			// Scroll down
			g.model.scrollOffset++
			return g, nil
			
		case terminus.KeyPgUp:
			// Page up
			g.model.scrollOffset -= g.model.viewHeight - 1
			if g.model.scrollOffset < 0 {
				g.model.scrollOffset = 0
			}
			return g, nil
			
		case terminus.KeyPgDown:
			// Page down
			g.model.scrollOffset += g.model.viewHeight - 1
			return g, nil
			
		case terminus.KeyHome:
			// Go to top
			g.model.scrollOffset = 0
			return g, nil
			
		case terminus.KeyEnd:
			// Go to bottom
			g.model.scrollOffset = 999999 // Will be adjusted in renderMessages
			return g, nil
		}

		// Check for manual clear/timestamp toggle
		if msg.Type == terminus.KeyRunes && len(msg.Runes) > 0 {
			if msg.Ctrl && msg.Runes[0] == 'l' {
				// Clear chat
				g.model.messages = []Message{}
				g.model.scrollOffset = 0
				if g.model.chat != nil {
					// Recreate chat to clear history
					return g, g.connectToGemini()
				}
				return g, nil
			} else if msg.Ctrl && msg.Runes[0] == 't' {
				// Toggle timestamps
				g.model.showTimestamp = !g.model.showTimestamp
				return g, nil
			}
		}

		// Convert KeySpace to KeyRunes for the input widget
		if msg.Type == terminus.KeySpace {
			msg = terminus.KeyMsg{
				Type:  terminus.KeyRunes,
				Runes: []rune{' '},
			}
		}
		
		// Pass other keys to input
		var cmd terminus.Cmd
		inputComp, cmd := g.model.input.Update(msg)
		if ti, ok := inputComp.(*widget.TextInput); ok {
			g.model.input = ti
		}
		return g, cmd

	case GeminiConnectedMsg:
		g.model.client = msg.Client
		g.model.chat = msg.Chat
		g.model.isConnected = true
		g.model.error = ""
		g.addSystemMessage("Connected to Gemini. Start chatting!")
		return g, nil

	case GeminiResponseMsg:
		g.model.isWaiting = false
		g.addMessage("assistant", msg.Response)
		return g, nil

	case GeminiErrorMsg:
		g.model.isWaiting = false
		g.model.error = msg.Error.Error()
		g.addSystemMessage(fmt.Sprintf("Error: %v", msg.Error))
		return g, nil

	case terminus.WindowSizeMsg:
		// Window resize handled automatically by terminal
		return g, nil
	}

	return g, nil
}

// View renders the component
func (g *GeminiChatComponent) View() string {
	// Title
	title := style.New().
		Bold(true).
		Foreground(style.Cyan).
		Render("🤖 Gemini Chat")

	// Status line
	var status string
	if g.model.error != "" {
		status = style.New().Foreground(style.Red).Render("❌ " + g.model.error)
	} else if g.model.isWaiting {
		status = style.New().Foreground(style.Yellow).Render("⏳ Waiting for response...")
	} else if g.model.isConnected {
		status = style.New().Foreground(style.Green).Render("✓ Connected")
	} else {
		status = style.New().Foreground(style.Yellow).Render("⏳ Connecting...")
	}

	// Message count
	msgCount := style.New().Faint(true).Render(fmt.Sprintf("%d messages", len(g.model.messages)))

	// Header
	header := fmt.Sprintf("%s  %s  %s", title, status, msgCount)

	// Separator
	separator := strings.Repeat("─", 120)

	// Help text
	help := style.New().Faint(true).Render(
		"Enter: send | Ctrl+L: clear | Ctrl+T: toggle timestamps | Ctrl+C: quit")

	// Input section with prompt
	inputSection := fmt.Sprintf("%s %s", 
		style.New().Foreground(style.Green).Bold(true).Render("You:"),
		g.model.input.View())

	// Render messages
	messagesView := g.renderMessages()
	
	return fmt.Sprintf(`%s
%s

%s

%s
%s

%s`,
		header,
		separator,
		messagesView,
		separator,
		inputSection,
		help,
	)
}

// sendToGemini sends a message to Gemini and returns the response
func (g *GeminiChatComponent) sendToGemini(message string) terminus.Cmd {
	return func() terminus.Msg {
		if g.model.chat == nil {
			return GeminiErrorMsg{Error: fmt.Errorf("not connected to Gemini")}
		}

		ctx := context.Background()
		resp, err := g.model.chat.SendMessage(ctx, genai.Text(message))
		if err != nil {
			return GeminiErrorMsg{Error: err}
		}

		if len(resp.Candidates) == 0 {
			return GeminiResponseMsg{Response: "No response from Gemini (no candidates)"}
		}

		var response string
		for _, part := range resp.Candidates[0].Content.Parts {
			if textPart, ok := part.(genai.Text); ok {
				response += string(textPart)
			} else {
				response += fmt.Sprintf("[%T]", part)
			}
		}

		if response == "" {
			response = "No response from Gemini (empty response)"
		}

		return GeminiResponseMsg{Response: response}
	}
}

// renderMessages renders all messages with scrolling
func (g *GeminiChatComponent) renderMessages() string {
	if len(g.model.messages) == 0 {
		return style.New().Faint(true).Render("No messages yet. Start chatting!")
	}
	
	// Build all message lines
	var allLines []string
	for _, msg := range g.model.messages {
		lines := g.formatMessage(msg)
		allLines = append(allLines, lines...)
		allLines = append(allLines, "") // Empty line between messages
	}
	
	// Remove last empty line
	if len(allLines) > 0 && allLines[len(allLines)-1] == "" {
		allLines = allLines[:len(allLines)-1]
	}
	
	// Adjust scroll offset
	maxOffset := len(allLines) - g.model.viewHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if g.model.scrollOffset > maxOffset {
		g.model.scrollOffset = maxOffset
	}
	if g.model.scrollOffset < 0 {
		g.model.scrollOffset = 0
	}
	
	// Get visible lines
	start := g.model.scrollOffset
	end := start + g.model.viewHeight
	if end > len(allLines) {
		end = len(allLines)
	}
	
	var result []string
	for i := start; i < end; i++ {
		result = append(result, allLines[i])
	}
	
	// Pad to view height
	for len(result) < g.model.viewHeight {
		result = append(result, "")
	}
	
	// Add scroll indicators
	if g.model.scrollOffset > 0 {
		result[0] = result[0] + style.New().Faint(true).Render(" ↑ scroll up")
	}
	if g.model.scrollOffset < maxOffset {
		lastIdx := len(result) - 1
		result[lastIdx] = result[lastIdx] + style.New().Faint(true).Render(" ↓ scroll down")
	}
	
	return strings.Join(result, "\n")
}

// formatMessage formats a single message into multiple lines
func (g *GeminiChatComponent) formatMessage(msg Message) []string {
	var lines []string
	
	// Format header
	var roleColor style.Color
	var rolePrefix string
	
	switch msg.Role {
	case "user":
		roleColor = style.Green
		rolePrefix = "You"
	case "assistant":
		roleColor = style.Blue
		rolePrefix = "Gemini"
	case "system":
		roleColor = style.Yellow
		rolePrefix = "System"
	}
	
	header := style.New().Foreground(roleColor).Bold(true).Render(rolePrefix + ":")
	if g.model.showTimestamp {
		timestamp := style.New().Faint(true).Render(msg.Timestamp.Format(" [15:04:05]"))
		header += timestamp
	}
	
	lines = append(lines, header)
	
	// Wrap and indent content
	contentLines := wrapText(msg.Content, 100)
	for _, line := range contentLines {
		lines = append(lines, "  " + line)
	}
	
	return lines
}

// addMessage adds a message to the chat
func (g *GeminiChatComponent) addMessage(role, content string) {
	g.model.messages = append(g.model.messages, Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
	// Auto-scroll to bottom on new message
	g.model.scrollOffset = 999999
}

// addSystemMessage adds a system message to the chat
func (g *GeminiChatComponent) addSystemMessage(content string) {
	g.model.messages = append(g.model.messages, Message{
		Role:      "system",
		Content:   content,
		Timestamp: time.Now(),
	})
	// Auto-scroll to bottom on new message
	g.model.scrollOffset = 999999
}


// wrapText wraps text to specified width while preserving newlines
func wrapText(text string, width int) []string {
	var result []string
	
	// First split by newlines to preserve them
	paragraphs := strings.Split(text, "\n")
	
	for _, paragraph := range paragraphs {
		if paragraph == "" {
			// Preserve empty lines
			result = append(result, "")
			continue
		}
		
		// For each paragraph, wrap long lines
		if len(paragraph) <= width {
			result = append(result, paragraph)
		} else {
			// Word wrap long paragraphs
			words := strings.Fields(paragraph)
			currentLine := ""
			
			for _, word := range words {
				if currentLine == "" {
					currentLine = word
				} else if len(currentLine)+1+len(word) <= width {
					currentLine += " " + word
				} else {
					result = append(result, currentLine)
					currentLine = word
				}
			}
			
			if currentLine != "" {
				result = append(result, currentLine)
			}
		}
	}
	
	return result
}

// Message types for Gemini communication
type GeminiConnectedMsg struct {
	Client *genai.Client
	Chat   *genai.ChatSession
}

type GeminiResponseMsg struct {
	Response string
}

type GeminiErrorMsg struct {
	Error error
}

func main() {
	// Create the program
	program := terminus.NewProgram(
		func() terminus.Component {
			return NewGeminiChatComponent()
		},
		terminus.WithStaticFiles(staticFiles, "static"),
		terminus.WithAddress(":8890"),
	)

	// Start the server
	if err := program.Start(); err != nil {
		log.Fatalf("Failed to start program: %v", err)
	}

	fmt.Println("Gemini Chat is running on http://localhost:8890")
	fmt.Println("Make sure GEMINI_API_KEY environment variable is set")
	fmt.Println("Press Ctrl+C to stop...")

	// Wait for the program to run
	program.Wait()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
	if err := program.Stop(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
}
