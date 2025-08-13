package main

import (
	"embed"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/yourusername/terminusgo/pkg/terminus"
	"github.com/yourusername/terminusgo/pkg/terminus/layout"
	"github.com/yourusername/terminusgo/pkg/terminus/widget"
)

//go:embed all:static/*
var staticFiles embed.FS

// SystemStats holds real-time system statistics
type SystemStats struct {
	CPUUsage    float64
	MemoryUsage float64
	MemoryTotal float64
	NetworkIn   float64
	NetworkOut  float64
	Processes   int
	Goroutines  int
	Uptime      time.Duration
}

// ProcessInfo represents a system process
type ProcessInfo struct {
	PID    int
	Name   string
	CPU    float64
	Memory float64
	Status string
}

// Alert represents a system alert
type Alert struct {
	ID        string
	Timestamp time.Time
	Level     string
	Message   string
}

// Dashboard is the main dashboard component
type Dashboard struct {
	// Layout management
	focusedPanel int
	panels       []string

	// Real-time data
	stats         SystemStats
	statsMutex    sync.RWMutex
	cpuHistory    []float64
	memHistory    []float64
	netInHistory  []float64
	netOutHistory []float64

	// Widgets
	processTable *widget.Table
	alertList    *widget.List
	commandInput *widget.TextInput

	// UI state
	refreshRate    time.Duration
	showHelp       bool
	selectedMetric int
	autoRefresh    bool

	// Data
	processes   []ProcessInfo
	alerts      []Alert
	startTime   time.Time
	lastUpdate  time.Time
	updateCount int

	// Performance optimization
	lastRender   string
	renderCache  map[string]string
	cacheEnabled bool

	// Spinners for loading states
	cpuSpinner     *widget.Spinner
	memSpinner     *widget.Spinner
	netSpinner     *widget.Spinner
	processSpinner *widget.Spinner
}

func NewDashboard() *Dashboard {
	d := &Dashboard{
		focusedPanel: 0,
		panels: []string{
			"CPU", "Memory", "Network", "Processes", "Alerts", "Command",
		},
		refreshRate:   time.Second,
		autoRefresh:   true,
		startTime:     time.Now(),
		renderCache:   make(map[string]string),
		cacheEnabled:  true,
		cpuHistory:    make([]float64, 0, 60),
		memHistory:    make([]float64, 0, 60),
		netInHistory:  make([]float64, 0, 60),
		netOutHistory: make([]float64, 0, 60),
		alerts:        make([]Alert, 0),
		processes:     make([]ProcessInfo, 0),
	}

	// Initialize process table
	d.processTable = widget.NewTable().
		SetShowHeader(true).
		SetShowRowNumbers(false).
		SetStyle(terminus.NewStyle()).
		SetHeaderStyle(terminus.NewStyle().Bold(true).Foreground(terminus.Cyan)).
		SetSelectedStyle(terminus.NewStyle().Reverse(true))

	// Set process table columns
	columns := []widget.TableColumn{
		{Title: "PID", Width: 8, Align: widget.AlignRight},
		{Title: "Name", Width: 20, Align: widget.AlignLeft},
		{Title: "CPU %", Width: 8, Align: widget.AlignRight},
		{Title: "Mem %", Width: 8, Align: widget.AlignRight},
		{Title: "Status", Width: 10, Align: widget.AlignCenter},
	}
	d.processTable.SetColumns(columns)

	// Initialize alert list
	d.alertList = widget.NewList().
		SetShowCursor(false).
		SetStyle(terminus.NewStyle())

	// Initialize command input
	d.commandInput = widget.NewTextInput().
		SetPlaceholder("Enter command...").
		SetStyle(terminus.NewStyle().Foreground(terminus.White)).
		SetFocusStyle(terminus.NewStyle().Foreground(terminus.Cyan).Underline(true)).
		SetOnSubmit(func(value string) terminus.Cmd {
			return d.executeCommand(value)
		})

	// Initialize spinners
	d.cpuSpinner = widget.NewSpinner().
		SetSpinnerStyle(widget.SpinnerDots).
		SetText("Loading CPU...").
		SetSpinnerColor(terminus.NewStyle().Foreground(terminus.Cyan))

	d.memSpinner = widget.NewSpinner().
		SetSpinnerStyle(widget.SpinnerLine).
		SetText("Loading Memory...").
		SetSpinnerColor(terminus.NewStyle().Foreground(terminus.Green))

	d.netSpinner = widget.NewSpinner().
		SetSpinnerStyle(widget.SpinnerCircle).
		SetText("Loading Network...").
		SetSpinnerColor(terminus.NewStyle().Foreground(terminus.Yellow))

	d.processSpinner = widget.NewSpinner().
		SetSpinnerStyle(widget.SpinnerBounce).
		SetText("Scanning processes...").
		SetSpinnerColor(terminus.NewStyle().Foreground(terminus.Magenta))

	// Start spinners
	d.cpuSpinner.Start()
	d.memSpinner.Start()
	d.netSpinner.Start()
	d.processSpinner.Start()

	// Generate initial data
	d.generateInitialData()

	return d
}

