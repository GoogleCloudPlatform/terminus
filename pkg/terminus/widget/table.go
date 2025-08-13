package widget

import (
	"fmt"
	"sort"
	"strings"

	"github.com/yourusername/terminusgo/pkg/terminus"
)

// TableCell represents a cell in a table
type TableCell interface {
	// Render returns the string representation of the cell
	Render() string
	// String returns a simple string representation (for sorting/filtering)
	String() string
	// Value returns the underlying value (for sorting)
	Value() interface{}
}

// SimpleTableCell is a basic string-based table cell
type SimpleTableCell struct {
	text string
}

// NewSimpleTableCell creates a new simple table cell
func NewSimpleTableCell(text string) *SimpleTableCell {
	return &SimpleTableCell{text: text}
}

// Render implements TableCell interface
func (s *SimpleTableCell) Render() string {
	return s.text
}

// String implements TableCell interface
func (s *SimpleTableCell) String() string {
	return s.text
}

// Value implements TableCell interface
func (s *SimpleTableCell) Value() interface{} {
	return s.text
}

// TableColumn represents a column definition
type TableColumn struct {
	Title    string
	Width    int
	MinWidth int
	MaxWidth int
	Sortable bool
	Align    Alignment
}

// Alignment represents text alignment
type Alignment int

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
)

// TableRow represents a row of data
type TableRow []TableCell

// Table is a table widget with columns and rows
type Table struct {
	Model

	// Data
	columns     []TableColumn
	rows        []TableRow
	selectedRow int
	selectedCol int

	// Display settings
	showHeader     bool
	showRowNumbers bool
	borderStyle    BorderStyle
	scrollOffsetX  int
	scrollOffsetY  int

	// Styling
	style           terminus.Style
	headerStyle     terminus.Style
	selectedStyle   terminus.Style
	borderColor     terminus.Style
	rowNumberStyle  terminus.Style

	// Sorting
	sortColumn int
	sortOrder  SortOrder

	// Selection
	cellSelection bool // If true, individual cells can be selected

	// Events
	onSelect func(row, col int, cell TableCell) terminus.Cmd
	onSort   func(column int, order SortOrder) terminus.Cmd
}

// BorderStyle represents the style of table borders
type BorderStyle int

const (
	BorderNone BorderStyle = iota
	BorderSimple
	BorderDouble
	BorderRounded
)

// SortOrder represents the sorting order
type SortOrder int

const (
	SortNone SortOrder = iota
	SortAsc
	SortDesc
)

// NewTable creates a new table widget
func NewTable() *Table {
	return &Table{
		Model:          NewModel(),
		columns:        make([]TableColumn, 0),
		rows:           make([]TableRow, 0),
		selectedRow:    0,
		selectedCol:    0,
		showHeader:     true,
		showRowNumbers: false,
		borderStyle:    BorderSimple,
		style:          terminus.NewStyle(),
		headerStyle:    terminus.NewStyle().Bold(true),
		selectedStyle:  terminus.NewStyle().Reverse(true),
		rowNumberStyle: terminus.NewStyle().Faint(true),
		sortColumn:     -1,
		sortOrder:      SortNone,
		cellSelection:  false,
	}
}

// SetColumns sets the table columns
func (t *Table) SetColumns(columns []TableColumn) *Table {
	t.columns = columns
	// Adjust selected column if necessary
	if t.selectedCol >= len(t.columns) {
		t.selectedCol = len(t.columns) - 1
	}
	if t.selectedCol < 0 && len(t.columns) > 0 {
		t.selectedCol = 0
	}
	return t
}

// AddColumn adds a single column
func (t *Table) AddColumn(column TableColumn) *Table {
	t.columns = append(t.columns, column)
	return t
}

// SetRows sets the table rows
func (t *Table) SetRows(rows []TableRow) *Table {
	t.rows = rows
	// Adjust selected row if necessary
	if t.selectedRow >= len(t.rows) {
		t.selectedRow = len(t.rows) - 1
	}
	if t.selectedRow < 0 && len(t.rows) > 0 {
		t.selectedRow = 0
	}
	return t
}

// AddRow adds a single row
func (t *Table) AddRow(row TableRow) *Table {
	t.rows = append(t.rows, row)
	return t
}

