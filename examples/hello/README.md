# Hello World Example

This is a simple "Hello World" example that demonstrates the basic concepts of TerminusGo and the Model-View-Update (MVU) architecture.

## What This Example Shows

1. **Basic Component Structure**: How to create a component that implements the `terminus.Component` interface
2. **State Management**: Using a model struct to manage application state
3. **Handling User Input**: Processing keyboard events and updating state accordingly
4. **Styling Text**: Using the `style` package to create visually appealing terminal output
5. **MVU Pattern**: The flow of Init → Update → View in action

## Running the Example

From the project root:

```bash
go run examples/hello/main.go
```

Then open your browser to http://localhost:8080

## How It Works

### The Model

The `HelloModel` struct contains all the application state:
- `greeting`: The initial greeting message
- `name`: The user's name as they type it
- `collectingName`: Whether we're currently collecting the user's name
- `showPersonalized`: Whether to show the personalized greeting

### The Update Function

The `Update` function handles all user input:
- **Enter key**: Starts name collection or submits the entered name
- **Character keys**: Add to the name when collecting input
- **Backspace**: Removes characters from the name
- **Escape**: Resets to the initial state
- **'r' key**: Resets the application
- **'q' or Ctrl+C**: Quits the application

### The View Function

The `View` function renders different UI states:
1. **Initial state**: Shows "Hello, World!" and prompts to press Enter
2. **Name collection**: Shows an input field with a blinking cursor
3. **Personalized greeting**: Shows "Hello, [name]!" with reset instructions

### Key Concepts Demonstrated

1. **Immutable State**: The model is treated as immutable - we modify fields but the Update function returns the component
2. **Message Handling**: Different key types are handled appropriately
3. **Conditional Rendering**: The view changes based on the model state
4. **Styling**: Different text styles are applied using the style package
5. **User Feedback**: Clear instructions and visual feedback (like the blinking cursor)

## Extending This Example

Try these modifications to learn more:
- Add a counter that tracks how many times the user has entered their name
- Add more color schemes or let the user choose colors
- Add validation to the name input (e.g., minimum length)
- Add a list of previous greetings
- Add animation effects using the Tick command