package main

import (
	"embed"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/yourusername/terminusgo/pkg/terminus"
	"github.com/yourusername/terminusgo/pkg/terminus/layout"
	"github.com/yourusername/terminusgo/pkg/terminus/widget"
)

//go:embed all:static/*
var staticFiles embed.FS

// Message represents a chat message
type Message struct {
	ID        int
	User      string
	Text      string
	Timestamp time.Time
	IsSystem  bool
}

// ChatModel represents the state of our chat application
type ChatModel struct {
	// Chat state
	messages      []Message
	messageList   *widget.List
	input         *widget.TextInput
	username      string
	nextMessageID int

	// UI state
	width        int
	height       int
	typingUsers  map[string]time.Time
	lastActivity time.Time

	// Settings
	showTimestamps bool
	use24Hour      bool
}

// ChatComponent is our main chat component
type ChatComponent struct {
	model ChatModel
}

// NewChatComponent creates a new chat component
func NewChatComponent() *ChatComponent {
	// Initialize widgets
	messageList := widget.NewList().
		SetShowCursor(false).
		SetWrap(false)

	input := widget.NewTextInput().
		SetPlaceholder("Type a message or /help for commands...").
		SetMaxLength(200)

	return &ChatComponent{
		model: ChatModel{
			messages:       make([]Message, 0),
			messageList:    messageList,
			input:          input,
			username:       "User",
			nextMessageID:  1,
			typingUsers:    make(map[string]time.Time),
			showTimestamps: true,
			use24Hour:      false,
			width:          80,
			height:         24,
		},
	}
}

// Init initializes the component
func (c *ChatComponent) Init() terminus.Cmd {
	// Focus the input widget
	c.model.input.Focus()

	// Add welcome message
	c.addSystemMessage("Welcome to TerminusGo Chat!")
	c.addSystemMessage("Type /help to see available commands")
	c.addSystemMessage(fmt.Sprintf("Your username is: %s", c.model.username))

	// Start simulated activity
	return terminus.Batch(
		c.simulateOtherUsers(),
		c.updateTypingIndicators(),
	)
}

// Update handles incoming messages and updates state
func (c *ChatComponent) Update(msg terminus.Msg) (terminus.Component, terminus.Cmd) {
	var cmds []terminus.Cmd

	switch msg := msg.(type) {
	case terminus.WindowSizeMsg:
		// Handle window resize
		c.model.width = msg.Width
		c.model.height = msg.Height
		c.updateLayout()

	case terminus.KeyMsg:
		// First, let the input widget handle the key
		var inputCmd terminus.Cmd
		_, inputCmd = c.model.input.Update(msg)
		cmds = append(cmds, inputCmd)

		// Handle special keys
		switch msg.Type {
		case terminus.KeyEnter:
			// Send message
			if text := strings.TrimSpace(c.model.input.Value()); text != "" {
				cmd := c.handleInput(text)
				cmds = append(cmds, cmd)
				c.model.input.Clear()
			}

		case terminus.KeyUp, terminus.KeyDown, terminus.KeyPgUp, terminus.KeyPgDown:
			// Let the message list handle scrolling
			_, listCmd := c.model.messageList.Update(msg)
			cmds = append(cmds, listCmd)
		}

		// Handle quit commands
		switch msg.String() {
		case "ctrl+c":
			return c, terminus.Quit
		}

	case simulatedMessageMsg:
		// Add simulated message from another user
		c.addMessage(msg.user, msg.text, false)
		cmds = append(cmds, c.simulateOtherUsers())

	case typingUpdateMsg:
		// Update typing indicators
		// Remove old typing indicators
		now := time.Now()
		for user, lastTyped := range c.model.typingUsers {
			if now.Sub(lastTyped) > 3*time.Second {
				delete(c.model.typingUsers, user)
			}
		}
		cmds = append(cmds, c.updateTypingIndicators())

	case simulatedTypingMsg:
		// Simulate another user typing
		c.model.typingUsers[msg.user] = time.Now()
	}

	return c, terminus.Batch(cmds...)
}

// View renders the chat interface
func (c *ChatComponent) View() string {
	// Calculate layout dimensions
	inputHeight := 3
	headerHeight := 3
	footerHeight := 2
	messageAreaHeight := c.model.height - inputHeight - headerHeight - footerHeight

	// Render components
	header := c.renderHeader()
	messages := c.renderMessages(messageAreaHeight)
	typing := c.renderTypingIndicator()
	input := c.renderInput()
	footer := c.renderFooter()

	// Combine all components
	return layout.Rows([]string{
		header,
		messages,
		typing,
		input,
		footer,
	}, 0)
}

