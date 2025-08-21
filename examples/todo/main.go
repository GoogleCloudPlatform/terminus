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
	"strings"
	"syscall"
	"time"

	"github.com/yourusername/terminusgo/pkg/terminus"
	"github.com/yourusername/terminusgo/pkg/terminus/layout"
	"github.com/yourusername/terminusgo/pkg/terminus/widget"
)

//go:embed all:static/*
var staticFiles embed.FS

// TodoItem represents a single todo item
type TodoItem struct {
	ID          int
	Text        string
	Completed   bool
	CreatedAt   time.Time
	CompletedAt *time.Time
}

// Render implements widget.ListItem interface
func (t *TodoItem) Render() string {
	checkbox := "[ ]"
	textStyle := terminus.NewStyle()

	if t.Completed {
		checkbox = "[✓]"
		textStyle = textStyle.Faint(true)
	}

	checkboxStyle := terminus.NewStyle().Foreground(terminus.Green)
	if !t.Completed {
		checkboxStyle = terminus.NewStyle().Foreground(terminus.Yellow)
	}

	return checkboxStyle.Render(checkbox) + " " + textStyle.Render(t.Text)
}

// String implements widget.ListItem interface
func (t *TodoItem) String() string {
	return t.Text
}

// FilterMode represents different filtering options
type FilterMode int

const (
	FilterAll FilterMode = iota
	FilterActive
	FilterCompleted
)

// TodoModel represents the state of the todo application
type TodoModel struct {
	todos      []*TodoItem
	nextID     int
	filterMode FilterMode
	focusIndex int
}

// TodoComponent is the main todo list component
type TodoComponent struct {
	model        TodoModel
	todoList     *widget.List
	textInput    *widget.TextInput
	focusManager *widget.FocusManager
	width        int
	height       int
}

// NewTodoComponent creates a new todo component
func NewTodoComponent() *TodoComponent {
	// Create widgets
	todoList := widget.NewList().
		SetShowCursor(true).
		SetCursorChar("▶ ").
		SetSelectedChar("  ").
		SetUnselectedChar("  ").
		SetWrap(true).
		SetCursorStyle(terminus.NewStyle().Foreground(terminus.Cyan)).
		SetSelectedStyle(terminus.NewStyle().Background(terminus.ANSI256(237)))
	todoList.SetSize(60, 15)

	textInput := widget.NewTextInput().
		SetPlaceholder("What needs to be done?").
		SetMaxLength(100).
		SetFocusStyle(terminus.NewStyle().Underline(true)).
		SetPlaceholderStyle(terminus.NewStyle().Faint(true))
	textInput.SetSize(60, 1)

	// Create focus manager
	focusManager := widget.NewFocusManager(textInput, todoList)

	component := &TodoComponent{
		model: TodoModel{
			todos:      make([]*TodoItem, 0),
			nextID:     1,
			filterMode: FilterAll,
			focusIndex: 0,
		},
		todoList:     todoList,
		textInput:    textInput,
		focusManager: focusManager,
		width:        80,
		height:       24,
	}

	// Set up event handlers
	textInput.SetOnSubmit(func(value string) terminus.Cmd {
		if strings.TrimSpace(value) != "" {
			component.addTodo(value)
			textInput.Clear()
			component.updateList()
		}
		return nil
	})

	todoList.SetOnSelect(func(index int, item widget.ListItem) terminus.Cmd {
		if todoItem, ok := item.(*TodoItem); ok {
			component.toggleTodo(todoItem.ID)
			component.updateList()
		}
		return nil
	})

	// Add some sample todos
	component.addTodo("Learn TerminusGo widget system")
	component.addTodo("Build an awesome todo app")
	component.addTodo("Master the MVU pattern")
	component.model.todos[0].Completed = true
	component.model.todos[0].CompletedAt = &time.Time{}
	*component.model.todos[0].CompletedAt = time.Now()

	component.updateList()

	return component
}

