# Terminus Tutorial

This tutorial will guide you through building your first Terminus application - a simple todo list that runs in the browser with a terminal interface.

## Prerequisites

- Go 1.16 or later
- Basic knowledge of Go programming
- A web browser

## Getting Started

### Step 1: Install Terminus

```bash
go get github.com/yourusername/terminusgo
```

### Step 2: Create Your Project

```bash
mkdir my-todo-app
cd my-todo-app
go mod init my-todo-app
```

### Step 3: Project Structure

Create the following directory structure:

```
my-todo-app/
├── main.go
├── static/
│   ├── index.html
│   ├── terminus-client.js
│   └── terminus.css
└── go.mod
```

## Building a Todo Application

### Step 1: Define Your Model

Create `main.go` and start with defining your data model:

```go
package main

import (
    "embed"
    "fmt"
    "log"
    
    "github.com/yourusername/terminusgo/pkg/terminus"
    "github.com/yourusername/terminusgo/pkg/terminus/style"
    "github.com/yourusername/terminusgo/pkg/terminus/widget"
)

// TodoItem represents a single todo
type TodoItem struct {
    ID        string
    Text      string
    Completed bool
}

// TodoModel represents the application state
type TodoModel struct {
    todos      []TodoItem
    input      *widget.TextInput
    list       *widget.List
    nextID     int
    focusIndex int // 0: input, 1: list
}
```

### Step 2: Create the Component

```go
// TodoComponent implements terminus.Component
type TodoComponent struct {
    model TodoModel
}

// NewTodoComponent creates a new todo component
func NewTodoComponent() *TodoComponent {
    // Create the input field
    input := widget.NewTextInput().
        SetPlaceholder("What needs to be done?").
        SetWidth(50)
    
    return &TodoComponent{
        model: TodoModel{
            todos:      []TodoItem{},
            input:      input,
            list:       widget.NewList([]widget.ListItem{}),
            nextID:     1,
            focusIndex: 0,
        },
    }
}
```

### Step 3: Implement Init

```go
func (t *TodoComponent) Init() terminus.Cmd {
    // Focus on the input field initially
    t.model.input.Focus()
    return nil
}
```

### Step 4: Implement Update

```go
func (t *TodoComponent) Update(msg terminus.Msg) (terminus.Component, terminus.Cmd) {
    switch msg := msg.(type) {
    case terminus.KeyMsg:
        switch msg.Type {
        case terminus.KeyTab:
            // Toggle focus between input and list
            if t.model.focusIndex == 0 {
                t.model.input.Blur()
                t.model.list.Focus()
                t.model.focusIndex = 1
            } else {
                t.model.list.Blur()
                t.model.input.Focus()
                t.model.focusIndex = 0
            }
            return t, nil
            
        case terminus.KeyEscape:
            return t, terminus.Quit
        }
        
        // Route key messages to the focused widget
        if t.model.focusIndex == 0 {
            // Update input
            var cmd terminus.Cmd
            t.model.input, cmd = t.model.input.Update(msg)
            
            // Check if user pressed Enter
            if msg.Type == terminus.KeyEnter {
                text := t.model.input.Value()
                if text != "" {
                    // Add new todo
                    t.addTodo(text)
                    t.model.input.SetValue("")
                }
            }
            
            return t, cmd
        } else {
            // Update list
            var cmd terminus.Cmd
            t.model.list, cmd = t.model.list.Update(msg)
            
            // Check if user pressed Enter to toggle completion
            if msg.Type == terminus.KeyEnter {
                if selected := t.model.list.SelectedItem(); selected != nil {
                    t.toggleTodo(selected.String())
                }
            }
            
            return t, cmd
        }
    }
    
    return t, nil
}
```

### Step 5: Helper Methods

```go
func (t *TodoComponent) addTodo(text string) {
    todo := TodoItem{
        ID:        fmt.Sprintf("%d", t.model.nextID),
        Text:      text,
        Completed: false,
    }
    t.model.todos = append(t.model.todos, todo)
    t.model.nextID++
    t.updateList()
}

func (t *TodoComponent) toggleTodo(id string) {
    for i := range t.model.todos {
        if t.model.todos[i].ID == id {
            t.model.todos[i].Completed = !t.model.todos[i].Completed
            break
        }
    }
    t.updateList()
}

func (t *TodoComponent) updateList() {
    items := make([]widget.ListItem, len(t.model.todos))
    for i, todo := range t.model.todos {
        items[i] = &todoListItem{todo: todo}
    }
    t.model.list.SetItems(items)
}

// todoListItem implements widget.ListItem
type todoListItem struct {
    todo TodoItem
}

func (item *todoListItem) String() string {
    return item.todo.ID
}

func (item *todoListItem) Render(selected bool) string {
    checkbox := "[ ]"
    if item.todo.Completed {
        checkbox = "[✓]"
    }
    
    text := fmt.Sprintf("%s %s", checkbox, item.todo.Text)
    
    if item.todo.Completed {
        text = style.New().
            Foreground(style.Gray).
            Strikethrough(true).
            Render(text)
    }
    
    if selected {
        text = style.New().
            Background(style.Blue).
            Foreground(style.White).
            Render(text)
    }
    
    return text
}

func (item *todoListItem) FilterValue() string {
    return item.todo.Text
}
```