// renderHeader renders the chat header
func (c *ChatComponent) renderHeader() string {
	titleStyle := terminus.NewStyle().Bold(true).Foreground(terminus.Cyan)
	infoStyle := terminus.NewStyle().Faint(true)

	title := titleStyle.Render("=== TerminusGo Chat Example ===")
	info := infoStyle.Render(fmt.Sprintf("Username: %s | Messages: %d | %s",
		c.model.username,
		len(c.model.messages),
		c.getOnlineUsers()))

	return layout.Center(title+"\n"+info, c.model.width, 2)
}

// renderMessages renders the message area
func (c *ChatComponent) renderMessages(height int) string {
	// Update message list dimensions
	c.model.messageList.SetSize(c.model.width, height)

	// Convert messages to list items
	items := make([]widget.ListItem, len(c.model.messages))
	for i, msg := range c.model.messages {
		items[i] = &messageListItem{message: msg, showTimestamp: c.model.showTimestamps, use24Hour: c.model.use24Hour}
	}
	c.model.messageList.SetItems(items)

	// Auto-scroll to bottom
	if len(items) > 0 {
		c.model.messageList.SetSelected(len(items) - 1)
	}

	return c.model.messageList.View()
}

// renderTypingIndicator renders the typing indicator
func (c *ChatComponent) renderTypingIndicator() string {
	if len(c.model.typingUsers) == 0 {
		return ""
	}

	typingStyle := terminus.NewStyle().Italic(true).Faint(true)

	users := make([]string, 0, len(c.model.typingUsers))
	for user := range c.model.typingUsers {
		users = append(users, user)
	}

	var text string
	if len(users) == 1 {
		text = fmt.Sprintf("%s is typing...", users[0])
	} else if len(users) == 2 {
		text = fmt.Sprintf("%s and %s are typing...", users[0], users[1])
	} else {
		text = fmt.Sprintf("%d users are typing...", len(users))
	}

	return typingStyle.Render(text)
}

// renderInput renders the input area
func (c *ChatComponent) renderInput() string {
	// Update input dimensions
	c.model.input.SetSize(c.model.width-4, 1)

	promptStyle := terminus.NewStyle().Foreground(terminus.Green)
	prompt := promptStyle.Render("> ")

	return prompt + c.model.input.View()
}

// renderFooter renders the footer
func (c *ChatComponent) renderFooter() string {
	helpStyle := terminus.NewStyle().Faint(true)
	return helpStyle.Render("Commands: /nick <name> | /clear | /time | /help | /quit | Ctrl+C to exit")
}

// handleInput processes user input
func (c *ChatComponent) handleInput(text string) terminus.Cmd {
	// Check for commands
	if strings.HasPrefix(text, "/") {
		parts := strings.Fields(text)
		if len(parts) == 0 {
			return nil
		}

		command := strings.ToLower(parts[0])
		args := parts[1:]

		switch command {
		case "/nick":
			if len(args) > 0 {
				oldName := c.model.username
				c.model.username = strings.Join(args, " ")
				c.addSystemMessage(fmt.Sprintf("%s changed their name to %s", oldName, c.model.username))
			} else {
				c.addSystemMessage("Usage: /nick <new name>")
			}

		case "/clear":
			c.model.messages = make([]Message, 0)
			c.model.nextMessageID = 1
			c.addSystemMessage("Chat cleared")

		case "/time":
			c.model.showTimestamps = !c.model.showTimestamps
			if c.model.showTimestamps {
				c.addSystemMessage("Timestamps enabled")
			} else {
				c.addSystemMessage("Timestamps disabled")
			}

		case "/24h":
			c.model.use24Hour = !c.model.use24Hour
			if c.model.use24Hour {
				c.addSystemMessage("24-hour time format enabled")
			} else {
				c.addSystemMessage("12-hour time format enabled")
			}

		case "/help":
			c.addSystemMessage("Available commands:")
			c.addSystemMessage("  /nick <name> - Change your username")
			c.addSystemMessage("  /clear - Clear all messages")
			c.addSystemMessage("  /time - Toggle timestamps")
			c.addSystemMessage("  /24h - Toggle 24-hour time format")
			c.addSystemMessage("  /quit - Exit the chat")
			c.addSystemMessage("  /help - Show this help message")

		case "/quit":
			return terminus.Quit

		default:
			c.addSystemMessage(fmt.Sprintf("Unknown command: %s", command))
		}

		return nil
	}

	// Regular message
	c.addMessage(c.model.username, text, false)
	c.model.lastActivity = time.Now()
	return nil
}

