package terminus

// Msg is a marker interface for messages that can be sent to components
type Msg interface{}

// Cmd represents a command that performs side effects and returns a message
type Cmd func() Msg

// Component is the core interface that all UI components must implement
type Component interface {
	// Init is called once when the component is first created
	// It returns an optional command to execute
	Init() Cmd

	// Update handles incoming messages and updates the component state
	// It returns the new component state and an optional command
	Update(msg Msg) (Component, Cmd)

	// View renders the component's current state as a string
	View() string
}