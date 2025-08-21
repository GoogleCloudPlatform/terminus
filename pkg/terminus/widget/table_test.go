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

package widget

import (
	"strings"
	"testing"

	"github.com/yourusername/terminusgo/pkg/terminus"
)

func TestSimpleTableCell(t *testing.T) {
	cell := NewSimpleTableCell("test cell")

	if cell.Render() != "test cell" {
		t.Errorf("Expected Render() to return 'test cell', got '%s'", cell.Render())
	}

	if cell.String() != "test cell" {
		t.Errorf("Expected String() to return 'test cell', got '%s'", cell.String())
	}

	if cell.Value() != "test cell" {
		t.Errorf("Expected Value() to return 'test cell', got '%v'", cell.Value())
	}
}

func TestTable(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "Default state",
			test: func(t *testing.T) {
				table := NewTable()

				if table.RowCount() != 0 {
					t.Error("New table should have no rows")
				}

				if table.ColCount() != 0 {
					t.Error("New table should have no columns")
				}

				if !table.IsEmpty() {
					t.Error("New table should be empty")
				}

				if table.SelectedRow() != 0 {
					t.Error("New table should have selectedRow 0")
				}

				if table.SelectedCol() != 0 {
					t.Error("New table should have selectedCol 0")
				}
			},
		},
		{
			name: "Set string data",
			test: func(t *testing.T) {
				table := NewTable()
				headers := []string{"Name", "Age", "City"}
				data := [][]string{
					{"Alice", "25", "New York"},
					{"Bob", "30", "San Francisco"},
					{"Charlie", "35", "Chicago"},
				}
				table.SetStringData(headers, data)

				if table.ColCount() != 3 {
					t.Errorf("Expected 3 columns, got %d", table.ColCount())
				}

				if table.RowCount() != 3 {
					t.Errorf("Expected 3 rows, got %d", table.RowCount())
				}

				if table.IsEmpty() {
					t.Error("Table with data should not be empty")
				}

				// Check selected cell
				selected := table.SelectedCell()
				if selected == nil || selected.String() != "Alice" {
					t.Error("Expected first cell to be selected")
				}
			},
		},
		{
			name: "Navigation - basic movement",
			test: func(t *testing.T) {
				table := NewTable()
				headers := []string{"A", "B", "C"}
				data := [][]string{
					{"1", "2", "3"},
					{"4", "5", "6"},
					{"7", "8", "9"},
				}
				table.SetStringData(headers, data)
				table.Focus()
				table.SetSize(20, 10)

				// Move down
				table.Update(terminus.KeyMsg{Type: terminus.KeyDown})
				if table.SelectedRow() != 1 {
					t.Errorf("Expected selected row 1, got %d", table.SelectedRow())
				}

				// Move down again
				table.Update(terminus.KeyMsg{Type: terminus.KeyDown})
				if table.SelectedRow() != 2 {
					t.Errorf("Expected selected row 2, got %d", table.SelectedRow())
				}

				// Move up
				table.Update(terminus.KeyMsg{Type: terminus.KeyUp})
				if table.SelectedRow() != 1 {
					t.Errorf("Expected selected row 1, got %d", table.SelectedRow())
				}
			},
		},
		{
			name: "Navigation - cell selection",
			test: func(t *testing.T) {
				table := NewTable()
				headers := []string{"A", "B", "C"}
				data := [][]string{
					{"1", "2", "3"},
					{"4", "5", "6"},
				}
				table.SetStringData(headers, data)
				table.SetCellSelection(true)
				table.Focus()

				// Move right
				table.Update(terminus.KeyMsg{Type: terminus.KeyRight})
				if table.SelectedCol() != 1 {
					t.Errorf("Expected selected col 1, got %d", table.SelectedCol())
				}

				// Move right again
				table.Update(terminus.KeyMsg{Type: terminus.KeyRight})
				if table.SelectedCol() != 2 {
					t.Errorf("Expected selected col 2, got %d", table.SelectedCol())
				}

				// Move left
				table.Update(terminus.KeyMsg{Type: terminus.KeyLeft})
				if table.SelectedCol() != 1 {
					t.Errorf("Expected selected col 1, got %d", table.SelectedCol())
				}
			},
		},
		{
			name: "Navigation - boundaries",
			test: func(t *testing.T) {
				table := NewTable()
				headers := []string{"A", "B"}
				data := [][]string{
					{"1", "2"},
					{"3", "4"},
				}
				table.SetStringData(headers, data)
				table.Focus()

				// At first row, move up should stay
				table.Update(terminus.KeyMsg{Type: terminus.KeyUp})
				if table.SelectedRow() != 0 {
					t.Errorf("Expected to stay at row 0, got %d", table.SelectedRow())
				}

				// Move to last row
				table.SetSelected(1, 0)
				// At last row, move down should stay
				table.Update(terminus.KeyMsg{Type: terminus.KeyDown})
				if table.SelectedRow() != 1 {
					t.Errorf("Expected to stay at row 1, got %d", table.SelectedRow())
				}
			},
		},
		{
			name: "Navigation - home and end",
			test: func(t *testing.T) {
				table := NewTable()
				headers := []string{"A", "B", "C"}
				data := [][]string{
					{"1", "2", "3"},
					{"4", "5", "6"},
					{"7", "8", "9"},
				}
				table.SetStringData(headers, data)
				table.SetCellSelection(true)
				table.Focus()
				table.SetSelected(1, 1)

				// Home should go to first row/col
				table.Update(terminus.KeyMsg{Type: terminus.KeyHome})
				if table.SelectedRow() != 0 || table.SelectedCol() != 0 {
					t.Errorf("Expected position (0,0) after Home, got (%d,%d)", table.SelectedRow(), table.SelectedCol())
				}

				// End should go to last row/col
				table.Update(terminus.KeyMsg{Type: terminus.KeyEnd})
				if table.SelectedRow() != 2 || table.SelectedCol() != 2 {
					t.Errorf("Expected position (2,2) after End, got (%d,%d)", table.SelectedRow(), table.SelectedCol())
				}
			},
		},
		{
			name: "Sorting",
			test: func(t *testing.T) {
				table := NewTable()
				headers := []string{"Name", "Age"}
				data := [][]string{
					{"Charlie", "35"},
					{"Alice", "25"},
					{"Bob", "30"},
				}
				table.SetStringData(headers, data)

				// Sort by name (ascending)
				table.SortByColumn(0, SortAsc)
				if table.rows[0][0].String() != "Alice" {
					t.Errorf("Expected first row to be Alice after sort, got %s", table.rows[0][0].String())
				}

				// Sort by name (descending)
				table.SortByColumn(0, SortDesc)
				if table.rows[0][0].String() != "Charlie" {
					t.Errorf("Expected first row to be Charlie after desc sort, got %s", table.rows[0][0].String())
				}
			},
		},
		{
			name: "Sort cycling",
			test: func(t *testing.T) {
				table := NewTable()
				headers := []string{"Name", "Age"}
				data := [][]string{
					{"Bob", "30"},
					{"Alice", "25"},
				}
				table.SetStringData(headers, data)
				table.Focus()

				originalOrder := table.rows[0][0].String()

				// Cycle through sort orders
				table.Update(terminus.KeyMsg{Type: terminus.KeyRunes, Runes: []rune{'s'}})
				// Should be ascending
				if table.rows[0][0].String() == originalOrder {
					t.Error("Expected sort to change order")
				}

				// Press 's' again - should be descending
				table.Update(terminus.KeyMsg{Type: terminus.KeyRunes, Runes: []rune{'s'}})

				// Press 's' again - should be no sort (original order)
				table.Update(terminus.KeyMsg{Type: terminus.KeyRunes, Runes: []rune{'s'}})
				if table.rows[0][0].String() != originalOrder {
					t.Error("Expected to return to original order")
				}
			},
		},
		{
			name: "Events",
			test: func(t *testing.T) {
				table := NewTable()
				headers := []string{"A", "B"}
				data := [][]string{
					{"1", "2"},
					{"3", "4"},
				}
				table.SetStringData(headers, data)
				table.Focus()

				var selectedRow, selectedCol int = -1, -1
				var selectedCell TableCell

				table.SetOnSelect(func(row, col int, cell TableCell) terminus.Cmd {
					selectedRow = row
					selectedCol = col
					selectedCell = cell
					return nil
				})

				// Enter should trigger onSelect
				table.Update(terminus.KeyMsg{Type: terminus.KeyEnter})
				if selectedRow != 0 || selectedCol != 0 || selectedCell.String() != "1" {
					t.Error("onSelect should be triggered on Enter")
				}
			},
		},
		{
			name: "Add column and row",
			test: func(t *testing.T) {
				table := NewTable()

				col := TableColumn{
					Title:    "Test",
					Width:    10,
					Sortable: true,
					Align:    AlignLeft,
				}
				table.AddColumn(col)

				if table.ColCount() != 1 {
					t.Errorf("Expected 1 column, got %d", table.ColCount())
				}

				row := TableRow{NewSimpleTableCell("test")}
				table.AddRow(row)

				if table.RowCount() != 1 {
					t.Errorf("Expected 1 row, got %d", table.RowCount())
				}
			},
		},
		{
			name: "Set selected",
			test: func(t *testing.T) {
				table := NewTable()
				headers := []string{"A", "B", "C"}
				data := [][]string{
					{"1", "2", "3"},
					{"4", "5", "6"},
				}
				table.SetStringData(headers, data)

				table.SetSelected(1, 2)
				if table.SelectedRow() != 1 || table.SelectedCol() != 2 {
					t.Errorf("Expected position (1,2), got (%d,%d)", table.SelectedRow(), table.SelectedCol())
				}

				// Invalid position should be ignored
				table.SetSelected(10, 10)
				if table.SelectedRow() != 1 || table.SelectedCol() != 2 {
					t.Errorf("Expected position to remain (1,2), got (%d,%d)", table.SelectedRow(), table.SelectedCol())
				}
			},
		},
		{
			name: "Unfocused ignores input",
			test: func(t *testing.T) {
				table := NewTable()
				headers := []string{"A", "B"}
				data := [][]string{
					{"1", "2"},
					{"3", "4"},
				}
				table.SetStringData(headers, data)
				// Don't focus the table

				originalRow := table.SelectedRow()
				table.Update(terminus.KeyMsg{Type: terminus.KeyDown})

				if table.SelectedRow() != originalRow {
					t.Error("Unfocused table should ignore input")
				}
			},
		},
		{
			name: "View rendering",
			test: func(t *testing.T) {
				table := NewTable()
				table.SetSize(40, 10)

				// Empty table
				view := table.View()
				if view != table.style.Render("No columns defined") {
					t.Error("Empty table should show 'No columns defined'")
				}

				// Table with data
				headers := []string{"Name", "Age"}
				data := [][]string{
					{"Alice", "25"},
					{"Bob", "30"},
				}
				table.SetStringData(headers, data)

				view = table.View()
				if view == "" {
					t.Error("Table with data should not have empty view")
				}

				// Should contain headers
				if !strings.Contains(view, "Name") || !strings.Contains(view, "Age") {
					t.Error("View should contain column headers")
				}
			},
		},
		{
			name: "Row numbers",
			test: func(t *testing.T) {
				table := NewTable()
				headers := []string{"A"}
				data := [][]string{
					{"1"},
					{"2"},
				}
				table.SetStringData(headers, data)
				table.SetShowRowNumbers(true)
				table.SetSize(20, 10)

				view := table.View()
				if !strings.Contains(view, "1 ") {
					t.Error("View should contain row numbers when enabled")
				}
			},
		},
		{
			name: "Header visibility",
			test: func(t *testing.T) {
				table := NewTable()
				headers := []string{"Name"}
				data := [][]string{{"Alice"}}
				table.SetStringData(headers, data)
				table.SetSize(20, 10)

				// With header
				view := table.View()
				if !strings.Contains(view, "Name") {
					t.Error("View should contain header when showHeader is true")
				}

				// Without header
				table.SetShowHeader(false)
				view = table.View()
				if strings.Contains(view, "Name") {
					t.Error("View should not contain header when showHeader is false")
				}
			},
		},
		{
			name: "Text alignment",
			test: func(t *testing.T) {
				table := NewTable()

				// Test alignment function directly
				leftAligned := table.alignText("test", 10, AlignLeft)
				if leftAligned != "test      " {
					t.Errorf("Expected left aligned 'test      ', got '%s'", leftAligned)
				}

				rightAligned := table.alignText("test", 10, AlignRight)
				if rightAligned != "      test" {
					t.Errorf("Expected right aligned '      test', got '%s'", rightAligned)
				}

				centerAligned := table.alignText("test", 10, AlignCenter)
				if centerAligned != "   test   " {
					t.Errorf("Expected center aligned '   test   ', got '%s'", centerAligned)
				}

				// Test truncation
				truncated := table.alignText("very long text", 5, AlignLeft)
				if truncated != "very " {
					t.Errorf("Expected truncated 'very ', got '%s'", truncated)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

func TestTableChaining(t *testing.T) {
	// Test that all setter methods return *Table for method chaining
	headers := []string{"A", "B"}
	data := [][]string{{"1", "2"}}

	table := NewTable().
		SetStringData(headers, data).
		SetSelected(0, 1).
		SetShowHeader(true).
		SetShowRowNumbers(false).
		SetBorderStyle(BorderSimple).
		SetCellSelection(true).
		SetStyle(terminus.NewStyle()).
		SetHeaderStyle(terminus.NewStyle().Bold(true)).
		SetSelectedStyle(terminus.NewStyle().Reverse(true)).
		SetRowNumberStyle(terminus.NewStyle().Faint(true)).
		SetOnSelect(func(int, int, TableCell) terminus.Cmd { return nil }).
		SetOnSort(func(int, SortOrder) terminus.Cmd { return nil })

	if table.SelectedCol() != 1 {
		t.Error("Method chaining should work correctly")
	}
}