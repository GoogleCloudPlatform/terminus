package terminus

import (
	"context"
	"sync"
)

// Engine manages the MVU (Model-View-Update) lifecycle for a component
type Engine struct {
	component Component
	msgQueue  chan Msg
	processor *CommandProcessor
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	mu        sync.RWMutex
	
	// Callbacks
	onRender func(view string)
	onQuit   func()
}

// NewEngine creates a new MVU engine with the given component
func NewEngine(component Component) *Engine {
	ctx, cancel := context.WithCancel(context.Background())
	e := &Engine{
		component: component,
		msgQueue:  make(chan Msg, 100),
		ctx:       ctx,
		cancel:    cancel,
	}
	
	// Create command processor with callback to send messages
	e.processor = NewCommandProcessor(4, e.SendMessage)
	
	return e
}

// SetRenderCallback sets the function to call when a new view is rendered
func (e *Engine) SetRenderCallback(fn func(view string)) {
	e.onRender = fn
}

// SetQuitCallback sets the function to call when the engine quits
func (e *Engine) SetQuitCallback(fn func()) {
	e.onQuit = fn
}

// Start begins the MVU loop
func (e *Engine) Start() error {
	// Start the command processor
	e.processor.Start()

	// Start the message processor
	e.wg.Add(1)
	go e.processMessages()

	// Initialize the component
	if cmd := e.component.Init(); cmd != nil {
		e.processor.Execute(cmd)
	}

	// Render initial view
	e.render()

	return nil
}

// Stop gracefully shuts down the engine
func (e *Engine) Stop() {
	e.cancel()
	e.processor.Stop()
	e.wg.Wait()
	close(e.msgQueue)
}

// SendMessage sends a message to the component
func (e *Engine) SendMessage(msg Msg) {
	select {
	case e.msgQueue <- msg:
	case <-e.ctx.Done():
	}
}

// processMessages handles the main update loop
func (e *Engine) processMessages() {
	defer e.wg.Done()

	for {
		select {
		case msg, ok := <-e.msgQueue:
			if !ok {
				return
			}

			// Check for quit message
			if _, isQuit := msg.(QuitMsg); isQuit {
				if e.onQuit != nil {
					e.onQuit()
				}
				e.cancel()
				return
			}

			// Update the component
			e.mu.Lock()
			newComponent, cmd := e.component.Update(msg)
			e.component = newComponent
			e.mu.Unlock()

			// Execute any resulting command
			if cmd != nil {
				e.processor.Execute(cmd)
			}

			// Render the new view
			e.render()

		case <-e.ctx.Done():
			return
		}
	}
}


// render calls the view method and invokes the render callback
func (e *Engine) render() {
	e.mu.RLock()
	view := e.component.View()
	e.mu.RUnlock()

	if e.onRender != nil {
		e.onRender(view)
	}
}

