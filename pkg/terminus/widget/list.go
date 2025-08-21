// Copyright 2025 The Terminus Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package widget

import (
	"strings"

	"github.com/yourusername/terminusgo/pkg/terminus"
)

// ListItem represents an item in a list
type ListItem interface {
	// Render returns the string representation of the item
	Render() string
	// String returns a simple string representation (for filtering/searching)
	String() string
}

// SimpleListItem is a basic string-based list item
type SimpleListItem struct {
	text string
}

// NewSimpleListItem creates a new simple list item
func NewSimpleListItem(text string) *SimpleListItem {
	return &SimpleListItem{text: text}
}

// Render implements ListItem interface
func (s *SimpleListItem) Render() string {
	return s.text
}

// String implements ListItem interface
func (s *SimpleListItem) String() string {
	return s.text
}

// List is a scrollable list widget
type List struct {
	Model

	// Data
	items        []ListItem
	selectedIdx  int
	scrollOffset int

	// Display settings
	showCursor      bool
	cursorChar      string
	selectedChar    string
	unselectedChar  string

	// Styling
	style              terminus.Style
	selectedStyle      terminus.Style
	cursorStyle        terminus.Style
	selectedCursorStyle terminus.Style

	// Behavior
	wrap bool // Whether to wrap around at top/bottom

	// Events
	onSelect func(int, ListItem) terminus.Cmd
	onChange func(int, ListItem) terminus.Cmd

	// Filtering
	filter         string
	filteredItems  []int // indices of items that match filter
	filteredIdx    int   // selected index in filtered view
}

// NewList creates a new list widget
func NewList() *List {
	return &List{
		Model:               NewModel(),
		selectedIdx:         0,
		showCursor:          true,
		cursorChar:          "> ",
		selectedChar:        "• ",
		unselectedChar:      "  ",
		style:               terminus.NewStyle(),
		selectedStyle:       terminus.NewStyle().Bold(true),
		cursorStyle:         terminus.NewStyle().Foreground(terminus.Cyan),
		selectedCursorStyle: terminus.NewStyle().Foreground(terminus.Cyan).Bold(true),
		wrap:                true,
		filteredItems:       make([]int, 0),
	}
}

// SetItems sets the list items
func (l *List) SetItems(items []ListItem) *List {
	l.items = items
	l.selectedIdx = 0
	l.scrollOffset = 0
	l.updateFiltered()
	return l
}

// AddItem adds a single item to the list
func (l *List) AddItem(item ListItem) *List {
	l.items = append(l.items, item)
	l.updateFiltered()
	return l
}

// SetStringItems is a convenience method for setting string items
func (l *List) SetStringItems(items []string) *List {
	listItems := make([]ListItem, len(items))
	for i, item := range items {
		listItems[i] = NewSimpleListItem(item)
	}
	return l.SetItems(listItems)
}

// Items returns the current items
func (l *List) Items() []ListItem {
	return l.items
}

// SelectedIndex returns the currently selected index in the full list
func (l *List) SelectedIndex() int {
	if l.isFiltered() {
		if l.filteredIdx >= 0 && l.filteredIdx < len(l.filteredItems) {
			return l.filteredItems[l.filteredIdx]
		}
		return -1
	}
	return l.selectedIdx
}

// SelectedItem returns the currently selected item
func (l *List) SelectedItem() ListItem {
	idx := l.SelectedIndex()
	if idx >= 0 && idx < len(l.items) {
		return l.items[idx]
	}
	return nil
}

// SetSelected sets the selected index
func (l *List) SetSelected(index int) *List {
	if index < 0 || index >= len(l.items) {
		return l
	}

	if l.isFiltered() {
		// Find the index in filtered view
		for i, filteredIdx := range l.filteredItems {
			if filteredIdx == index {
				l.filteredIdx = i
				break
			}
		}
	} else {
		l.selectedIdx = index
		l.filteredIdx = index // Keep them in sync when not filtered
	}

	l.updateScrollOffset()
	return l
}

// SetCursorChar sets the cursor character
func (l *List) SetCursorChar(char string) *List {
	l.cursorChar = char
	return l
}

// SetSelectedChar sets the character for selected items
func (l *List) SetSelectedChar(char string) *List {
	l.selectedChar = char
	return l
}

// SetUnselectedChar sets the character for unselected items
func (l *List) SetUnselectedChar(char string) *List {
	l.unselectedChar = char
	return l
}

// SetShowCursor sets whether to show the cursor
func (l *List) SetShowCursor(show bool) *List {
	l.showCursor = show
	return l
}

// SetWrap sets whether to wrap around at top/bottom
func (l *List) SetWrap(wrap bool) *List {
	l.wrap = wrap
	return l
}

// SetStyle sets the default style
func (l *List) SetStyle(style terminus.Style) *List {
	l.style = style
	return l
}

// SetSelectedStyle sets the style for selected items
func (l *List) SetSelectedStyle(style terminus.Style) *List {
	l.selectedStyle = style
	return l
}

