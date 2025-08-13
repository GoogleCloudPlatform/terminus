# TerminusGo Dashboard Example

This is the most advanced example in the TerminusGo framework, showcasing a complete system performance dashboard with real-time updates, multiple panels, and interactive widgets.

## Features Demonstrated

### 1. **Complex Layouts**
- Multi-panel dashboard using the Grid layout system
- Responsive design with 3x3 grid arrangement
- Different box styles for visual hierarchy
- Panel spanning across multiple grid cells

### 2. **Widgets**
- **Table Widget**: Process list with sortable columns and cell selection
- **List Widget**: Alert log with color-coded severity levels
- **TextInput Widget**: Command input for system operations
- **Spinner Widgets**: Loading states with different animation styles

### 3. **Real-Time Data**
- CPU usage with historical line chart
- Memory usage with progress bar and chart
- Network I/O with dual-line chart
- Process monitoring with live updates
- System alerts and notifications

### 4. **Performance Optimizations**
- Render caching to minimize redraws
- Efficient data structures for historical data
- Throttled updates with configurable refresh rates
- Smart diffing to update only changed elements

### 5. **Interactive Features**
- Tab navigation between panels
- Keyboard shortcuts for all major functions
- Focus management for input widgets
- Help overlay system

### 6. **ASCII Art Charts**
- Line charts for time-series data
- Dual-line charts for comparative metrics
- Progress bars with color coding
- Custom chart rendering with proper scaling

## Running the Example

```bash
cd examples/dashboard
go run main.go
```

Then open your browser to http://localhost:8895

## Keyboard Controls

- **Tab**: Switch between panels
- **R**: Toggle auto-refresh
- **+/-**: Increase/decrease refresh rate
- **C**: Clear all alerts
- **H**: Show/hide help
- **P**: Toggle performance caching
- **Q**: Quit application

## Panel Overview

### CPU Usage Panel
- Real-time CPU percentage
- 60-second historical line chart
- Color-coded based on usage levels

### Memory Usage Panel
- Current memory usage and total
- Progress bar visualization
- Historical usage chart

### Network Panel
- Inbound/outbound traffic rates
- Dual-line chart for comparison
- MB/s measurement

### Process Table
- Top processes by CPU usage
- Sortable columns
- Status indicators with colors

### Alerts Panel
- Time-stamped system alerts
- Color-coded severity levels
- Scrollable list view

### System Info Panel
- System uptime
- Process and goroutine counts
- Update statistics
- Refresh status

### Command Panel
- Execute system commands
- Command history
- Result notifications

## Architecture Highlights

### Performance Optimization
```go
// Render caching for static content
if d.cacheEnabled && d.updateCount > 0 && d.updateCount%5 != 0 {
    if cached, ok := d.renderCache["full"]; ok && cached != "" {
        return cached
    }
}
```

### Real-Time Updates
```go
// Efficient data structure for time-series
d.cpuHistory = append(d.cpuHistory, d.stats.CPUUsage)
if len(d.cpuHistory) > 60 {
    d.cpuHistory = d.cpuHistory[1:]
}
```

### Focus Management
```go
// Tab navigation between panels
case terminus.KeyTab:
    d.focusedPanel = (d.focusedPanel + 1) % len(d.panels)
    // Update widget focus states...
```

## Customization

The dashboard can be easily customized:

1. **Add New Metrics**: Extend the `SystemStats` struct
2. **Create New Panels**: Add to the grid layout
3. **Custom Charts**: Modify the chart rendering functions
4. **New Widgets**: Integrate additional TerminusGo widgets
5. **Data Sources**: Replace simulated data with real system metrics

## Performance Considerations

- The dashboard uses render caching to avoid unnecessary redraws
- Historical data is capped at 60 data points to limit memory usage
- Updates are throttled based on the refresh rate
- Grid layout minimizes string concatenation overhead

This example serves as a comprehensive showcase of TerminusGo's capabilities and can be used as a starting point for building complex terminal applications in the browser.