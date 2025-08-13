# Terminus Examples

This directory contains example applications demonstrating various features and patterns in Terminus.

## 🚀 Running Examples

All examples can be run with:

```bash
go run ./examples/<example-name>/
```

Then open your browser to `http://localhost:8890`

## 📚 Example Applications

### Hello World (`hello/`)
The simplest possible Terminus application. Great starting point for beginners.

**Features demonstrated:**
- Basic MVU pattern
- Component lifecycle
- Simple styling
- Keyboard input handling

### Todo List (`todo/`)
A fully functional todo list application with persistence.

**Features demonstrated:**
- TextInput widget for adding todos
- List widget for displaying items
- State management
- Focus switching between widgets
- Data persistence patterns

### Chat Application (`chat/`)
Real-time chat with multiple users.

**Features demonstrated:**
- Message history
- User presence
- Real-time updates
- Scrollable content
- Status indicators

### Dashboard (`dashboard/`)
Complex layout example showing various metrics.

**Features demonstrated:**
- Advanced layout techniques
- Multiple widgets
- Box drawing
- Real-time data updates
- Grid layouts

### Widget Showcase (`widgets/`)
Interactive demonstration of all available widgets.

**Features demonstrated:**
- TextInput with validation
- List with filtering
- Table with sorting
- Spinner animations
- Focus management
- Widget styling

### Text Input Forms (`textinput/`)
Form handling with validation.

**Features demonstrated:**
- Multiple input fields
- Validation logic
- Error display
- Form submission
- Custom validators

### Commands Demo (`commands/`)
Advanced command usage patterns.

**Features demonstrated:**
- HTTP requests
- Timer commands
- Command batching
- Async operations
- Error handling

### Layout System (`layout/`)
Box drawing and layout utilities.

**Features demonstrated:**
- Box borders (single, double, rounded)
- Padding and margins
- Column/row layouts
- Nested layouts
- Responsive design

### Gemini Chat (`gemini_chat/`)
Interactive chat interface for Google's Gemini AI.

**Features demonstrated:**
- AI integration with external APIs
- Async message handling
- Real-time chat interface
- Error handling and recovery
- Environment variable configuration
- Message history with timestamps
- Keyboard shortcuts

## 🎯 Learning Path

1. **Start with Hello World** - Understand the basics
2. **Move to Text Input** - Learn about widgets
3. **Try the Todo List** - See a complete application
4. **Explore Widgets** - Discover all UI components
5. **Study Dashboard** - Learn complex layouts
6. **Build Chat** - Understand real-time features
7. **Review Commands** - Master async operations

## 💡 Common Patterns

### Component Structure
```go
type MyComponent struct {
    // State fields
    count int
    input *widget.TextInput
}

func (m *MyComponent) Init() terminus.Cmd {
    // Initialize
    return nil
}

func (m *MyComponent) Update(msg terminus.Msg) (terminus.Component, terminus.Cmd) {
    // Handle messages
    return m, nil
}

func (m *MyComponent) View() string {
    // Render view
    return "Hello"
}
```

### Handling Input
```go
case terminus.KeyMsg:
    switch msg.Type {
    case terminus.KeyEnter:
        // Submit
    case terminus.KeyEscape:
        return m, terminus.Quit
    case terminus.KeyTab:
        // Switch focus
    }
```

### Using Widgets
```go
// Create
input := widget.NewTextInput().
    SetPlaceholder("Enter text...").
    SetWidth(40)

// Update
input, cmd = input.Update(msg)

// Render
view := input.View()
```

### Styling Text
```go
import "github.com/yourusername/terminusgo/pkg/terminus/style"

styled := style.New().
    Bold(true).
    Foreground(style.Cyan).
    Render("Hello, World!")
```

## 🛠️ Creating Your Own Example

1. Create a new directory under `examples/`
2. Add `main.go` with your component
3. Create `static/` directory
4. Copy standard files:
   - `index.html`
   - `terminus-client.js`
   - `terminus.css`
5. Update this README

## 📝 Best Practices

- Keep examples focused on specific features
- Add comments explaining key concepts
- Use meaningful variable names
- Include keyboard shortcuts in the UI
- Provide visual feedback for actions
- Handle errors gracefully

## 🤝 Contributing

We welcome new examples! Ideas:
- File browser
- Calendar widget
- Data visualization
- Game examples
- API integration demos

Submit a PR with your example!