// SetCursorStyle sets the cursor style
func (l *List) SetCursorStyle(style terminus.Style) *List {
	l.cursorStyle = style
	return l
}

// SetSelectedCursorStyle sets the cursor style for selected items
func (l *List) SetSelectedCursorStyle(style terminus.Style) *List {
	l.selectedCursorStyle = style
	return l
}

// SetOnSelect sets the selection callback (triggered by Enter)
func (l *List) SetOnSelect(callback func(int, ListItem) terminus.Cmd) *List {
	l.onSelect = callback
	return l
}

// SetOnChange sets the change callback (triggered when selection changes)
func (l *List) SetOnChange(callback func(int, ListItem) terminus.Cmd) *List {
	l.onChange = callback
	return l
}

// SetFilter sets a filter string for the list
func (l *List) SetFilter(filter string) *List {
	l.filter = filter
	l.updateFiltered()
	return l
}

// Filter returns the current filter string
func (l *List) Filter() string {
	return l.filter
}

// isFiltered returns whether filtering is active
func (l *List) isFiltered() bool {
	return l.filter != ""
}

// updateFiltered updates the filtered items list
func (l *List) updateFiltered() {
	l.filteredItems = l.filteredItems[:0] // Clear slice but keep capacity

	if !l.isFiltered() {
		// No filter, show all items
		for i := range l.items {
			l.filteredItems = append(l.filteredItems, i)
		}
		l.filteredIdx = l.selectedIdx
	} else {
		// Apply filter
		filter := strings.ToLower(l.filter)
		currentSelected := -1
		if l.selectedIdx >= 0 && l.selectedIdx < len(l.items) {
			currentSelected = l.selectedIdx
		}
		
		for i, item := range l.items {
			if strings.Contains(strings.ToLower(item.String()), filter) {
				l.filteredItems = append(l.filteredItems, i)
			}
		}
		
		// Try to preserve selection, otherwise reset to first item
		l.filteredIdx = 0
		if currentSelected >= 0 {
			for i, filteredIdx := range l.filteredItems {
				if filteredIdx == currentSelected {
					l.filteredIdx = i
					break
				}
			}
		}
	}

	l.updateScrollOffset()
}

// updateScrollOffset updates the scroll offset based on selection
func (l *List) updateScrollOffset() {
	if len(l.filteredItems) == 0 {
		l.scrollOffset = 0
		return
	}

	currentIdx := l.filteredIdx
	if currentIdx < l.scrollOffset {
		l.scrollOffset = currentIdx
	} else if currentIdx >= l.scrollOffset+l.height {
		l.scrollOffset = currentIdx - l.height + 1
	}

	// Ensure scroll offset is valid
	maxScroll := len(l.filteredItems) - l.height
	if maxScroll < 0 {
		maxScroll = 0
	}
	if l.scrollOffset > maxScroll {
		l.scrollOffset = maxScroll
	}
	if l.scrollOffset < 0 {
		l.scrollOffset = 0
	}
}

// Init implements the Component interface
func (l *List) Init() terminus.Cmd {
	return nil
}

// Update implements the Component interface
func (l *List) Update(msg terminus.Msg) (terminus.Component, terminus.Cmd) {
	if !l.Focused() {
		return l, nil
	}

	var cmd terminus.Cmd

	switch msg := msg.(type) {
	case terminus.KeyMsg:
		switch msg.Type {
		case terminus.KeyUp:
			l.moveUp()
			if l.onChange != nil {
				cmd = l.onChange(l.SelectedIndex(), l.SelectedItem())
			}

		case terminus.KeyDown:
			l.moveDown()
			if l.onChange != nil {
				cmd = l.onChange(l.SelectedIndex(), l.SelectedItem())
			}

		case terminus.KeyHome:
			l.moveToFirst()
			if l.onChange != nil {
				cmd = l.onChange(l.SelectedIndex(), l.SelectedItem())
			}

		case terminus.KeyEnd:
			l.moveToLast()
			if l.onChange != nil {
				cmd = l.onChange(l.SelectedIndex(), l.SelectedItem())
			}

		case terminus.KeyPgUp:
			l.movePageUp()
			if l.onChange != nil {
				cmd = l.onChange(l.SelectedIndex(), l.SelectedItem())
			}

		case terminus.KeyPgDown:
			l.movePageDown()
			if l.onChange != nil {
				cmd = l.onChange(l.SelectedIndex(), l.SelectedItem())
			}

		case terminus.KeyEnter:
			if l.onSelect != nil {
				cmd = l.onSelect(l.SelectedIndex(), l.SelectedItem())
			}
		}
	}

	return l, cmd
}

// moveUp moves selection up one item
func (l *List) moveUp() {
	if len(l.filteredItems) == 0 {
		return
	}

	if l.filteredIdx > 0 {
		l.filteredIdx--
	} else if l.wrap {
		l.filteredIdx = len(l.filteredItems) - 1
	}

	if !l.isFiltered() {
		l.selectedIdx = l.filteredIdx
	}

	l.updateScrollOffset()
}

