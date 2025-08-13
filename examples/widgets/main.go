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
	"github.com/yourusername/terminusgo/pkg/terminus/widget"
)

//go:embed all:static/*
var staticFiles embed.FS

// WidgetShowcase demonstrates all the widgets in the library
type WidgetShowcase struct {
	// Widget selection
	currentView    ViewType
	focusedSection int

	// TextInput demo
	textInput      *widget.TextInput
	textInputValue string

	// List demo
	list         *widget.List
	listSelected string
	filterInput  *widget.TextInput

	// Table demo
	table         *widget.Table
	tableSelected string

	// Spinner demo
	spinner      *widget.Spinner
	spinnerStyle widget.SpinnerStyle
	isLoading    bool

	// Status
	statusMessage string
	statusStyle   terminus.Style
}

type ViewType int

const (
	ViewTextInput ViewType = iota
	ViewList
	ViewTable
	ViewSpinner
	ViewAll
)

func NewWidgetShowcase() *WidgetShowcase {
	showcase := &WidgetShowcase{
		currentView:  ViewAll,
		statusStyle:  terminus.NewStyle().Foreground(terminus.Green),
	}

	// Initialize TextInput
	showcase.textInput = widget.NewTextInput().
		SetPlaceholder("Type something here...").
		SetMaxLength(50).
		SetStyle(terminus.NewStyle().Foreground(terminus.White)).
		SetFocusStyle(terminus.NewStyle().Foreground(terminus.Cyan).Underline(true)).
		SetOnSubmit(func(value string) terminus.Cmd {
			showcase.textInputValue = value
			showcase.statusMessage = fmt.Sprintf("TextInput submitted: %s", value)
			showcase.statusStyle = terminus.NewStyle().Foreground(terminus.Green)
			return nil
		}).
		SetOnChange(func(value string) terminus.Cmd {
			showcase.textInputValue = value
			if len(value) > 40 {
				showcase.statusMessage = "Approaching max length!"
				showcase.statusStyle = terminus.NewStyle().Foreground(terminus.Yellow)
			}
			return nil
		})

	// Initialize List
	showcase.list = widget.NewList().
		SetStringItems([]string{
			"🍎 Apple",
			"🍌 Banana",
			"🍒 Cherry",
			"🍇 Grapes",
			"🥝 Kiwi",
			"🍋 Lemon",
			"🥭 Mango",
			"🍊 Orange",
			"🍑 Peach",
			"🍓 Strawberry",
			"🍉 Watermelon",
		}).
		SetShowCursor(true).
		SetCursorChar("→ ").
		SetStyle(terminus.NewStyle()).
		SetSelectedStyle(terminus.NewStyle().Bold(true).Foreground(terminus.Yellow)).
		SetOnSelect(func(idx int, item widget.ListItem) terminus.Cmd {
			showcase.listSelected = item.String()
			showcase.statusMessage = fmt.Sprintf("Selected: %s (index %d)", item.String(), idx)
			showcase.statusStyle = terminus.NewStyle().Foreground(terminus.Green)
			return nil
		}).
		SetOnChange(func(idx int, item widget.ListItem) terminus.Cmd {
			if item != nil {
				showcase.statusMessage = fmt.Sprintf("Hovering: %s", item.String())
				showcase.statusStyle = terminus.NewStyle().Foreground(terminus.Blue)
			}
			return nil
		})

	// Initialize filter input for list
	showcase.filterInput = widget.NewTextInput().
		SetPlaceholder("Filter fruits...").
		SetMaxLength(20).
		SetOnChange(func(value string) terminus.Cmd {
			showcase.list.SetFilter(value)
			return nil
		})

	// Initialize Table
	headers := []string{"ID", "Name", "Price", "Stock"}
	data := [][]string{
		{"001", "Laptop", "$999", "12"},
		{"002", "Mouse", "$29", "45"},
		{"003", "Keyboard", "$79", "23"},
		{"004", "Monitor", "$299", "8"},
		{"005", "Headphones", "$149", "34"},
		{"006", "Webcam", "$89", "17"},
		{"007", "USB Hub", "$39", "56"},
		{"008", "Desk Lamp", "$49", "29"},
	}
	
	showcase.table = widget.NewTable().
		SetStringData(headers, data).
		SetShowHeader(true).
		SetShowRowNumbers(true).
		SetCellSelection(true).
		SetStyle(terminus.NewStyle()).
		SetHeaderStyle(terminus.NewStyle().Bold(true).Foreground(terminus.Cyan)).
		SetSelectedStyle(terminus.NewStyle().Reverse(true)).
		SetOnSelect(func(row, col int, cell widget.TableCell) terminus.Cmd {
			if cell != nil {
				showcase.tableSelected = cell.String()
				showcase.statusMessage = fmt.Sprintf("Table cell selected: %s (row %d, col %d)", cell.String(), row, col)
				showcase.statusStyle = terminus.NewStyle().Foreground(terminus.Green)
			}
			return nil
		})

	// Adjust column widths
	columns := []widget.TableColumn{
		{Title: "ID", Width: 5, Align: widget.AlignCenter, Sortable: true},
		{Title: "Name", Width: 15, Align: widget.AlignLeft, Sortable: true},
		{Title: "Price", Width: 10, Align: widget.AlignRight, Sortable: true},
		{Title: "Stock", Width: 8, Align: widget.AlignCenter, Sortable: true},
	}
	showcase.table.SetColumns(columns)

	// Initialize Spinner
	showcase.spinner = widget.NewSpinner().
		SetSpinnerStyle(widget.SpinnerDots).
		SetText("Loading data...").
		SetTextPosition(widget.TextRight).
		SetSpeed(100 * time.Millisecond).
		SetSpinnerColor(terminus.NewStyle().Foreground(terminus.Cyan))

	// Set initial sizes
	showcase.textInput.SetSize(50, 1)
	showcase.filterInput.SetSize(30, 1)
	showcase.list.SetSize(40, 8)
	showcase.table.SetSize(50, 10)

	return showcase
}