// SetStringData is a convenience method for setting string data
func (t *Table) SetStringData(headers []string, data [][]string) *Table {
	// Set up columns
	columns := make([]TableColumn, len(headers))
	for i, header := range headers {
		columns[i] = TableColumn{
			Title:    header,
			Width:    15, // Default width
			MinWidth: 5,
			MaxWidth: 50,
			Sortable: true,
			Align:    AlignLeft,
		}
	}
	t.SetColumns(columns)

	// Set up rows
	rows := make([]TableRow, len(data))
	for i, rowData := range data {
		row := make(TableRow, len(rowData))
		for j, cell := range rowData {
			row[j] = NewSimpleTableCell(cell)
		}
		rows[i] = row
	}
	t.SetRows(rows)

	return t
}

// SetShowHeader sets whether to show the header row
func (t *Table) SetShowHeader(show bool) *Table {
	t.showHeader = show
	return t
}

// SetShowRowNumbers sets whether to show row numbers
func (t *Table) SetShowRowNumbers(show bool) *Table {
	t.showRowNumbers = show
	return t
}

// SetBorderStyle sets the border style
func (t *Table) SetBorderStyle(style BorderStyle) *Table {
	t.borderStyle = style
	return t
}

// SetCellSelection enables/disables individual cell selection
func (t *Table) SetCellSelection(enabled bool) *Table {
	t.cellSelection = enabled
	return t
}

// SetStyle sets the default style
func (t *Table) SetStyle(style terminus.Style) *Table {
	t.style = style
	return t
}

// SetHeaderStyle sets the header style
func (t *Table) SetHeaderStyle(style terminus.Style) *Table {
	t.headerStyle = style
	return t
}

// SetSelectedStyle sets the selected cell/row style
func (t *Table) SetSelectedStyle(style terminus.Style) *Table {
	t.selectedStyle = style
	return t
}

// SetRowNumberStyle sets the row number style
func (t *Table) SetRowNumberStyle(style terminus.Style) *Table {
	t.rowNumberStyle = style
	return t
}

// SetOnSelect sets the selection callback
func (t *Table) SetOnSelect(callback func(row, col int, cell TableCell) terminus.Cmd) *Table {
	t.onSelect = callback
	return t
}

// SetOnSort sets the sort callback
func (t *Table) SetOnSort(callback func(column int, order SortOrder) terminus.Cmd) *Table {
	t.onSort = callback
	return t
}

// SelectedRow returns the selected row index
func (t *Table) SelectedRow() int {
	return t.selectedRow
}

// SelectedCol returns the selected column index
func (t *Table) SelectedCol() int {
	return t.selectedCol
}

// SelectedCell returns the selected cell
func (t *Table) SelectedCell() TableCell {
	if t.selectedRow >= 0 && t.selectedRow < len(t.rows) &&
		t.selectedCol >= 0 && t.selectedCol < len(t.rows[t.selectedRow]) {
		return t.rows[t.selectedRow][t.selectedCol]
	}
	return nil
}

// SetSelected sets the selected row and column
func (t *Table) SetSelected(row, col int) *Table {
	if row >= 0 && row < len(t.rows) {
		t.selectedRow = row
	}
	if col >= 0 && col < len(t.columns) {
		t.selectedCol = col
	}
	t.updateScrollOffset()
	return t
}