// addTodo adds a new todo item
func (c *TodoComponent) addTodo(text string) {
	todo := &TodoItem{
		ID:        c.model.nextID,
		Text:      text,
		Completed: false,
		CreatedAt: time.Now(),
	}
	c.model.todos = append(c.model.todos, todo)
	c.model.nextID++
}

// toggleTodo toggles the completion status of a todo
func (c *TodoComponent) toggleTodo(id int) {
	for _, todo := range c.model.todos {
		if todo.ID == id {
			todo.Completed = !todo.Completed
			if todo.Completed {
				now := time.Now()
				todo.CompletedAt = &now
			} else {
				todo.CompletedAt = nil
			}
			break
		}
	}
}

// deleteTodo removes a todo by ID
func (c *TodoComponent) deleteTodo(id int) {
	filtered := make([]*TodoItem, 0, len(c.model.todos))
	for _, todo := range c.model.todos {
		if todo.ID != id {
			filtered = append(filtered, todo)
		}
	}
	c.model.todos = filtered
}

// clearCompleted removes all completed todos
func (c *TodoComponent) clearCompleted() {
	filtered := make([]*TodoItem, 0, len(c.model.todos))
	for _, todo := range c.model.todos {
		if !todo.Completed {
			filtered = append(filtered, todo)
		}
	}
	c.model.todos = filtered
}

// getFilteredTodos returns todos based on current filter
func (c *TodoComponent) getFilteredTodos() []widget.ListItem {
	items := make([]widget.ListItem, 0)

	for _, todo := range c.model.todos {
		switch c.model.filterMode {
		case FilterAll:
			items = append(items, todo)
		case FilterActive:
			if !todo.Completed {
				items = append(items, todo)
			}
		case FilterCompleted:
			if todo.Completed {
				items = append(items, todo)
			}
		}
	}

	return items
}

// updateList updates the list widget with current todos
func (c *TodoComponent) updateList() {
	c.todoList.SetItems(c.getFilteredTodos())
}

// Init implements terminus.Component
func (c *TodoComponent) Init() terminus.Cmd {
	return nil
}

// Update implements terminus.Component
func (c *TodoComponent) Update(msg terminus.Msg) (terminus.Component, terminus.Cmd) {
	switch msg := msg.(type) {
	case terminus.KeyMsg:
		// Check for global shortcuts first
		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			return c, terminus.Quit
		case "ctrl+a":
			// Toggle all todos
			allCompleted := true
			for _, todo := range c.model.todos {
				if !todo.Completed {
					allCompleted = false
					break
				}
			}
			for _, todo := range c.model.todos {
				todo.Completed = !allCompleted
				if todo.Completed {
					now := time.Now()
					todo.CompletedAt = &now
				} else {
					todo.CompletedAt = nil
				}
			}
			c.updateList()
			return c, nil
		case "ctrl+k":
			// Clear completed todos
			c.clearCompleted()
			c.updateList()
			return c, nil
		case "1":
			// Show all todos
			c.model.filterMode = FilterAll
			c.updateList()
			return c, nil
		case "2":
			// Show active todos
			c.model.filterMode = FilterActive
			c.updateList()
			return c, nil
		case "3":
			// Show completed todos
			c.model.filterMode = FilterCompleted
			c.updateList()
			return c, nil
		}

		// Handle tab navigation
		if c.focusManager.HandleKey(msg) {
			return c, nil
		}

		// Forward to focused widget
		if c.textInput.Focused() {
			_, cmd := c.textInput.Update(msg)
			return c, cmd
		} else if c.todoList.Focused() {
			// Handle delete key for todo list
			if msg.Type == terminus.KeyDelete || msg.String() == "d" {
				if item := c.todoList.SelectedItem(); item != nil {
					if todoItem, ok := item.(*TodoItem); ok {
						c.deleteTodo(todoItem.ID)
						c.updateList()
					}
				}
				return c, nil
			}

			_, cmd := c.todoList.Update(msg)
			return c, cmd
		}

	case terminus.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height
		// Update widget sizes
		listHeight := c.height - 12 // Leave room for header, input, and footer
		if listHeight < 5 {
			listHeight = 5
		}
		c.todoList.SetSize(c.width-20, listHeight)
		c.textInput.SetSize(c.width-20, 1)
		return c, nil
	}

	return c, nil
}