func (w *WidgetShowcase) Init() terminus.Cmd {
	return nil
}

func (w *WidgetShowcase) Update(msg terminus.Msg) (terminus.Component, terminus.Cmd) {
	switch msg := msg.(type) {
	case terminus.KeyMsg:
		switch msg.Type {
		case terminus.KeyCtrlC:
			return w, terminus.Quit
		
		case terminus.KeyRunes:
			if len(msg.Runes) > 0 {
				switch msg.Runes[0] {
				case 'q', 'Q':
					return w, terminus.Quit
				case '1':
					w.currentView = ViewTextInput
					w.statusMessage = "Switched to TextInput view"
				case '2':
					w.currentView = ViewList
					w.statusMessage = "Switched to List view"
				case '3':
					w.currentView = ViewTable
					w.statusMessage = "Switched to Table view"
				case '4':
					w.currentView = ViewSpinner
					w.statusMessage = "Switched to Spinner view"
				case '5':
					w.currentView = ViewAll
					w.statusMessage = "Switched to All widgets view"
				case 'l', 'L':
					// Toggle loading spinner
					if w.isLoading {
						w.spinner.Stop()
						w.isLoading = false
						w.statusMessage = "Spinner stopped"
					} else {
						w.spinner.Start()
						w.isLoading = true
						w.statusMessage = "Spinner started"
					}
				case 'n', 'N':
					// Next spinner style
					w.spinnerStyle = (w.spinnerStyle + 1) % 8
					w.spinner.SetSpinnerStyle(w.spinnerStyle)
					styles := []string{"Dots", "Line", "Circle", "Arrow", "Bounce", "Pulse", "Clock", "Braille"}
					w.statusMessage = fmt.Sprintf("Spinner style: %s", styles[w.spinnerStyle])
				case 's', 'S':
					// Sort table by current column (when in table view)
					if w.currentView == ViewTable {
						return w.table.Update(msg)
					}
				}
			}

		case terminus.KeyTab:
			// Tab between sections in "All" view
			if w.currentView == ViewAll {
				w.focusedSection = (w.focusedSection + 1) % 4
				// Clear focus from all widgets
				w.textInput.Blur()
				w.filterInput.Blur()
				w.list.Blur()
				w.table.Blur()
				// Focus the current section
				switch w.focusedSection {
				case 0:
					w.textInput.Focus()
				case 1:
					w.filterInput.Focus()
				case 2:
					w.list.Focus()
				case 3:
					w.table.Focus()
				}
				w.statusMessage = fmt.Sprintf("Focused section %d", w.focusedSection+1)
				return w, nil
			}
		}

		// Forward to appropriate widget based on view
		switch w.currentView {
		case ViewTextInput:
			return w.textInput.Update(msg)
		case ViewList:
			// Check if filter input is focused
			if w.filterInput.Focused() {
				return w.filterInput.Update(msg)
			}
			return w.list.Update(msg)
		case ViewTable:
			return w.table.Update(msg)
		case ViewSpinner:
			return w.spinner.Update(msg)
		case ViewAll:
			// Forward to focused widget
			switch w.focusedSection {
			case 0:
				return w.textInput.Update(msg)
			case 1:
				newFilter, cmd := w.filterInput.Update(msg)
				w.filterInput = newFilter.(*widget.TextInput)
				return w, cmd
			case 2:
				return w.list.Update(msg)
			case 3:
				return w.table.Update(msg)
			}
		}

	case widget.SpinnerTickMsg:
		// Forward spinner tick messages
		newSpinner, cmd := w.spinner.Update(msg)
		w.spinner = newSpinner.(*widget.Spinner)
		return w, cmd
	}

	return w, nil
}

