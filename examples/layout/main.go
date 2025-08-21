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

	"github.com/yourusername/terminusgo/pkg/terminus"
	"github.com/yourusername/terminusgo/pkg/terminus/layout"
	"github.com/yourusername/terminusgo/pkg/terminus/widget"
)

//go:embed all:static/*
var staticFiles embed.FS

// LayoutDemo demonstrates various layout utilities
type LayoutDemo struct {
	currentExample int
	examples       []Example
	focusManager   *widget.FocusManager
}

type Example struct {
	name string
	view func() string
}

func NewLayoutDemo() *LayoutDemo {
	demo := &LayoutDemo{
		currentExample: 0,
		focusManager:   widget.NewFocusManager(),
	}

	// Define examples
	demo.examples = []Example{
		{"Box Styles", demo.boxStylesExample},
		{"Layout Columns", demo.columnsExample},
		{"Layout Grid", demo.gridExample},
		{"Alignment", demo.alignmentExample},
		{"Complex Layout", demo.complexExample},
	}

	return demo
}

func (d *LayoutDemo) Init() terminus.Cmd {
	return nil
}

func (d *LayoutDemo) Update(msg terminus.Msg) (terminus.Component, terminus.Cmd) {
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
				case 'n', 'N':
					d.currentExample = (d.currentExample + 1) % len(d.examples)
				case 'p', 'P':
					d.currentExample = (d.currentExample - 1 + len(d.examples)) % len(d.examples)
				}
			}
		}
	}
	return d, nil
}

func (d *LayoutDemo) View() string {
	var result strings.Builder

	// Title
	titleStyle := terminus.NewStyle().Bold(true).Foreground(terminus.Cyan)
	result.WriteString(titleStyle.Render("Layout Utilities Demo"))
	result.WriteString("\n\n")

	// Navigation
	navStyle := terminus.NewStyle().Faint(true)
	result.WriteString(navStyle.Render("Navigation: [N]ext, [P]revious, [Q]uit"))
	result.WriteString("\n")
	result.WriteString(navStyle.Render(fmt.Sprintf("Example %d/%d: %s",
		d.currentExample+1, len(d.examples), d.examples[d.currentExample].name)))
	result.WriteString("\n\n")

	// Separator
	result.WriteString(layout.HorizontalLine(60, layout.BoxStyleSingle))
	result.WriteString("\n\n")

	// Current example
	result.WriteString(d.examples[d.currentExample].view())

	return result.String()
}

func (d *LayoutDemo) boxStylesExample() string {
	var result strings.Builder

	// Show different box styles
	styles := []struct {
		name  string
		style layout.BoxStyle
	}{
		{"Single", layout.BoxStyleSingle},
		{"Double", layout.BoxStyleDouble},
		{"Rounded", layout.BoxStyleRounded},
		{"Bold", layout.BoxStyleBold},
		{"ASCII", layout.BoxStyleASCII},
	}

	// Create boxes for each style
	boxes := make([]string, len(styles))
	for i, s := range styles {
		content := fmt.Sprintf("%s Style\nBox Drawing", s.name)
		box := layout.NewBox(content).
			WithStyle(s.style).
			WithTitle(s.name).
			WithUniformPadding(1).
			WithWidth(15).
			WithHeight(4)
		boxes[i] = box.Render()
	}

	// Arrange in columns
	result.WriteString(layout.Columns(boxes[:3], []int{20, 20, 20}, 2))
	result.WriteString("\n\n")
	result.WriteString(layout.Columns(boxes[3:], []int{20, 20}, 2))

	return result.String()
}

func (d *LayoutDemo) columnsExample() string {
	var result strings.Builder

	// Create some content
	col1 := `Column 1
This is the first
column with multiple
lines of text.`

	col2 := `Column 2
Here's another
column with
different content.`

	col3 := `Column 3
The third column
can have its own
width too!`

	// Show columns with different configurations
	result.WriteString("Equal width columns:\n")
	result.WriteString(layout.Columns([]string{col1, col2, col3}, []int{20, 20, 20}, 2))
	result.WriteString("\n\n")

	result.WriteString("Different width columns:\n")
	result.WriteString(layout.Columns([]string{col1, col2, col3}, []int{15, 25, 20}, 1))
	result.WriteString("\n\n")

	// Columns with boxes
	box1 := layout.DrawBox("Box in\nColumn 1", layout.BoxStyleSingle)
	box2 := layout.DrawBox("Box in\nColumn 2", layout.BoxStyleDouble)

	result.WriteString("Columns with boxes:\n")
	result.WriteString(layout.Columns([]string{box1, box2}, []int{15, 15}, 3))

	return result.String()
}

