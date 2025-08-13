package layout

import (
	"strings"
)

// Alignment represents text alignment
type Alignment int

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
	AlignTop
	AlignMiddle
	AlignBottom
)

// Layout represents a layout container
type Layout struct {
	width  int
	height int
}

// NewLayout creates a new layout with specified dimensions
func NewLayout(width, height int) *Layout {
	return &Layout{
		width:  width,
		height: height,
	}
}

// Columns arranges content in columns
func Columns(contents []string, widths []int, gap int) string {
	if len(contents) == 0 || len(widths) == 0 {
		return ""
	}

	// Ensure we have enough widths
	if len(widths) < len(contents) {
		// Use last width for remaining columns
		lastWidth := widths[len(widths)-1]
		for i := len(widths); i < len(contents); i++ {
			widths = append(widths, lastWidth)
		}
	}

	// Split each content into lines
	contentLines := make([][]string, len(contents))
	maxLines := 0
	for i, content := range contents {
		contentLines[i] = strings.Split(content, "\n")
		if len(contentLines[i]) > maxLines {
			maxLines = len(contentLines[i])
		}
	}

	var result strings.Builder
	gapStr := strings.Repeat(" ", gap)

	// Render each row
	for row := 0; row < maxLines; row++ {
		if row > 0 {
			result.WriteString("\n")
		}

		for col := 0; col < len(contents); col++ {
			if col > 0 {
				result.WriteString(gapStr)
			}

			// Get line for this column, or empty if beyond content
			line := ""
			if row < len(contentLines[col]) {
				line = contentLines[col][row]
			}

			// Truncate or pad to column width
			if col < len(widths) {
				line = padOrTruncate(line, widths[col], AlignLeft)
			}
			result.WriteString(line)
		}
	}

	return result.String()
}

// Rows arranges content in rows
func Rows(contents []string, gap int) string {
	if len(contents) == 0 {
		return ""
	}

	gapStr := strings.Repeat("\n", gap+1)
	return strings.Join(contents, gapStr)
}

// Center centers content within a given width and height
func Center(content string, width, height int) string {
	lines := strings.Split(content, "\n")

	// Vertical centering
	contentHeight := len(lines)
	topPadding := (height - contentHeight) / 2
	bottomPadding := height - contentHeight - topPadding

	var result strings.Builder

	// Add top padding
	for i := 0; i < topPadding; i++ {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(strings.Repeat(" ", width))
	}

	// Add centered content
	for i, line := range lines {
		if topPadding > 0 || i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(padOrTruncate(line, width, AlignCenter))
	}

	// Add bottom padding
	for i := 0; i < bottomPadding; i++ {
		result.WriteString("\n")
		result.WriteString(strings.Repeat(" ", width))
	}

	return result.String()
}

// Align aligns content within bounds
func Align(content string, width, height int, horizontal, vertical Alignment) string {
	lines := strings.Split(content, "\n")

	// Prepare aligned lines
	alignedLines := make([]string, 0, height)

	// Vertical alignment
	contentHeight := len(lines)
	var startLine int
	switch vertical {
	case AlignTop:
		startLine = 0
	case AlignMiddle:
		startLine = (height - contentHeight) / 2
	case AlignBottom:
		startLine = height - contentHeight
	}

	// Build result
	for i := 0; i < height; i++ {
		var line string
		contentIdx := i - startLine

		if contentIdx >= 0 && contentIdx < len(lines) {
			line = padOrTruncate(lines[contentIdx], width, horizontal)
		} else {
			line = strings.Repeat(" ", width)
		}

		alignedLines = append(alignedLines, line)
	}

	return strings.Join(alignedLines, "\n")
}

// Margin adds margin around content
func Margin(content string, top, right, bottom, left int) string {
	lines := strings.Split(content, "\n")

	// Calculate content width
	maxWidth := 0
	for _, line := range lines {
		lineWidth := visibleLength(line)
		if lineWidth > maxWidth {
			maxWidth = lineWidth
		}
	}

	totalWidth := maxWidth + left + right
	var result strings.Builder

	// Top margin
	emptyLine := strings.Repeat(" ", totalWidth)
	for i := 0; i < top; i++ {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(emptyLine)
	}

	// Content with left/right margin
	leftMargin := strings.Repeat(" ", left)
	for i, line := range lines {
		if top > 0 || i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(leftMargin)
		result.WriteString(line)
		result.WriteString(strings.Repeat(" ", maxWidth-visibleLength(line)+right))
	}

	// Bottom margin
	for i := 0; i < bottom; i++ {
		result.WriteString("\n")
		result.WriteString(emptyLine)
	}

	return result.String()
}

