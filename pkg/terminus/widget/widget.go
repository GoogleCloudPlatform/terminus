package widget

import (
	"github.com/yourusername/terminusgo/pkg/terminus"
)

// Widget represents a reusable UI component
type Widget interface {
	terminus.Component
	
	// Focus management
	Focus()
	Blur()
	Focused() bool
	
	// Size management
	SetSize(width, height int)
	GetSize() (width, height int)
	
	// Position management
	SetPosition(x, y int)
	GetPosition() (x, y int)
}

// Model is the base model for all widgets
type Model struct {
	focused  bool
	width    int
	height   int
	x        int
	y        int
	disabled bool
}

// NewModel creates a new base widget model
func NewModel() Model {
	return Model{
		width:  10,
		height: 1,
	}
}

// Focus sets the widget as focused
func (m *Model) Focus() {
	m.focused = true
}

// Blur removes focus from the widget
func (m *Model) Blur() {
	m.focused = false
}

// Focused returns whether the widget is focused
func (m *Model) Focused() bool {
	return m.focused
}

// SetSize sets the widget dimensions
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// GetSize returns the widget dimensions
func (m *Model) GetSize() (width, height int) {
	return m.width, m.height
}

// SetPosition sets the widget position
func (m *Model) SetPosition(x, y int) {
	m.x = x
	m.y = y
}

// GetPosition returns the widget position
func (m *Model) GetPosition() (x, y int) {
	return m.x, m.y
}

// SetDisabled sets the disabled state
func (m *Model) SetDisabled(disabled bool) {
	m.disabled = disabled
}

// Disabled returns whether the widget is disabled
func (m *Model) Disabled() bool {
	return m.disabled
}

// FocusManager manages focus between widgets
type FocusManager struct {
	widgets []Widget
	current int
}

// NewFocusManager creates a new focus manager
func NewFocusManager(widgets ...Widget) *FocusManager {
	fm := &FocusManager{
		widgets: widgets,
		current: -1,
	}
	
	// Focus first widget if available
	if len(widgets) > 0 {
		fm.current = 0
		widgets[0].Focus()
	}
	
	return fm
}

// AddWidget adds a widget to the focus manager
func (fm *FocusManager) AddWidget(w Widget) {
	fm.widgets = append(fm.widgets, w)
	if fm.current == -1 && len(fm.widgets) == 1 {
		fm.current = 0
		w.Focus()
	}
}

// Next moves focus to the next widget
func (fm *FocusManager) Next() {
	if len(fm.widgets) == 0 {
		return
	}
	
	if fm.current >= 0 {
		fm.widgets[fm.current].Blur()
	}
	
	fm.current = (fm.current + 1) % len(fm.widgets)
	fm.widgets[fm.current].Focus()
}

// Previous moves focus to the previous widget
func (fm *FocusManager) Previous() {
	if len(fm.widgets) == 0 {
		return
	}
	
	if fm.current >= 0 {
		fm.widgets[fm.current].Blur()
	}
	
	fm.current = (fm.current - 1 + len(fm.widgets)) % len(fm.widgets)
	fm.widgets[fm.current].Focus()
}

// Current returns the currently focused widget
func (fm *FocusManager) Current() Widget {
	if fm.current >= 0 && fm.current < len(fm.widgets) {
		return fm.widgets[fm.current]
	}
	return nil
}

// HandleKey handles tab navigation between widgets
func (fm *FocusManager) HandleKey(msg terminus.KeyMsg) bool {
	switch msg.Type {
	case terminus.KeyTab:
		if msg.Shift {
			fm.Previous()
		} else {
			fm.Next()
		}
		return true
	}
	return false
}

// Container is a widget that can contain other widgets
type Container struct {
	Model
	children []Widget
	focus    *FocusManager
}

// NewContainer creates a new container widget
func NewContainer() *Container {
	return &Container{
		Model:    NewModel(),
		children: make([]Widget, 0),
		focus:    NewFocusManager(),
	}
}

// AddChild adds a child widget to the container
func (c *Container) AddChild(w Widget) {
	c.children = append(c.children, w)
	c.focus.AddWidget(w)
}

// Children returns the child widgets
func (c *Container) Children() []Widget {
	return c.children
}

// Init implements the Component interface
func (c *Container) Init() terminus.Cmd {
	// Initialize all children
	var cmds []terminus.Cmd
	for _, child := range c.children {
		if cmd := child.Init(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	
	if len(cmds) > 0 {
		return terminus.Batch(cmds...)
	}
	return nil
}

// Update implements the Component interface
func (c *Container) Update(msg terminus.Msg) (terminus.Component, terminus.Cmd) {
	// Handle focus management first
	if keyMsg, ok := msg.(terminus.KeyMsg); ok {
		if c.focus.HandleKey(keyMsg) {
			return c, nil
		}
	}
	
	// Forward message to focused child
	if focused := c.focus.Current(); focused != nil {
		newChild, cmd := focused.Update(msg)
		
		// Update the child in our list
		for i, child := range c.children {
			if child == focused {
				c.children[i] = newChild.(Widget)
				break
			}
		}
		
		return c, cmd
	}
	
	return c, nil
}

// View implements the Component interface
func (c *Container) View() string {
	// Simple vertical layout for now
	result := ""
	for i, child := range c.children {
		if i > 0 {
			result += "\n"
		}
		result += child.View()
	}
	return result
}