func (d *Dashboard) Init() terminus.Cmd {
	// Start auto-refresh
	return d.startAutoRefresh()
}

func (d *Dashboard) Update(msg terminus.Msg) (terminus.Component, terminus.Cmd) {
	var cmds []terminus.Cmd

	switch msg := msg.(type) {
	case terminus.KeyMsg:
		cmd := d.handleKeyPress(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case refreshMsg:
		d.updateStats()
		d.updateCount++
		d.lastUpdate = time.Now()

		// Continue auto-refresh if enabled
		if d.autoRefresh {
			cmds = append(cmds, d.scheduleRefresh())
		}

		// Stop spinners after first update
		if d.updateCount == 1 {
			d.cpuSpinner.Stop()
			d.memSpinner.Stop()
			d.netSpinner.Stop()
			d.processSpinner.Stop()
		}

	case commandResultMsg:
		d.addAlert("info", msg.result)

	case widget.SpinnerTickMsg:
		// Forward to all spinners
		d.cpuSpinner.Update(msg)
		d.memSpinner.Update(msg)
		d.netSpinner.Update(msg)
		d.processSpinner.Update(msg)
	}

	// Update focused widget
	switch d.panels[d.focusedPanel] {
	case "Processes":
		if d.processTable != nil {
			newTable, cmd := d.processTable.Update(msg)
			d.processTable = newTable.(*widget.Table)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	case "Alerts":
		if d.alertList != nil {
			newList, cmd := d.alertList.Update(msg)
			d.alertList = newList.(*widget.List)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	case "Command":
		if d.commandInput != nil {
			newInput, cmd := d.commandInput.Update(msg)
			d.commandInput = newInput.(*widget.TextInput)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	if len(cmds) > 0 {
		return d, terminus.Batch(cmds...)
	}
	return d, nil
}

func (d *Dashboard) View() string {
	// Performance optimization: check if we can use cached render
	if d.cacheEnabled && d.updateCount > 0 && d.updateCount%5 != 0 {
		if cached, ok := d.renderCache["full"]; ok && cached != "" {
			return cached
		}
	}

	var result strings.Builder

	// Header
	d.renderHeader(&result)
	result.WriteString("\n")

	// Main content area using grid layout
	grid := layout.NewGrid(3, 3).SetGap(1)

	// Top row: CPU, Memory, Network graphs
	grid.SetCell(0, 0, d.renderCPUPanel())
	grid.SetCell(1, 0, d.renderMemoryPanel())
	grid.SetCell(2, 0, d.renderNetworkPanel())

	// Middle row: Process table (spans 2 columns), Alerts
	processPanel := d.renderProcessPanel()
	grid.SetCell(0, 1, processPanel)
	grid.SetCell(1, 1, "") // Process panel spans this cell
	grid.SetCell(2, 1, d.renderAlertsPanel())

	// Bottom row: System info, Command input (spans 2 columns)
	grid.SetCell(0, 2, d.renderSystemInfoPanel())
	commandPanel := d.renderCommandPanel()
	grid.SetCell(1, 2, commandPanel)
	grid.SetCell(2, 2, "") // Command panel spans this cell

	// Set column widths
	grid.SetColumnWidth(0, 40)
	grid.SetColumnWidth(1, 40)
	grid.SetColumnWidth(2, 40)

	result.WriteString(grid.Render())
	result.WriteString("\n")

	// Footer
	d.renderFooter(&result)

	// Help overlay
	if d.showHelp {
		result.WriteString("\n\n")
		result.WriteString(d.renderHelp())
	}

	rendered := result.String()

	// Cache the render
	if d.cacheEnabled {
		d.renderCache["full"] = rendered
	}

	return rendered
}

// Panel rendering methods

func (d *Dashboard) renderHeader(result *strings.Builder) {
	titleStyle := terminus.NewStyle().Bold(true).Foreground(terminus.Cyan)
	headerBox := layout.NewBox(
		titleStyle.Render("System Performance Dashboard"),
	).WithStyle(layout.BoxStyleDouble).
		WithUniformPadding(1).
		WithWidth(122)

	result.WriteString(headerBox.Render())
}

func (d *Dashboard) renderCPUPanel() string {
	d.statsMutex.RLock()
	cpuUsage := d.stats.CPUUsage
	history := append([]float64{}, d.cpuHistory...)
	d.statsMutex.RUnlock()

	var content strings.Builder

	// Title with current value
	titleStyle := terminus.NewStyle().Bold(true)
	if cpuUsage > 80 {
		titleStyle = titleStyle.Foreground(terminus.Red)
	} else if cpuUsage > 60 {
		titleStyle = titleStyle.Foreground(terminus.Yellow)
	} else {
		titleStyle = titleStyle.Foreground(terminus.Green)
	}

	content.WriteString(titleStyle.Render(fmt.Sprintf("CPU: %.1f%%", cpuUsage)))
	content.WriteString("\n\n")

	// Show spinner or graph
	if d.updateCount == 0 {
		content.WriteString(d.cpuSpinner.View())
	} else {
		// ASCII chart
		chart := d.renderLineChart(history, 35, 8, "CPU Usage %")
		content.WriteString(chart)
	}

	// Create box with appropriate style
	boxStyle := layout.BoxStyleSingle
	if d.focusedPanel == 0 {
		boxStyle = layout.BoxStyleDouble
	}

	return layout.NewBox(content.String()).
		WithStyle(boxStyle).
		WithTitle("CPU Usage").
		WithUniformPadding(1).
		Render()
}

func (d *Dashboard) renderMemoryPanel() string {
	d.statsMutex.RLock()
	memUsage := d.stats.MemoryUsage
	memTotal := d.stats.MemoryTotal
	history := append([]float64{}, d.memHistory...)
	d.statsMutex.RUnlock()

	var content strings.Builder

	// Title with current value
	titleStyle := terminus.NewStyle().Bold(true)
	memPercent := (memUsage / memTotal) * 100
	if memPercent > 80 {
		titleStyle = titleStyle.Foreground(terminus.Red)
	} else if memPercent > 60 {
		titleStyle = titleStyle.Foreground(terminus.Yellow)
	} else {
		titleStyle = titleStyle.Foreground(terminus.Green)
	}

	content.WriteString(titleStyle.Render(fmt.Sprintf("Memory: %.1f/%.1f GB (%.1f%%)",
		memUsage, memTotal, memPercent)))
	content.WriteString("\n\n")

	// Show spinner or graph
	if d.updateCount == 0 {
		content.WriteString(d.memSpinner.View())
	} else {
		// Progress bar
		bar := d.renderProgressBar(memPercent, 30)
		content.WriteString(bar)
		content.WriteString("\n\n")

		// ASCII chart
		chart := d.renderLineChart(history, 35, 6, "Memory %")
		content.WriteString(chart)
	}

	boxStyle := layout.BoxStyleSingle
	if d.focusedPanel == 1 {
		boxStyle = layout.BoxStyleDouble
	}

	return layout.NewBox(content.String()).
		WithStyle(boxStyle).
		WithTitle("Memory Usage").
		WithUniformPadding(1).
		Render()
}

func (d *Dashboard) renderNetworkPanel() string {
	d.statsMutex.RLock()
	netIn := d.stats.NetworkIn
	netOut := d.stats.NetworkOut
	inHistory := append([]float64{}, d.netInHistory...)
	outHistory := append([]float64{}, d.netOutHistory...)
	d.statsMutex.RUnlock()

	var content strings.Builder

	// Title
	titleStyle := terminus.NewStyle().Bold(true).Foreground(terminus.Blue)
	content.WriteString(titleStyle.Render("Network I/O"))
	content.WriteString("\n\n")

	// Show spinner or stats
	if d.updateCount == 0 {
		content.WriteString(d.netSpinner.View())
	} else {
		// Current stats
		inStyle := terminus.NewStyle().Foreground(terminus.Green)
		outStyle := terminus.NewStyle().Foreground(terminus.Yellow)

		content.WriteString(inStyle.Render(fmt.Sprintf("↓ In:  %.2f MB/s", netIn)))
		content.WriteString("\n")
		content.WriteString(outStyle.Render(fmt.Sprintf("↑ Out: %.2f MB/s", netOut)))
		content.WriteString("\n\n")

		// Dual line chart
		chart := d.renderDualLineChart(inHistory, outHistory, 35, 6, "In", "Out")
		content.WriteString(chart)
	}

	boxStyle := layout.BoxStyleSingle
	if d.focusedPanel == 2 {
		boxStyle = layout.BoxStyleDouble
	}

	return layout.NewBox(content.String()).
		WithStyle(boxStyle).
		WithTitle("Network").
		WithUniformPadding(1).
		Render()
}

func (d *Dashboard) renderProcessPanel() string {
	var content strings.Builder

	// Show spinner or table
	if d.updateCount == 0 {
		content.WriteString(d.processSpinner.View())
	} else {
		// Update table data
		headers := []string{"PID", "Name", "CPU %", "Mem %", "Status"}
		data := make([][]string, len(d.processes))
		for i, p := range d.processes {
			statusStyle := terminus.NewStyle()
			switch p.Status {
			case "Running":
				statusStyle = statusStyle.Foreground(terminus.Green)
			case "Sleeping":
				statusStyle = statusStyle.Foreground(terminus.Blue)
			case "Stopped":
				statusStyle = statusStyle.Foreground(terminus.Red)
			}

			data[i] = []string{
				fmt.Sprintf("%d", p.PID),
				p.Name,
				fmt.Sprintf("%.1f", p.CPU),
				fmt.Sprintf("%.1f", p.Memory),
				statusStyle.Render(p.Status),
			}
		}

		d.processTable.SetStringData(headers, data)
		d.processTable.SetSize(70, 10)
		content.WriteString(d.processTable.View())
	}

	boxStyle := layout.BoxStyleSingle
	if d.focusedPanel == 3 {
		boxStyle = layout.BoxStyleDouble
	}

	box := layout.NewBox(content.String()).
		WithStyle(boxStyle).
		WithTitle(fmt.Sprintf("Processes (%d)", len(d.processes))).
		WithUniformPadding(1)

	// Span 2 columns
	box.WithWidth(81)

	return box.Render()
}

func (d *Dashboard) renderAlertsPanel() string {
	var content strings.Builder

	// Convert alerts to list items
	items := make([]widget.ListItem, len(d.alerts))
	for i, alert := range d.alerts {
		levelStyle := terminus.NewStyle()
		switch alert.Level {
		case "error":
			levelStyle = levelStyle.Foreground(terminus.Red)
		case "warning":
			levelStyle = levelStyle.Foreground(terminus.Yellow)
		case "info":
			levelStyle = levelStyle.Foreground(terminus.Blue)
		}

		timeStr := alert.Timestamp.Format("15:04:05")
		items[i] = widget.NewSimpleListItem(fmt.Sprintf("%s [%s] %s",
			timeStr,
			levelStyle.Render(strings.ToUpper(alert.Level)),
			alert.Message,
		))
	}

	d.alertList.SetItems(items)
	d.alertList.SetSize(35, 10)
	content.WriteString(d.alertList.View())

	boxStyle := layout.BoxStyleSingle
	if d.focusedPanel == 4 {
		boxStyle = layout.BoxStyleDouble
	}

	return layout.NewBox(content.String()).
		WithStyle(boxStyle).
		WithTitle(fmt.Sprintf("Alerts (%d)", len(d.alerts))).
		WithUniformPadding(1).
		Render()
}

func (d *Dashboard) renderSystemInfoPanel() string {
	d.statsMutex.RLock()
	stats := d.stats
	d.statsMutex.RUnlock()

	var content strings.Builder

	infoStyle := terminus.NewStyle().Foreground(terminus.White)
	labelStyle := terminus.NewStyle().Faint(true)

	content.WriteString(labelStyle.Render("Uptime:     "))
	content.WriteString(infoStyle.Render(d.formatDuration(stats.Uptime)))
	content.WriteString("\n")

	content.WriteString(labelStyle.Render("Processes:  "))
	content.WriteString(infoStyle.Render(fmt.Sprintf("%d", stats.Processes)))
	content.WriteString("\n")

	content.WriteString(labelStyle.Render("Goroutines: "))
	content.WriteString(infoStyle.Render(fmt.Sprintf("%d", stats.Goroutines)))
	content.WriteString("\n")

	content.WriteString(labelStyle.Render("Updates:    "))
	content.WriteString(infoStyle.Render(fmt.Sprintf("%d", d.updateCount)))
	content.WriteString("\n")

	content.WriteString(labelStyle.Render("Refresh:    "))
	if d.autoRefresh {
		content.WriteString(terminus.NewStyle().Foreground(terminus.Green).Render("ON"))
	} else {
		content.WriteString(terminus.NewStyle().Foreground(terminus.Red).Render("OFF"))
	}
	content.WriteString(fmt.Sprintf(" (%s)", d.refreshRate))

	return layout.NewBox(content.String()).
		WithStyle(layout.BoxStyleRounded).
		WithTitle("System Info").
		WithUniformPadding(1).
		Render()
}

func (d *Dashboard) renderCommandPanel() string {
	var content strings.Builder

	content.WriteString("Execute system commands:\n")
	d.commandInput.SetSize(60, 1)
	content.WriteString(d.commandInput.View())

	boxStyle := layout.BoxStyleSingle
	if d.focusedPanel == 5 {
		boxStyle = layout.BoxStyleDouble
	}

	box := layout.NewBox(content.String()).
		WithStyle(boxStyle).
		WithTitle("Command").
		WithUniformPadding(1)

	// Span 2 columns
	box.WithWidth(81)

	return box.Render()
}

func (d *Dashboard) renderFooter(result *strings.Builder) {
	footerStyle := terminus.NewStyle().Faint(true)
	shortcuts := []string{
		"[Tab] Switch Panel",
		"[R] Toggle Refresh",
		"[+/-] Change Rate",
		"[C] Clear Alerts",
		"[H] Help",
		"[Q] Quit",
	}

	footer := footerStyle.Render(strings.Join(shortcuts, " | "))
	result.WriteString(layout.Center(footer, 122, 1))
}

func (d *Dashboard) renderHelp() string {
	helpContent := `
Keyboard Shortcuts:

Navigation:
  Tab         - Switch between panels
  Arrow Keys  - Navigate within tables/lists

Controls:
  R           - Toggle auto-refresh
  +           - Increase refresh rate
  -           - Decrease refresh rate
  C           - Clear all alerts
  H           - Toggle this help
  Q           - Quit application

Panel-Specific:
  Enter       - Select item (in lists/tables)
  S           - Sort table column (in process table)
  /           - Filter (in process table)

Performance:
  P           - Toggle render caching
  D           - Show debug info
`

	return layout.DrawBoxWithTitle(helpContent, "Help", layout.BoxStyleDouble)
}

// Chart rendering methods

func (d *Dashboard) renderLineChart(data []float64, width, height int, label string) string {
	if len(data) == 0 {
		return "No data"
	}

	// Ensure we don't exceed the desired width
	if len(data) > width {
		data = data[len(data)-width:]
	}

	// Find max for scaling
	max := 100.0
	for _, v := range data {
		if v > max {
			max = v
		}
	}

	// Create the chart
	var chart strings.Builder

	// Y-axis labels
	for y := height - 1; y >= 0; y-- {
		value := (float64(y) / float64(height-1)) * max
		if y == height-1 {
			chart.WriteString(fmt.Sprintf("%3.0f┤", value))
		} else if y == 0 {
			chart.WriteString(fmt.Sprintf("%3.0f┤", value))
		} else {
			chart.WriteString("   │")
		}

		// Plot the line
		for x := 0; x < len(data); x++ {
			scaledValue := (data[x] / max) * float64(height-1)
			plotY := int(math.Round(scaledValue))

			if plotY == y {
				if x > 0 && x < len(data)-1 {
					// Check neighbors for line direction
					prevY := int(math.Round((data[x-1] / max) * float64(height-1)))
					nextY := int(math.Round((data[x+1] / max) * float64(height-1)))

					if prevY < y && nextY > y {
						chart.WriteString("╱")
					} else if prevY > y && nextY < y {
						chart.WriteString("╲")
					} else {
						chart.WriteString("─")
					}
				} else {
					chart.WriteString("●")
				}
			} else {
				chart.WriteString(" ")
			}
		}
		chart.WriteString("\n")
	}

	// X-axis
	chart.WriteString("   └")
	chart.WriteString(strings.Repeat("─", len(data)))
	chart.WriteString("\n")

	// X-axis label
	chart.WriteString("    ")
	chart.WriteString(layout.Center(label, len(data), 1))

	return chart.String()
}

func (d *Dashboard) renderDualLineChart(data1, data2 []float64, width, height int, label1, label2 string) string {
	if len(data1) == 0 && len(data2) == 0 {
		return "No data"
	}

	// Ensure equal length
	maxLen := len(data1)
	if len(data2) > maxLen {
		maxLen = len(data2)
	}
	if maxLen > width {
		if len(data1) > width {
			data1 = data1[len(data1)-width:]
		}
		if len(data2) > width {
			data2 = data2[len(data2)-width:]
		}
		maxLen = width
	}

	// Find global max for scaling
	max := 0.0
	for _, v := range data1 {
		if v > max {
			max = v
		}
	}
	for _, v := range data2 {
		if v > max {
			max = v
		}
	}
	if max == 0 {
		max = 1
	}

	var chart strings.Builder

	// Legend
	inStyle := terminus.NewStyle().Foreground(terminus.Green)
	outStyle := terminus.NewStyle().Foreground(terminus.Yellow)
	chart.WriteString(inStyle.Render("─ " + label1))
	chart.WriteString("  ")
	chart.WriteString(outStyle.Render("─ " + label2))
	chart.WriteString("\n")

	// Y-axis and plot
	for y := height - 1; y >= 0; y-- {
		value := (float64(y) / float64(height-1)) * max
		if y == height-1 || y == 0 {
			chart.WriteString(fmt.Sprintf("%3.0f┤", value))
		} else {
			chart.WriteString("   │")
		}

		// Plot both lines
		for x := 0; x < maxLen; x++ {
			char := " "

			if x < len(data1) {
				scaledValue1 := (data1[x] / max) * float64(height-1)
				plotY1 := int(math.Round(scaledValue1))
				if plotY1 == y {
					char = inStyle.Render("─")
				}
			}

			if x < len(data2) {
				scaledValue2 := (data2[x] / max) * float64(height-1)
				plotY2 := int(math.Round(scaledValue2))
				if plotY2 == y {
					if char != " " {
						char = "┼" // Both lines intersect
					} else {
						char = outStyle.Render("─")
					}
				}
			}

			chart.WriteString(char)
		}
		chart.WriteString("\n")
	}

	// X-axis
	chart.WriteString("   └")
	chart.WriteString(strings.Repeat("─", maxLen))

	return chart.String()
}

func (d *Dashboard) renderProgressBar(percent float64, width int) string {
	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	var bar strings.Builder
	bar.WriteString("[")

	style := terminus.NewStyle()
	if percent > 80 {
		style = style.Foreground(terminus.Red)
	} else if percent > 60 {
		style = style.Foreground(terminus.Yellow)
	} else {
		style = style.Foreground(terminus.Green)
	}

	bar.WriteString(style.Render(strings.Repeat("█", filled)))
	bar.WriteString(strings.Repeat("░", width-filled))
	bar.WriteString("]")

	return bar.String()
}

// Helper methods

func (d *Dashboard) handleKeyPress(msg terminus.KeyMsg) terminus.Cmd {
	switch msg.Type {
	case terminus.KeyCtrlC:
		return terminus.Quit

	case terminus.KeyTab:
		// Clear focus from current panel
		switch d.panels[d.focusedPanel] {
		case "Processes":
			d.processTable.Blur()
		case "Alerts":
			d.alertList.Blur()
		case "Command":
			d.commandInput.Blur()
		}

		// Move to next panel
		d.focusedPanel = (d.focusedPanel + 1) % len(d.panels)

		// Focus new panel
		switch d.panels[d.focusedPanel] {
		case "Processes":
			d.processTable.Focus()
		case "Alerts":
			d.alertList.Focus()
		case "Command":
			d.commandInput.Focus()
		}

		// Clear render cache when switching panels
		d.renderCache = make(map[string]string)

		return nil

	case terminus.KeyRunes:
		if len(msg.Runes) > 0 {
			switch msg.Runes[0] {
			case 'q', 'Q':
				return terminus.Quit
			case 'r', 'R':
				d.autoRefresh = !d.autoRefresh
				if d.autoRefresh {
					return d.scheduleRefresh()
				}
				return nil
			case '+':
				if d.refreshRate > 500*time.Millisecond {
					d.refreshRate -= 500 * time.Millisecond
				}
				return nil
			case '-':
				if d.refreshRate < 5*time.Second {
					d.refreshRate += 500 * time.Millisecond
				}
				return nil
			case 'c', 'C':
				d.alerts = make([]Alert, 0)
				d.addAlert("info", "Alerts cleared")
				return nil
			case 'h', 'H':
				d.showHelp = !d.showHelp
				return nil
			case 'p', 'P':
				d.cacheEnabled = !d.cacheEnabled
				if !d.cacheEnabled {
					d.renderCache = make(map[string]string)
				}
				return nil
			}
		}
	}

	return nil
}

func (d *Dashboard) generateInitialData() {
	// Initialize with some data
	d.statsMutex.Lock()
	defer d.statsMutex.Unlock()

	// CPU history
	for i := 0; i < 30; i++ {
		d.cpuHistory = append(d.cpuHistory, rand.Float64()*100)
	}

	// Memory history
	for i := 0; i < 30; i++ {
		d.memHistory = append(d.memHistory, 40+rand.Float64()*30)
	}

	// Network history
	for i := 0; i < 30; i++ {
		d.netInHistory = append(d.netInHistory, rand.Float64()*10)
		d.netOutHistory = append(d.netOutHistory, rand.Float64()*5)
	}

	// Initial stats
	d.stats = SystemStats{
		CPUUsage:    rand.Float64() * 100,
		MemoryUsage: 4 + rand.Float64()*4,
		MemoryTotal: 16.0,
		NetworkIn:   rand.Float64() * 10,
		NetworkOut:  rand.Float64() * 5,
		Processes:   120 + rand.Intn(30),
		Goroutines:  runtime.NumGoroutine(),
		Uptime:      time.Since(d.startTime),
	}

	// Generate some processes
	processNames := []string{
		"systemd", "kernel", "chrome", "firefox", "code",
		"docker", "postgres", "redis", "nginx", "node",
		"python3", "go", "java", "mysql", "mongodb",
	}

	for i := 0; i < 15; i++ {
		d.processes = append(d.processes, ProcessInfo{
			PID:    1000 + i,
			Name:   processNames[i%len(processNames)],
			CPU:    rand.Float64() * 20,
			Memory: rand.Float64() * 10,
			Status: []string{"Running", "Sleeping", "Stopped"}[rand.Intn(3)],
		})
	}

	// Initial alerts
	d.addAlert("info", "Dashboard started")
	d.addAlert("warning", "High memory usage detected")
}

func (d *Dashboard) updateStats() {
	d.statsMutex.Lock()
	defer d.statsMutex.Unlock()

	// Simulate real-time data updates
	d.stats.CPUUsage = math.Max(0, math.Min(100, d.stats.CPUUsage+(rand.Float64()-0.5)*10))
	d.stats.MemoryUsage = math.Max(1, math.Min(d.stats.MemoryTotal-0.5, d.stats.MemoryUsage+(rand.Float64()-0.5)*0.5))
	d.stats.NetworkIn = math.Max(0, d.stats.NetworkIn+(rand.Float64()-0.5)*2)
	d.stats.NetworkOut = math.Max(0, d.stats.NetworkOut+(rand.Float64()-0.5)*1)
	d.stats.Processes = 120 + rand.Intn(30)
	d.stats.Goroutines = runtime.NumGoroutine()
	d.stats.Uptime = time.Since(d.startTime)

	// Update history (keep last 60 values)
	d.cpuHistory = append(d.cpuHistory, d.stats.CPUUsage)
	if len(d.cpuHistory) > 60 {
		d.cpuHistory = d.cpuHistory[1:]
	}

	memPercent := (d.stats.MemoryUsage / d.stats.MemoryTotal) * 100
	d.memHistory = append(d.memHistory, memPercent)
	if len(d.memHistory) > 60 {
		d.memHistory = d.memHistory[1:]
	}

	d.netInHistory = append(d.netInHistory, d.stats.NetworkIn)
	if len(d.netInHistory) > 60 {
		d.netInHistory = d.netInHistory[1:]
	}

	d.netOutHistory = append(d.netOutHistory, d.stats.NetworkOut)
	if len(d.netOutHistory) > 60 {
		d.netOutHistory = d.netOutHistory[1:]
	}

	// Update processes
	for i := range d.processes {
		d.processes[i].CPU = math.Max(0, math.Min(100, d.processes[i].CPU+(rand.Float64()-0.5)*5))
		d.processes[i].Memory = math.Max(0, math.Min(50, d.processes[i].Memory+(rand.Float64()-0.5)*2))

		// Occasionally change status
		if rand.Float64() < 0.1 {
			statuses := []string{"Running", "Sleeping", "Stopped"}
			d.processes[i].Status = statuses[rand.Intn(len(statuses))]
		}
	}

	// Sort processes by CPU usage
	for i := 0; i < len(d.processes)-1; i++ {
		for j := i + 1; j < len(d.processes); j++ {
			if d.processes[j].CPU > d.processes[i].CPU {
				d.processes[i], d.processes[j] = d.processes[j], d.processes[i]
			}
		}
	}

	// Generate occasional alerts
	if rand.Float64() < 0.1 {
		alertTypes := []struct {
			level   string
			message string
		}{
			{"warning", fmt.Sprintf("CPU usage high: %.1f%%", d.stats.CPUUsage)},
			{"info", "Process monitor check completed"},
			{"error", "Failed to connect to monitoring service"},
			{"warning", fmt.Sprintf("Memory usage: %.1f%%", memPercent)},
			{"info", "Network throughput normal"},
		}

		alert := alertTypes[rand.Intn(len(alertTypes))]
		if (alert.level == "warning" && d.stats.CPUUsage > 70) ||
			(alert.level == "warning" && memPercent > 70) ||
			alert.level == "info" ||
			(alert.level == "error" && rand.Float64() < 0.3) {
			d.addAlert(alert.level, alert.message)
		}
	}
}

func (d *Dashboard) addAlert(level, message string) {
	alert := Alert{
		ID:        fmt.Sprintf("alert-%d", len(d.alerts)),
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
	}

	d.alerts = append([]Alert{alert}, d.alerts...)

	// Keep only last 20 alerts
	if len(d.alerts) > 20 {
		d.alerts = d.alerts[:20]
	}
}

func (d *Dashboard) executeCommand(cmd string) terminus.Cmd {
	return func() terminus.Msg {
		// Simulate command execution
		time.Sleep(500 * time.Millisecond)

		switch cmd {
		case "clear":
			d.alerts = make([]Alert, 0)
			return commandResultMsg{result: "Alerts cleared"}
		case "stats":
			return commandResultMsg{result: fmt.Sprintf("Updates: %d, Uptime: %s",
				d.updateCount, d.formatDuration(d.stats.Uptime))}
		case "gc":
			runtime.GC()
			return commandResultMsg{result: "Garbage collection completed"}
		default:
			return commandResultMsg{result: fmt.Sprintf("Command executed: %s", cmd)}
		}
	}
}

func (d *Dashboard) formatDuration(dur time.Duration) string {
	days := int(dur.Hours()) / 24
	hours := int(dur.Hours()) % 24
	minutes := int(dur.Minutes()) % 60
	seconds := int(dur.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

func (d *Dashboard) startAutoRefresh() terminus.Cmd {
	return d.scheduleRefresh()
}

func (d *Dashboard) scheduleRefresh() terminus.Cmd {
	return terminus.Tick(d.refreshRate, func(t time.Time) terminus.Msg {
		return refreshMsg{time: t}
	})
}

// Message types

type refreshMsg struct {
	time time.Time
}

type commandResultMsg struct {
	result string
}

// Main function

func main() {
	// Component factory
	factory := func() terminus.Component {
		return NewDashboard()
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

	fmt.Println("Dashboard is running on http://localhost:8890")
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

	fmt.Println("Dashboard stopped.")
}