func (w *WidgetShowcase) View() string {
	var result strings.Builder

	// Title
	titleStyle := terminus.NewStyle().Bold(true).Foreground(terminus.Cyan)
	result.WriteString(titleStyle.Render("TerminusGo Widget Showcase"))
	result.WriteString("\n\n")

	// Navigation
	navStyle := terminus.NewStyle().Faint(true)
	result.WriteString(navStyle.Render("Press 1-5 to switch views | Tab to navigate | 'q' to quit"))
	result.WriteString("\n\n")

	// Show current view
	switch w.currentView {
	case ViewTextInput:
		w.renderTextInputView(&result)
	case ViewList:
		w.renderListView(&result)
	case ViewTable:
		w.renderTableView(&result)
	case ViewSpinner:
		w.renderSpinnerView(&result)
	case ViewAll:
		w.renderAllView(&result)
	}

	// Status message
	if w.statusMessage != "" {
		result.WriteString("\n\n")
		result.WriteString(w.statusStyle.Render("Status: " + w.statusMessage))
	}

	return result.String()
}

func (w *WidgetShowcase) renderTextInputView(result *strings.Builder) {
	headerStyle := terminus.NewStyle().Bold(true).Underline(true)
	result.WriteString(headerStyle.Render("TextInput Widget"))
	result.WriteString("\n\n")

	result.WriteString("Features:\n")
	result.WriteString("• Placeholder text\n")
	result.WriteString("• Max length (50 chars)\n")
	result.WriteString("• Real-time onChange callbacks\n")
	result.WriteString("• Submit with Enter\n")
	result.WriteString("• Custom styling\n\n")

	result.WriteString("Try it:\n")
	result.WriteString(w.textInput.View())
	result.WriteString("\n\n")
	
	if w.textInputValue != "" {
		result.WriteString(fmt.Sprintf("Current value: %s", w.textInputValue))
	}
}

func (w *WidgetShowcase) renderListView(result *strings.Builder) {
	headerStyle := terminus.NewStyle().Bold(true).Underline(true)
	result.WriteString(headerStyle.Render("List Widget"))
	result.WriteString("\n\n")

	result.WriteString("Features:\n")
	result.WriteString("• Scrollable list with keyboard navigation\n")
	result.WriteString("• Real-time filtering\n")
	result.WriteString("• Custom item rendering\n")
	result.WriteString("• Selection callbacks\n")
	result.WriteString("• Customizable cursor\n\n")

	result.WriteString("Filter: ")
	result.WriteString(w.filterInput.View())
	result.WriteString("\n\n")

	result.WriteString("Fruits (↑↓ to navigate, Enter to select):\n")
	result.WriteString(w.list.View())
	
	if w.listSelected != "" {
		result.WriteString("\n\nSelected: " + w.listSelected)
	}
	
	result.WriteString(fmt.Sprintf("\n\nShowing %d of %d items", w.list.FilteredLen(), w.list.Len()))
}

