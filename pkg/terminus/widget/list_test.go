package widget

import (
	"fmt"
	"testing"

	"github.com/yourusername/terminusgo/pkg/terminus"
)

func TestSimpleListItem(t *testing.T) {
	item := NewSimpleListItem("test item")
	
	if item.Render() != "test item" {
		t.Errorf("Expected Render() to return 'test item', got '%s'", item.Render())
	}
	
	if item.String() != "test item" {
		t.Errorf("Expected String() to return 'test item', got '%s'", item.String())
	}
}

func TestList(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "Default state",
			test: func(t *testing.T) {
				list := NewList()
				
				if list.Len() != 0 {
					t.Error("New list should be empty")
				}
				
				if !list.IsEmpty() {
					t.Error("New list should report as empty")
				}
				
				if list.SelectedIndex() != 0 {
					t.Error("New list should have selectedIdx 0")
				}
				
				if list.SelectedItem() != nil {
					t.Error("Empty list should return nil for SelectedItem")
				}
			},
		},
		{
			name: "Set string items",
			test: func(t *testing.T) {
				list := NewList()
				items := []string{"item1", "item2", "item3"}
				list.SetStringItems(items)
				
				if list.Len() != 3 {
					t.Errorf("Expected 3 items, got %d", list.Len())
				}
				
				if list.IsEmpty() {
					t.Error("List with items should not be empty")
				}
				
				if list.SelectedIndex() != 0 {
					t.Errorf("Expected selected index 0, got %d", list.SelectedIndex())
				}
				
				selected := list.SelectedItem()
				if selected == nil || selected.String() != "item1" {
					t.Error("Expected first item to be selected")
				}
			},
		},
		{
			name: "Navigation - basic movement",
			test: func(t *testing.T) {
				list := NewList()
				list.SetStringItems([]string{"item1", "item2", "item3"})
				list.Focus()
				list.SetSize(10, 5)
				
				// Move down
				list.Update(terminus.KeyMsg{Type: terminus.KeyDown})
				if list.SelectedIndex() != 1 {
					t.Errorf("Expected selected index 1, got %d", list.SelectedIndex())
				}
				
				// Move down again
				list.Update(terminus.KeyMsg{Type: terminus.KeyDown})
				if list.SelectedIndex() != 2 {
					t.Errorf("Expected selected index 2, got %d", list.SelectedIndex())
				}
				
				// Move up
				list.Update(terminus.KeyMsg{Type: terminus.KeyUp})
				if list.SelectedIndex() != 1 {
					t.Errorf("Expected selected index 1, got %d", list.SelectedIndex())
				}
			},
		},
		{
			name: "Navigation - wrapping",
			test: func(t *testing.T) {
				list := NewList()
				list.SetStringItems([]string{"item1", "item2", "item3"})
				list.Focus()
				list.SetWrap(true)
				
				// At first item, move up should wrap to last
				list.Update(terminus.KeyMsg{Type: terminus.KeyUp})
				if list.SelectedIndex() != 2 {
					t.Errorf("Expected wrapped to index 2, got %d", list.SelectedIndex())
				}
				
				// At last item, move down should wrap to first
				list.Update(terminus.KeyMsg{Type: terminus.KeyDown})
				if list.SelectedIndex() != 0 {
					t.Errorf("Expected wrapped to index 0, got %d", list.SelectedIndex())
				}
			},
		},
		{
			name: "Navigation - no wrapping",
			test: func(t *testing.T) {
				list := NewList()
				list.SetStringItems([]string{"item1", "item2", "item3"})
				list.Focus()
				list.SetWrap(false)
				
				// At first item, move up should stay at first
				list.Update(terminus.KeyMsg{Type: terminus.KeyUp})
				if list.SelectedIndex() != 0 {
					t.Errorf("Expected to stay at index 0, got %d", list.SelectedIndex())
				}
				
				// Move to last item
				list.SetSelected(2)
				
				// At last item, move down should stay at last
				list.Update(terminus.KeyMsg{Type: terminus.KeyDown})
				if list.SelectedIndex() != 2 {
					t.Errorf("Expected to stay at index 2, got %d", list.SelectedIndex())
				}
			},
		},
		{
			name: "Navigation - home and end",
			test: func(t *testing.T) {
				list := NewList()
				list.SetStringItems([]string{"item1", "item2", "item3", "item4", "item5"})
				list.Focus()
				list.SetSelected(2)
				
				// Home should go to first
				list.Update(terminus.KeyMsg{Type: terminus.KeyHome})
				if list.SelectedIndex() != 0 {
					t.Errorf("Expected index 0 after Home, got %d", list.SelectedIndex())
				}
				
				// End should go to last
				list.Update(terminus.KeyMsg{Type: terminus.KeyEnd})
				if list.SelectedIndex() != 4 {
					t.Errorf("Expected index 4 after End, got %d", list.SelectedIndex())
				}
			},
		},
		{
			name: "Navigation - page up/down",
			test: func(t *testing.T) {
				list := NewList()
				items := make([]string, 20)
				for i := range items {
					items[i] = fmt.Sprintf("item%d", i+1)
				}
				list.SetStringItems(items)
				list.Focus()
				list.SetSize(10, 5) // 5 items visible at once
				
				// Page down
				list.Update(terminus.KeyMsg{Type: terminus.KeyPgDown})
				if list.SelectedIndex() != 5 {
					t.Errorf("Expected index 5 after PageDown, got %d", list.SelectedIndex())
				}
				
				// Page up
				list.Update(terminus.KeyMsg{Type: terminus.KeyPgUp})
				if list.SelectedIndex() != 0 {
					t.Errorf("Expected index 0 after PageUp, got %d", list.SelectedIndex())
				}
			},
		},
		{
			name: "Filtering",
			test: func(t *testing.T) {
				list := NewList()
				list.SetStringItems([]string{"apple", "banana", "cherry", "apricot", "blueberry"})
				
				// Filter for items containing "ap"
				list.SetFilter("ap")
				
				if list.FilteredLen() != 2 {
					t.Errorf("Expected 2 filtered items, got %d", list.FilteredLen())
				}
				
				// Should select first filtered item (apple)
				selected := list.SelectedItem()
				if selected == nil || selected.String() != "apple" {
					t.Errorf("Expected 'apple' to be selected, got %v", selected)
				}
				
				// Navigate in filtered view
				list.Focus()
				list.Update(terminus.KeyMsg{Type: terminus.KeyDown})
				selected = list.SelectedItem()
				if selected == nil || selected.String() != "apricot" {
					t.Errorf("Expected 'apricot' to be selected, got %v", selected)
				}
				
				// Clear filter
				list.SetFilter("")
				if list.FilteredLen() != 5 {
					t.Errorf("Expected 5 items after clearing filter, got %d", list.FilteredLen())
				}
			},
		},
		{
			name: "Events",
			test: func(t *testing.T) {
				list := NewList()
				list.SetStringItems([]string{"item1", "item2", "item3"})
				list.Focus()
				
				var selectedIdx int = -1
				var selectedItem ListItem
				var changeIdx int = -1
				var changeItem ListItem
				
				list.SetOnSelect(func(idx int, item ListItem) terminus.Cmd {
					selectedIdx = idx
					selectedItem = item
					return nil
				})
				
				list.SetOnChange(func(idx int, item ListItem) terminus.Cmd {
					changeIdx = idx
					changeItem = item
					return nil
				})
				
				// Move down should trigger onChange
				list.Update(terminus.KeyMsg{Type: terminus.KeyDown})
				if changeIdx != 1 || changeItem.String() != "item2" {
					t.Error("onChange should be triggered on navigation")
				}
				
				// Enter should trigger onSelect
				list.Update(terminus.KeyMsg{Type: terminus.KeyEnter})
				if selectedIdx != 1 || selectedItem.String() != "item2" {
					t.Error("onSelect should be triggered on Enter")
				}
			},
		},
		{
			name: "Add item",
			test: func(t *testing.T) {
				list := NewList()
				list.SetStringItems([]string{"item1", "item2"})
				
				list.AddItem(NewSimpleListItem("item3"))
				
				if list.Len() != 3 {
					t.Errorf("Expected 3 items after adding, got %d", list.Len())
				}
			},
		},
		{
			name: "Set selected",
			test: func(t *testing.T) {
				list := NewList()
				list.SetStringItems([]string{"item1", "item2", "item3"})
				
				list.SetSelected(2)
				if list.SelectedIndex() != 2 {
					t.Errorf("Expected selected index 2, got %d", list.SelectedIndex())
				}
				
				// Invalid index should be ignored
				list.SetSelected(10)
				if list.SelectedIndex() != 2 {
					t.Errorf("Expected selected index to remain 2, got %d", list.SelectedIndex())
				}
			},
		},
		{
			name: "Unfocused ignores input",
			test: func(t *testing.T) {
				list := NewList()
				list.SetStringItems([]string{"item1", "item2", "item3"})
				// Don't focus the list
				
				originalIdx := list.SelectedIndex()
				list.Update(terminus.KeyMsg{Type: terminus.KeyDown})
				
				if list.SelectedIndex() != originalIdx {
					t.Error("Unfocused list should ignore input")
				}
			},
		},
		{
			name: "View rendering",
			test: func(t *testing.T) {
				list := NewList()
				list.SetStringItems([]string{"item1", "item2", "item3"})
				list.SetSize(20, 3)
				
				view := list.View()
				if view == "" {
					t.Error("View should not be empty")
				}
				
				// Empty list should show appropriate message
				emptyList := NewList()
				emptyList.SetSize(20, 3)
				emptyView := emptyList.View()
				if emptyView != emptyList.style.Render("No items") {
					t.Error("Empty list should show 'No items'")
				}
			},
		},
		{
			name: "Scrolling",
			test: func(t *testing.T) {
				list := NewList()
				items := make([]string, 10)
				for i := range items {
					items[i] = fmt.Sprintf("item%d", i+1)
				}
				list.SetStringItems(items)
				list.SetSize(20, 3) // Only 3 items visible
				list.Focus()
				
				// Move to item beyond visible area
				list.SetSelected(5)
				
				// Should automatically scroll
				view := list.View()
				if view == "" {
					t.Error("View should not be empty after scrolling")
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

func TestListChaining(t *testing.T) {
	// Test that all setter methods return *List for method chaining
	list := NewList().
		SetStringItems([]string{"item1", "item2"}).
		SetSelected(1).
		SetCursorChar("→ ").
		SetSelectedChar("★ ").
		SetUnselectedChar("  ").
		SetShowCursor(true).
		SetWrap(false).
		SetStyle(terminus.NewStyle()).
		SetSelectedStyle(terminus.NewStyle().Bold(true)).
		SetCursorStyle(terminus.NewStyle()).
		SetSelectedCursorStyle(terminus.NewStyle()).
		SetOnSelect(func(int, ListItem) terminus.Cmd { return nil }).
		SetOnChange(func(int, ListItem) terminus.Cmd { return nil }).
		SetFilter("item")
	
	if list.SelectedIndex() != 1 {
		t.Error("Method chaining should work correctly")
	}
}