### Step 6: Implement View

```go
func (t *TodoComponent) View() string {
    title := style.New().
        Bold(true).
        Foreground(style.Cyan).
        Render("TODO LIST")
    
    help := style.New().
        Faint(true).
        Render("Tab: switch focus | Enter: add/toggle | Esc: quit")
    
    stats := fmt.Sprintf("Total: %d | Completed: %d", 
        len(t.model.todos), 
        t.countCompleted())
    
    return fmt.Sprintf(`%s

%s

%s

%s

%s`, title, t.model.input.View(), t.model.list.View(), stats, help)
}

func (t *TodoComponent) countCompleted() int {
    count := 0
    for _, todo := range t.model.todos {
        if todo.Completed {
            count++
        }
    }
    return count
}
```

### Step 7: Create the Main Function

```go
//go:embed all:static/*
var staticFiles embed.FS

func main() {
    program := terminus.NewProgram(
        func() terminus.Component {
            return NewTodoComponent()
        },
        terminus.WithAddress(":8080"),
        terminus.WithStaticFiles(staticFiles, "static"),
    )
    
    fmt.Println("Starting Todo app on http://localhost:8080")
    
    if err := program.Start(); err != nil {
        log.Fatal(err)
    }
    
    program.Wait()
}
```

### Step 8: Create Static Files

Create `static/index.html`:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Todo List - Terminus</title>
    <link rel="stylesheet" href="/terminus.css">
</head>
<body>
    <div id="terminal-container">
        <div id="terminal" class="terminal" tabindex="0"></div>
    </div>
    <script src="/terminus-client.js"></script>
</body>
</html>
```

Copy the `terminus-client.js` and `terminus.css` files from the Terminus repository to your `static` directory.

### Step 9: Run Your Application

```bash
go run main.go
```

Open your browser and navigate to `http://localhost:8080`. You should see your todo list application running in a terminal interface!

## Advanced Features

### Adding Persistence

You can add persistence by saving todos to a file:

```go
import (
    "encoding/json"
    "os"
)

func (t *TodoComponent) saveTodos() error {
    data, err := json.Marshal(t.model.todos)
    if err != nil {
        return err
    }
    return os.WriteFile("todos.json", data, 0644)
}

func (t *TodoComponent) loadTodos() error {
    data, err := os.ReadFile("todos.json")
    if err != nil {
        if os.IsNotExist(err) {
            return nil // No file yet
        }
        return err
    }
    return json.Unmarshal(data, &t.model.todos)
}
```

### Adding Filters

Create a filter to show all, active, or completed todos:

```go
type FilterType int

const (
    FilterAll FilterType = iota
    FilterActive
    FilterCompleted
)

// Add to TodoModel
filter FilterType

// Update your list items based on filter
func (t *TodoComponent) updateList() {
    var items []widget.ListItem
    for _, todo := range t.model.todos {
        switch t.model.filter {
        case FilterActive:
            if todo.Completed {
                continue
            }
        case FilterCompleted:
            if !todo.Completed {
                continue
            }
        }
        items = append(items, &todoListItem{todo: todo})
    }
    t.model.list.SetItems(items)
}
```

### Adding HTTP Sync

Sync todos with a backend API:

```go
func (t *TodoComponent) syncTodos() terminus.Cmd {
    return terminus.Post("https://api.example.com/todos", t.model.todos)
}

// Handle the response in Update
case terminus.HTTPRequestMsg:
    if msg.Error != nil {
        // Handle error
        return t, nil
    }
    // Todos synced successfully
    return t, nil
```

## Best Practices

### 1. Component Organization

Keep components focused and small. Split large components into smaller ones:

```go
// Instead of one large component
type AppComponent struct {
    todos    []Todo
    settings Settings
    profile  Profile
}

// Split into focused components
type TodoListComponent struct { todos []Todo }
type SettingsComponent struct { settings Settings }
type ProfileComponent struct { profile Profile }
```

### 2. Message Design

Create clear, descriptive messages:

```go
// Good
type TodoAddedMsg struct { Text string }
type TodoCompletedMsg struct { ID string }
type FilterChangedMsg struct { Filter FilterType }

// Less clear
type UpdateMsg struct { Type string; Data interface{} }
```

### 3. Error Handling

Always handle errors gracefully:

```go
type ErrorMsg struct { Error error }

func (t *TodoComponent) Update(msg terminus.Msg) (terminus.Component, terminus.Cmd) {
    switch msg := msg.(type) {
    case ErrorMsg:
        t.model.error = msg.Error.Error()
        return t, nil
    }
    // ...
}
```

### 4. Styling Consistency

Create reusable styles:

```go
var (
    titleStyle = style.New().Bold(true).Foreground(style.Cyan)
    errorStyle = style.New().Foreground(style.Red)
    successStyle = style.New().Foreground(style.Green)
)
```

## Next Steps

1. Explore the [Widget Gallery](../examples/widgets) for more UI components
2. Check out the [API Reference](api.md) for detailed documentation
3. Look at the [Example Applications](../examples) for inspiration
4. Join the community and share your creations!

Happy coding with Terminus!