func (d *LayoutDemo) gridExample() string {
	var result strings.Builder

	// Create a data grid
	result.WriteString("Data Grid Example:\n\n")

	grid := layout.NewGrid(4, 3).SetGap(1)

	// Headers
	headers := []string{"ID", "Name", "Status", "Progress"}
	headerStyle := terminus.NewStyle().Bold(true).Underline(true)
	for i, h := range headers {
		grid.SetCell(i, 0, headerStyle.Render(h))
	}

	// Data rows
	data := [][]string{
		{"001", "Task Alpha", "Active", "75%"},
		{"002", "Task Beta", "Complete", "100%"},
	}

	statusStyles := map[string]terminus.Style{
		"Active":   terminus.NewStyle().Foreground(terminus.Yellow),
		"Complete": terminus.NewStyle().Foreground(terminus.Green),
	}

	for row, rowData := range data {
		for col, cell := range rowData {
			if col == 2 { // Status column
				if s, ok := statusStyles[cell]; ok {
					cell = s.Render(cell)
				}
			}
			grid.SetCell(col, row+1, cell)
		}
	}

	// Set column widths
	grid.SetColumnWidth(0, 5)
	grid.SetColumnWidth(1, 15)
	grid.SetColumnWidth(2, 10)
	grid.SetColumnWidth(3, 10)

	result.WriteString(grid.Render())
	result.WriteString("\n\n")

	// Show a grid with multiline cells
	result.WriteString("Grid with multiline cells:\n\n")

	multiGrid := layout.NewGrid(2, 2).SetGap(2)
	multiGrid.SetCell(0, 0, "Multi\nLine\nCell")
	multiGrid.SetCell(1, 0, "Single")
	multiGrid.SetCell(0, 1, "Another")
	multiGrid.SetCell(1, 1, "Multi\nLine")

	result.WriteString(multiGrid.Render())

	return result.String()
}

func (d *LayoutDemo) alignmentExample() string {
	var result strings.Builder

	// Show different alignments
	width, height := 20, 5
	content := "Hello"

	alignments := []struct {
		name string
		h    layout.Alignment
		v    layout.Alignment
	}{
		{"Top Left", layout.AlignLeft, layout.AlignTop},
		{"Top Center", layout.AlignCenter, layout.AlignTop},
		{"Top Right", layout.AlignRight, layout.AlignTop},
		{"Middle Left", layout.AlignLeft, layout.AlignMiddle},
		{"Center", layout.AlignCenter, layout.AlignMiddle},
		{"Middle Right", layout.AlignRight, layout.AlignMiddle},
		{"Bottom Left", layout.AlignLeft, layout.AlignBottom},
		{"Bottom Center", layout.AlignCenter, layout.AlignBottom},
		{"Bottom Right", layout.AlignRight, layout.AlignBottom},
	}

	// Create boxes showing each alignment
	boxes := make([]string, 0, 3)
	for i := 0; i < len(alignments); i += 3 {
		row := make([]string, 0, 3)
		for j := 0; j < 3 && i+j < len(alignments); j++ {
			a := alignments[i+j]
			aligned := layout.Align(content, width, height, a.h, a.v)

			// Put in a box to show boundaries
			box := layout.NewBox(aligned).
				WithTitle(a.name).
				WithStyle(layout.BoxStyleSingle)
			row = append(row, box.Render())
		}
		boxes = append(boxes, layout.Columns(row, []int{25, 25, 25}, 2))
	}

	result.WriteString(layout.Rows(boxes, 1))

	return result.String()
}

func (d *LayoutDemo) complexExample() string {
	var result strings.Builder

	// Create a complex layout combining multiple utilities

	// Header with title
	header := layout.Center("📊 Dashboard", 60, 3)
	headerBox := layout.DrawBoxWithTitle(header, "System Status", layout.BoxStyleDouble)

	// Stats in columns
	stat1 := layout.NewBox("CPU: 45%\nMem: 2.3GB").
		WithStyle(layout.BoxStyleRounded).
		WithTitle("Resources").
		WithUniformPadding(1)

	stat2 := layout.NewBox("Requests: 1.2k/s\nLatency: 45ms").
		WithStyle(layout.BoxStyleRounded).
		WithTitle("Performance").
		WithUniformPadding(1)

	stat3 := layout.NewBox("Users: 523\nSessions: 89").
		WithStyle(layout.BoxStyleRounded).
		WithTitle("Activity").
		WithUniformPadding(1)

	statsRow := layout.Columns(
		[]string{stat1.Render(), stat2.Render(), stat3.Render()},
		[]int{20, 20, 20},
		1,
	)

	// Main content area with margin
	content := `Recent Events:
• User login from 192.168.1.100
• Database backup completed
• Cache refresh triggered
• API rate limit warning`

	contentWithMargin := layout.Margin(content, 1, 2, 1, 2)
	contentBox := layout.DrawBoxWithTitle(contentWithMargin, "Activity Log", layout.BoxStyleSingle)

	// Footer
	footer := layout.Center("Press Q to quit", 60, 1)

	// Combine everything
	result.WriteString(headerBox)
	result.WriteString("\n\n")
	result.WriteString(statsRow)
	result.WriteString("\n\n")
	result.WriteString(contentBox)
	result.WriteString("\n")
	result.WriteString(layout.HorizontalLine(60, layout.BoxStyleSingle))
	result.WriteString("\n")
	result.WriteString(footer)

	return result.String()
}

func main() {
	// Component factory
	factory := func() terminus.Component {
		return NewLayoutDemo()
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

	fmt.Println("Layout Demo is running on http://localhost:8890")
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
