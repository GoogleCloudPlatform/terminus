# TerminusGo Todo List Example

This example demonstrates a fully functional todo list application using the TerminusGo widget library. It showcases advanced widget usage, focus management, and layout utilities.

## Features

- **Add new todos**: Type in the text input and press Enter
- **Toggle completion**: Select a todo with arrow keys and press Enter
- **Delete todos**: Press Delete or 'd' when a todo is selected
- **Filter todos**: Press 1 (All), 2 (Active), or 3 (Completed)
- **Bulk operations**: 
  - Ctrl+A: Toggle all todos
  - Ctrl+K: Clear all completed todos
- **Focus management**: Tab to switch between input field and todo list

## Running the Example

```bash
cd examples/todo
go run main.go
```

Then open http://localhost:8081 in your browser.

## Key Concepts Demonstrated

### 1. Widget Composition
- Uses `TextInput` widget for adding new todos
- Uses `List` widget to display and manage todos
- Custom `TodoItem` struct implements `ListItem` interface

### 2. Focus Management
- `FocusManager` handles Tab navigation between widgets
- Each widget maintains its own focused state
- Keyboard input is routed to the focused widget

### 3. Layout Utilities
- `layout.Center()` for centering title and footer
- `layout.Margin()` for consistent spacing
- Dynamic sizing based on terminal dimensions

### 4. State Management
- Todo items are stored in the component's model
- Filter state determines which todos are displayed
- All state changes trigger re-renders automatically

### 5. Event Handling
- Text input has submit handler for adding todos
- List has select handler for toggling todos
- Global keyboard shortcuts for bulk operations

## Code Structure

- `TodoItem`: Represents a single todo with completion state
- `TodoModel`: Application state including todos and filter mode
- `TodoComponent`: Main component that orchestrates the UI
- Event handlers update state and trigger re-renders

## Customization Ideas

- Add due dates to todos
- Implement todo editing (double-click or 'e' key)
- Add categories or tags
- Persist todos to a file or database
- Add sorting options (by date, alphabetical, etc.)
- Implement undo/redo functionality