// View implements terminus.Component
func (c *TodoComponent) View() string {
	// Styles
	titleStyle := terminus.NewStyle().Bold(true).Foreground(terminus.Cyan)
	headerStyle := terminus.NewStyle().Foreground(terminus.Magenta)
	footerStyle := terminus.NewStyle().Faint(true)
	statsStyle := terminus.NewStyle().Foreground(terminus.Yellow)
	filterStyle := terminus.NewStyle().Foreground(terminus.Green)
	selectedFilterStyle := terminus.NewStyle().Bold(true).Underline(true).Foreground(terminus.Green)

	// Build the view
	var view strings.Builder

	// Title
	title := titleStyle.Render("╔═══════════════════════════════════════╗\n") +
		titleStyle.Render("║        TerminusGo Todo List          ║\n") +
		titleStyle.Render("╚═══════════════════════════════════════╝")
	view.WriteString(layout.Center(title, c.width, 3))
	view.WriteString("\n\n")

	// Input field
	inputSection := headerStyle.Render("Add New Todo:") + "\n" +
		c.textInput.View()
	view.WriteString(layout.Margin(inputSection, 0, 10, 1, 10))
	view.WriteString("\n")

	// Filter tabs
	filters := []string{"All", "Active", "Completed"}
	var filterTabs []string
	for i, filter := range filters {
		s := filterStyle
		if FilterMode(i) == c.model.filterMode {
			s = selectedFilterStyle
		}
		filterTabs = append(filterTabs, fmt.Sprintf("[%d] %s", i+1, s.Render(filter)))
	}
	filterLine := strings.Join(filterTabs, "  ")
	view.WriteString(layout.Margin(filterLine, 0, 10, 1, 10))
	view.WriteString("\n")

	// Todo list
	listView := c.todoList.View()
	view.WriteString(layout.Margin(listView, 0, 10, 1, 10))
	view.WriteString("\n")

	// Statistics
	totalCount := len(c.model.todos)
	activeCount := 0
	completedCount := 0
	for _, todo := range c.model.todos {
		if todo.Completed {
			completedCount++
		} else {
			activeCount++
		}
	}

	stats := fmt.Sprintf("%s: %d | %s: %d | %s: %d",
		statsStyle.Render("Total"), totalCount,
		statsStyle.Render("Active"), activeCount,
		statsStyle.Render("Completed"), completedCount)
	view.WriteString(layout.Center(stats, c.width, 1))
	view.WriteString("\n")

	// Instructions
	instructions := []string{
		"Tab: Switch focus | Enter: Add/Toggle todo | Delete/d: Remove todo",
		"Ctrl+A: Toggle all | Ctrl+K: Clear completed | Ctrl+C: Quit",
	}
	for _, instruction := range instructions {
		view.WriteString(layout.Center(footerStyle.Render(instruction), c.width, 1))
		view.WriteString("\n")
	}

	return view.String()
}

func main() {
	// Create and configure the TerminusGo program
	program := terminus.NewProgram(
		func() terminus.Component {
			return NewTodoComponent()
		},
		terminus.WithStaticFiles(staticFiles, "static"),
		terminus.WithAddress(":8890"),
	)

	// Start the server
	if err := program.Start(); err != nil {
		log.Fatalf("Failed to start program: %v", err)
	}

	fmt.Println("TerminusGo Todo List is running on http://localhost:8890")
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