// SortByColumn sorts the table by the specified column
func (t *Table) SortByColumn(column int, order SortOrder) *Table {
	if column < 0 || column >= len(t.columns) || !t.columns[column].Sortable {
		return t
	}

	if order == SortNone {
		// Cycle through sort orders
		if t.sortColumn == column {
			switch t.sortOrder {
			case SortNone:
				order = SortAsc
			case SortAsc:
				order = SortDesc
			case SortDesc:
				order = SortNone
			}
		} else {
			order = SortAsc
		}
	}

	t.sortColumn = column
	t.sortOrder = order

	if order != SortNone {
		// Remember current selection
		var selectedCell TableCell
		if t.selectedRow >= 0 && t.selectedRow < len(t.rows) &&
			t.selectedCol >= 0 && t.selectedCol < len(t.rows[t.selectedRow]) {
			selectedCell = t.rows[t.selectedRow][t.selectedCol]
		}

		// Sort the rows
		sort.Slice(t.rows, func(i, j int) bool {
			if column >= len(t.rows[i]) || column >= len(t.rows[j]) {
				return false
			}

			cell1 := t.rows[i][column]
			cell2 := t.rows[j][column]

			// Compare values
			val1 := cell1.Value()
			val2 := cell2.Value()

			// Try to compare as strings
			str1 := fmt.Sprintf("%v", val1)
			str2 := fmt.Sprintf("%v", val2)

			result := strings.Compare(str1, str2)
			if order == SortDesc {
				result = -result
			}
			return result < 0
		})

		// Try to restore selection
		if selectedCell != nil {
			for i, row := range t.rows {
				if t.selectedCol < len(row) && row[t.selectedCol] == selectedCell {
					t.selectedRow = i
					break
				}
			}
		}
	}

	return t
}

// updateScrollOffset updates scroll offsets based on selection
func (t *Table) updateScrollOffset() {
	// Vertical scrolling
	visibleRows := t.height
	if t.showHeader {
		visibleRows--
	}

	if t.selectedRow < t.scrollOffsetY {
		t.scrollOffsetY = t.selectedRow
	} else if t.selectedRow >= t.scrollOffsetY+visibleRows {
		t.scrollOffsetY = t.selectedRow - visibleRows + 1
	}

	if t.scrollOffsetY < 0 {
		t.scrollOffsetY = 0
	}
	if t.scrollOffsetY > len(t.rows)-visibleRows {
		t.scrollOffsetY = len(t.rows) - visibleRows
		if t.scrollOffsetY < 0 {
			t.scrollOffsetY = 0
		}
	}

	// Horizontal scrolling (simplified - would need column width calculation)
	// For now, just ensure selected column is visible
	if t.selectedCol < t.scrollOffsetX {
		t.scrollOffsetX = t.selectedCol
	}
	// TODO: Implement proper horizontal scrolling based on column widths
}

// Init implements the Component interface
func (t *Table) Init() terminus.Cmd {
	return nil
}

// Update implements the Component interface
func (t *Table) Update(msg terminus.Msg) (terminus.Component, terminus.Cmd) {
	if !t.Focused() {
		return t, nil
	}

	var cmd terminus.Cmd

	switch msg := msg.(type) {
	case terminus.KeyMsg:
		switch msg.Type {
		case terminus.KeyUp:
			if t.selectedRow > 0 {
				t.selectedRow--
				t.updateScrollOffset()
			}

		case terminus.KeyDown:
			if t.selectedRow < len(t.rows)-1 {
				t.selectedRow++
				t.updateScrollOffset()
			}

		case terminus.KeyLeft:
			if t.cellSelection && t.selectedCol > 0 {
				t.selectedCol--
				t.updateScrollOffset()
			}

		case terminus.KeyRight:
			if t.cellSelection && t.selectedCol < len(t.columns)-1 {
				t.selectedCol++
				t.updateScrollOffset()
			}

		case terminus.KeyHome:
			t.selectedRow = 0
			if t.cellSelection {
				t.selectedCol = 0
			}
			t.updateScrollOffset()

		case terminus.KeyEnd:
			if len(t.rows) > 0 {
				t.selectedRow = len(t.rows) - 1
			}
			if t.cellSelection && len(t.columns) > 0 {
				t.selectedCol = len(t.columns) - 1
			}
			t.updateScrollOffset()

		case terminus.KeyEnter:
			if t.onSelect != nil {
				cmd = t.onSelect(t.selectedRow, t.selectedCol, t.SelectedCell())
			}

		case terminus.KeyRunes:
			if len(msg.Runes) > 0 {
				switch msg.Runes[0] {
				case 's', 'S':
					// Sort by current column
					if t.selectedCol >= 0 && t.selectedCol < len(t.columns) {
						t.SortByColumn(t.selectedCol, SortNone)
						if t.onSort != nil {
							cmd = t.onSort(t.selectedCol, t.sortOrder)
						}
					}
				}
			}
		}
	}

	return t, cmd
}