// AddPadding adds padding around content (similar to margin but typically used inside containers)
func AddPadding(content string, top, right, bottom, left int) string {
	// For our purposes, padding and margin work the same way
	return Margin(content, top, right, bottom, left)
}

// padOrTruncate ensures a string is exactly the specified width
func padOrTruncate(s string, width int, align Alignment) string {
	visLen := visibleLength(s)
	
	if visLen >= width {
		// TODO: Implement proper ANSI-aware truncation
		// For now, if the visible length is already at or over width, return as-is
		// to avoid breaking ANSI sequences
		return s
	}

	padding := width - visLen
	switch align {
	case AlignLeft:
		return s + strings.Repeat(" ", padding)
	case AlignRight:
		return strings.Repeat(" ", padding) + s
	case AlignCenter:
		leftPad := padding / 2
		rightPad := padding - leftPad
		return strings.Repeat(" ", leftPad) + s + strings.Repeat(" ", rightPad)
	default:
		return s + strings.Repeat(" ", padding)
	}
}

// Grid creates a grid layout
type Grid struct {
	cols    int
	rows    int
	gap     int
	cells   [][]string
	widths  []int
	heights []int
}

// NewGrid creates a new grid layout
func NewGrid(cols, rows int) *Grid {
	cells := make([][]string, rows)
	for i := range cells {
		cells[i] = make([]string, cols)
	}

	// Initialize heights to 1 (minimum for empty cells)
	heights := make([]int, rows)
	for i := range heights {
		heights[i] = 1
	}

	return &Grid{
		cols:    cols,
		rows:    rows,
		gap:     1,
		cells:   cells,
		widths:  make([]int, cols),
		heights: heights,
	}
}

// SetCell sets content for a specific cell
func (g *Grid) SetCell(col, row int, content string) *Grid {
	if row >= 0 && row < g.rows && col >= 0 && col < g.cols {
		g.cells[row][col] = content

		// Update column width if needed
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			lineWidth := visibleLength(line)
			if lineWidth > g.widths[col] {
				g.widths[col] = lineWidth
			}
		}

		// Update row height if needed
		if len(lines) > g.heights[row] {
			g.heights[row] = len(lines)
		}
	}
	return g
}

// SetGap sets the gap between cells
func (g *Grid) SetGap(gap int) *Grid {
	g.gap = gap
	return g
}

// SetColumnWidth sets a fixed width for a column
func (g *Grid) SetColumnWidth(col, width int) *Grid {
	if col >= 0 && col < g.cols {
		g.widths[col] = width
	}
	return g
}

// SetRowHeight sets a fixed height for a row
func (g *Grid) SetRowHeight(row, height int) *Grid {
	if row >= 0 && row < g.rows {
		g.heights[row] = height
	}
	return g
}

// Render renders the grid
func (g *Grid) Render() string {
	var result strings.Builder
	gapH := strings.Repeat(" ", g.gap)

	for row := 0; row < g.rows; row++ {
		// Render each line of this row
		for line := 0; line < g.heights[row]; line++ {
			if row > 0 || line > 0 {
				result.WriteString("\n")
			}

			for col := 0; col < g.cols; col++ {
				if col > 0 {
					result.WriteString(gapH)
				}

				// Get the content for this cell
				cellContent := ""
				if g.cells[row][col] != "" {
					lines := strings.Split(g.cells[row][col], "\n")
					if line < len(lines) {
						cellContent = lines[line]
					}
				}

				// Pad to column width
				result.WriteString(padOrTruncate(cellContent, g.widths[col], AlignLeft))
			}
		}

		// Add vertical gap
		if row < g.rows-1 {
			for i := 0; i < g.gap; i++ {
				result.WriteString("\n")
			}
		}
	}

	return result.String()
}