// moveDown moves selection down one item
func (l *List) moveDown() {
	if len(l.filteredItems) == 0 {
		return
	}

	if l.filteredIdx < len(l.filteredItems)-1 {
		l.filteredIdx++
	} else if l.wrap {
		l.filteredIdx = 0
	}

	if !l.isFiltered() {
		l.selectedIdx = l.filteredIdx
	}

	l.updateScrollOffset()
}

// moveToFirst moves selection to first item
func (l *List) moveToFirst() {
	if len(l.filteredItems) == 0 {
		return
	}

	l.filteredIdx = 0
	if !l.isFiltered() {
		l.selectedIdx = 0
	}
	l.updateScrollOffset()
}

// moveToLast moves selection to last item
func (l *List) moveToLast() {
	if len(l.filteredItems) == 0 {
		return
	}

	l.filteredIdx = len(l.filteredItems) - 1
	if !l.isFiltered() {
		l.selectedIdx = l.filteredIdx
	}
	l.updateScrollOffset()
}

// movePageUp moves selection up one page
func (l *List) movePageUp() {
	if len(l.filteredItems) == 0 {
		return
	}

	l.filteredIdx -= l.height
	if l.filteredIdx < 0 {
		l.filteredIdx = 0
	}

	if !l.isFiltered() {
		l.selectedIdx = l.filteredIdx
	}

	l.updateScrollOffset()
}

// movePageDown moves selection down one page
func (l *List) movePageDown() {
	if len(l.filteredItems) == 0 {
		return
	}

	l.filteredIdx += l.height
	if l.filteredIdx >= len(l.filteredItems) {
		l.filteredIdx = len(l.filteredItems) - 1
	}

	if !l.isFiltered() {
		l.selectedIdx = l.filteredIdx
	}

	l.updateScrollOffset()
}

// View implements the Component interface
func (l *List) View() string {
	if len(l.filteredItems) == 0 {
		if l.isFiltered() {
			return l.style.Render("No items match filter")
		}
		return l.style.Render("No items")
	}

	var result strings.Builder

	// Calculate visible range
	start := l.scrollOffset
	end := start + l.height
	if end > len(l.filteredItems) {
		end = len(l.filteredItems)
	}

	// Render visible items
	for i := start; i < end; i++ {
		if i > start {
			result.WriteString("\n")
		}

		itemIdx := l.filteredItems[i]
		item := l.items[itemIdx]
		isSelected := (i == l.filteredIdx)

		// Build the line
		var line strings.Builder

		// Add cursor or marker
		if l.showCursor && isSelected {
			cursorStyle := l.cursorStyle
			if isSelected {
				cursorStyle = l.selectedCursorStyle
			}
			line.WriteString(cursorStyle.Render(l.cursorChar))
		} else if isSelected {
			line.WriteString(l.selectedChar)
		} else {
			line.WriteString(l.unselectedChar)
		}

		// Add item content
		itemText := item.Render()
		if isSelected {
			itemText = l.selectedStyle.Render(itemText)
		} else {
			itemText = l.style.Render(itemText)
		}
		line.WriteString(itemText)

		// Truncate if too long
		lineStr := line.String()
		if len(lineStr) > l.width {
			// This is a simplified truncation - in reality we'd need to handle ANSI codes properly
			lineStr = lineStr[:l.width-3] + "..."
		}

		result.WriteString(lineStr)
	}

	// Add scroll indicators if needed
	if l.height > 0 {
		totalLines := result.String()
		lines := strings.Split(totalLines, "\n")
		
		// Pad to fill height
		for len(lines) < l.height {
			lines = append(lines, "")
		}

		// Add scroll indicators
		if l.scrollOffset > 0 {
			// Can scroll up
			if len(lines) > 0 {
				lines[0] = l.addScrollIndicator(lines[0], "↑")
			}
		}
		if l.scrollOffset+l.height < len(l.filteredItems) {
			// Can scroll down
			if len(lines) > 0 {
				lines[len(lines)-1] = l.addScrollIndicator(lines[len(lines)-1], "↓")
			}
		}

		result.Reset()
		for i, line := range lines {
			if i > 0 {
				result.WriteString("\n")
			}
			result.WriteString(line)
		}
	}

	return result.String()
}

// addScrollIndicator adds a scroll indicator to the end of a line
func (l *List) addScrollIndicator(line, indicator string) string {
	if len(line) < l.width-1 {
		// Pad the line and add indicator
		padding := l.width - len(line) - 1
		return line + strings.Repeat(" ", padding) + indicator
	}
	if len(line) >= 1 {
		// Replace last character with indicator
		return line[:len(line)-1] + indicator
	}
	return indicator
}

// Len returns the number of items in the list
func (l *List) Len() int {
	return len(l.items)
}

// FilteredLen returns the number of items matching the current filter
func (l *List) FilteredLen() int {
	return len(l.filteredItems)
}

// IsEmpty returns whether the list is empty
func (l *List) IsEmpty() bool {
	return len(l.items) == 0
}