// View implements the Component interface
func (t *Table) View() string {
	if len(t.columns) == 0 {
		return t.style.Render("No columns defined")
	}

	var result strings.Builder

	// Calculate column widths (simplified)
	colWidths := make([]int, len(t.columns))
	for i, col := range t.columns {
		colWidths[i] = col.Width
		if colWidths[i] <= 0 {
			colWidths[i] = 10 // Default width
		}
	}

	rowNumWidth := 0
	if t.showRowNumbers {
		rowNumWidth = len(fmt.Sprintf("%d", len(t.rows))) + 2
	}

	// Render header
	if t.showHeader {
		if t.showRowNumbers {
			result.WriteString(t.rowNumberStyle.Render(fmt.Sprintf("%*s", rowNumWidth, "")))
		}

		for i, col := range t.columns {
			if i > 0 || t.showRowNumbers {
				result.WriteString("|")
			}

			header := col.Title
			if t.sortColumn == i {
				switch t.sortOrder {
				case SortAsc:
					header += " ↑"
				case SortDesc:
					header += " ↓"
				}
			}

			header = t.alignText(header, colWidths[i], col.Align)
			result.WriteString(t.headerStyle.Render(header))
		}
		result.WriteString("\n")

		// Header separator
		if t.showRowNumbers {
			result.WriteString(strings.Repeat("-", rowNumWidth))
		}
		for i := range t.columns {
			if i > 0 || t.showRowNumbers {
				result.WriteString("+")
			}
			result.WriteString(strings.Repeat("-", colWidths[i]))
		}
		result.WriteString("\n")
	}

	// Calculate visible rows
	visibleRows := t.height
	if t.showHeader {
		visibleRows -= 2 // Header + separator
	}

	// Render visible rows
	start := t.scrollOffsetY
	end := start + visibleRows
	if end > len(t.rows) {
		end = len(t.rows)
	}

	for rowIdx := start; rowIdx < end; rowIdx++ {
		if rowIdx > start {
			result.WriteString("\n")
		}

		row := t.rows[rowIdx]
		isSelected := (rowIdx == t.selectedRow)

		// Row number
		if t.showRowNumbers {
			rowNum := fmt.Sprintf("%*d ", rowNumWidth-1, rowIdx+1)
			if isSelected && !t.cellSelection {
				rowNum = t.selectedStyle.Render(rowNum)
			} else {
				rowNum = t.rowNumberStyle.Render(rowNum)
			}
			result.WriteString(rowNum)
		}

		// Cells
		for colIdx, col := range t.columns {
			if colIdx > 0 || t.showRowNumbers {
				result.WriteString("|")
			}

			var cellText string
			if colIdx < len(row) {
				cellText = row[colIdx].Render()
			}

			cellText = t.alignText(cellText, colWidths[colIdx], col.Align)

			// Apply styling
			if isSelected && (t.cellSelection && colIdx == t.selectedCol || !t.cellSelection) {
				cellText = t.selectedStyle.Render(cellText)
			} else {
				cellText = t.style.Render(cellText)
			}

			result.WriteString(cellText)
		}
	}

	// Pad remaining height
	currentLines := strings.Count(result.String(), "\n") + 1
	for currentLines < t.height {
		result.WriteString("\n")
		currentLines++
	}

	return result.String()
}

// alignText aligns text within the given width
func (t *Table) alignText(text string, width int, align Alignment) string {
	if len(text) >= width {
		return text[:width]
	}

	padding := width - len(text)
	switch align {
	case AlignLeft:
		return text + strings.Repeat(" ", padding)
	case AlignRight:
		return strings.Repeat(" ", padding) + text
	case AlignCenter:
		leftPad := padding / 2
		rightPad := padding - leftPad
		return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
	default:
		return text + strings.Repeat(" ", padding)
	}
}

// RowCount returns the number of rows
func (t *Table) RowCount() int {
	return len(t.rows)
}

// ColCount returns the number of columns
func (t *Table) ColCount() int {
	return len(t.columns)
}

// IsEmpty returns whether the table has no data
func (t *Table) IsEmpty() bool {
	return len(t.rows) == 0 || len(t.columns) == 0
}