func (w *WidgetShowcase) renderTableView(result *strings.Builder) {
	headerStyle := terminus.NewStyle().Bold(true).Underline(true)
	result.WriteString(headerStyle.Render("Table Widget"))
	result.WriteString("\n\n")

	result.WriteString("Features:\n")
	result.WriteString("• Sortable columns (press 's' on a column)\n")
	result.WriteString("• Cell selection mode\n")
	result.WriteString("• Row numbers\n")
	result.WriteString("• Custom alignment\n")
	result.WriteString("• Arrow keys for navigation\n\n")

	result.WriteString(w.table.View())
	
	if w.tableSelected != "" {
		result.WriteString("\n\nSelected cell: " + w.tableSelected)
	}
}

func (w *WidgetShowcase) renderSpinnerView(result *strings.Builder) {
	headerStyle := terminus.NewStyle().Bold(true).Underline(true)
	result.WriteString(headerStyle.Render("Spinner Widget"))
	result.WriteString("\n\n")

	result.WriteString("Features:\n")
	result.WriteString("• Multiple animation styles\n")
	result.WriteString("• Customizable speed\n")
	result.WriteString("• Loading text\n")
	result.WriteString("• Start/stop control\n\n")

	result.WriteString("Controls:\n")
	result.WriteString("• Press 'l' to start/stop loading\n")
	result.WriteString("• Press 'n' to change spinner style\n\n")

	result.WriteString("Demo:\n")
	result.WriteString(w.spinner.View())
	
	if w.isLoading {
		result.WriteString(fmt.Sprintf("\n\nRunning for: %s", w.spinner.ElapsedTime().Round(time.Second)))
	}
}

func (w *WidgetShowcase) renderAllView(result *strings.Builder) {
	headerStyle := terminus.NewStyle().Bold(true).Underline(true)
	sectionStyle := terminus.NewStyle().Bold(true)
	focusIndicator := terminus.NewStyle().Foreground(terminus.Yellow).Render("→ ")
	
	result.WriteString(headerStyle.Render("All Widgets (Tab to switch focus)"))
	result.WriteString("\n\n")

	// TextInput
	if w.focusedSection == 0 {
		result.WriteString(focusIndicator)
	}
	result.WriteString(sectionStyle.Render("TextInput:"))
	result.WriteString("\n")
	result.WriteString(w.textInput.View())
	result.WriteString("\n\n")

	// List with filter
	if w.focusedSection == 1 {
		result.WriteString(focusIndicator)
	}
	result.WriteString(sectionStyle.Render("List Filter:"))
	result.WriteString("\n")
	result.WriteString(w.filterInput.View())
	result.WriteString("\n\n")

	if w.focusedSection == 2 {
		result.WriteString(focusIndicator)
	}
	result.WriteString(sectionStyle.Render("List:"))
	result.WriteString("\n")
	result.WriteString(w.list.View())
	result.WriteString("\n\n")

	// Table
	if w.focusedSection == 3 {
		result.WriteString(focusIndicator)
	}
	result.WriteString(sectionStyle.Render("Table:"))
	result.WriteString("\n")
	result.WriteString(w.table.View())
	result.WriteString("\n\n")

	// Spinner
	result.WriteString(sectionStyle.Render("Spinner:"))
	result.WriteString(" ")
	result.WriteString(w.spinner.View())
}

func main() {
	// Component factory
	factory := func() terminus.Component {
		return NewWidgetShowcase()
	}

	// Create program with static files
	program := terminus.NewProgram(
		factory,
		terminus.WithStaticFiles(staticFiles, "static"),
		terminus.WithAddress(":8890"), // Use different port
	)

	// Start the program
	if err := program.Start(); err != nil {
		log.Fatalf("Failed to start program: %v", err)
	}

	fmt.Println("Widget Showcase is running on http://localhost:8890")
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