// addMessage adds a message to the chat
func (c *ChatComponent) addMessage(user, text string, isSystem bool) {
	msg := Message{
		ID:        c.model.nextMessageID,
		User:      user,
		Text:      text,
		Timestamp: time.Now(),
		IsSystem:  isSystem,
	}
	c.model.messages = append(c.model.messages, msg)
	c.model.nextMessageID++
}

// addSystemMessage adds a system message
func (c *ChatComponent) addSystemMessage(text string) {
	c.addMessage("System", text, true)
}

// updateLayout updates widget dimensions based on terminal size
func (c *ChatComponent) updateLayout() {
	// Message list gets most of the space
	messageHeight := c.model.height - 8 // Leave room for header, input, etc.
	if messageHeight < 5 {
		messageHeight = 5
	}
	c.model.messageList.SetSize(c.model.width, messageHeight)

	// Input gets full width minus prompt
	c.model.input.SetSize(c.model.width-4, 1)
}

// getOnlineUsers returns a string showing online users
func (c *ChatComponent) getOnlineUsers() string {
	// In a real app, this would track actual users
	// For demo, we'll simulate it
	activeUsers := 1 // Just us
	for user := range c.model.typingUsers {
		if user != c.model.username {
			activeUsers++
		}
	}

	if activeUsers == 1 {
		return "1 user online"
	}
	return fmt.Sprintf("%d users online", activeUsers)
}

// messageListItem implements widget.ListItem for chat messages
type messageListItem struct {
	message       Message
	showTimestamp bool
	use24Hour     bool
}

func (m *messageListItem) Render() string {
	userStyle := terminus.NewStyle().Bold(true).Foreground(terminus.Cyan)
	systemStyle := terminus.NewStyle().Italic(true).Foreground(terminus.Yellow)
	timeStyle := terminus.NewStyle().Faint(true)
	textStyle := terminus.NewStyle()

	var result strings.Builder

	// Timestamp
	if m.showTimestamp {
		var timeFormat string
		if m.use24Hour {
			timeFormat = "15:04"
		} else {
			timeFormat = "3:04 PM"
		}
		timestamp := m.message.Timestamp.Format(timeFormat)
		result.WriteString(timeStyle.Render("[" + timestamp + "] "))
	}

	// User and message
	if m.message.IsSystem {
		result.WriteString(systemStyle.Render("*** " + m.message.Text + " ***"))
	} else {
		result.WriteString(userStyle.Render(m.message.User + ": "))
		result.WriteString(textStyle.Render(m.message.Text))
	}

	return result.String()
}

func (m *messageListItem) String() string {
	return fmt.Sprintf("%s: %s", m.message.User, m.message.Text)
}

// Custom message types for async operations
type simulatedMessageMsg struct {
	user string
	text string
}

type simulatedTypingMsg struct {
	user string
}

type typingUpdateMsg struct{}

// simulateOtherUsers creates random messages from simulated users
func (c *ChatComponent) simulateOtherUsers() terminus.Cmd {
	return func() terminus.Msg {
		// Wait a random time between 5-15 seconds
		delay := time.Duration(5+rand.Intn(10)) * time.Second
		time.Sleep(delay)

		// List of simulated users and messages
		users := []string{"Alice", "Bob", "Charlie", "Diana", "Eve"}
		messages := []string{
			"Hey everyone!",
			"How's it going?",
			"Did you see the latest update?",
			"That's interesting!",
			"I agree with that",
			"Good point!",
			"Thanks for sharing",
			"Anyone here?",
			"What do you think?",
			"Nice!",
			"Cool feature!",
			"Testing the chat...",
			"This is pretty neat",
			"Love the real-time updates",
		}

		// Pick random user and message
		user := users[rand.Intn(len(users))]
		text := messages[rand.Intn(len(messages))]

		// Simulate typing first
		go func() {
			// Show typing indicator
			time.Sleep(1 * time.Second)
			// Note: In a real implementation, we'd send this through the message queue
		}()

		return simulatedMessageMsg{user: user, text: text}
	}
}

// updateTypingIndicators periodically updates typing indicators
func (c *ChatComponent) updateTypingIndicators() terminus.Cmd {
	return terminus.Tick(1*time.Second, func(t time.Time) terminus.Msg {
		return typingUpdateMsg{}
	})
}

func init() {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())
}

func main() {
	// Create and configure the program
	program := terminus.NewProgram(
		func() terminus.Component {
			return NewChatComponent()
		},
		terminus.WithStaticFiles(staticFiles, "static"),
		terminus.WithAddress(":8890"),
	)

	// Start the server
	if err := program.Start(); err != nil {
		log.Fatalf("Failed to start program: %v", err)
	}

	fmt.Println("TerminusGo Chat Example is running on http://localhost:8890")
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
