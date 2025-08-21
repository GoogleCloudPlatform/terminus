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

package terminus

import (
	"context"
	"sync"
	"time"
)

// CancellableCmd represents a command that can be cancelled
type CancellableCmd struct {
	cmd    Cmd
	cancel context.CancelFunc
	done   chan struct{}
}

// CancellationRegistry manages cancellable commands
type CancellationRegistry struct {
	mu       sync.Mutex
	commands map[string]*CancellableCmd
}

// NewCancellationRegistry creates a new cancellation registry
func NewCancellationRegistry() *CancellationRegistry {
	return &CancellationRegistry{
		commands: make(map[string]*CancellableCmd),
	}
}

// globalRegistry is the default cancellation registry
var globalRegistry = NewCancellationRegistry()

// WithCancel creates a cancellable command with a unique ID
func WithCancel(id string, cmd func(ctx context.Context) Msg) Cmd {
	return globalRegistry.WithCancel(id, cmd)
}

// Cancel cancels a command by ID
func Cancel(id string) {
	globalRegistry.Cancel(id)
}

// CancelAll cancels all registered commands
func CancelAll() {
	globalRegistry.CancelAll()
}

// WithCancel creates a cancellable command with a unique ID using this registry
func (r *CancellationRegistry) WithCancel(id string, cmd func(ctx context.Context) Msg) Cmd {
	return func() Msg {
		ctx, cancel := context.WithCancel(context.Background())
		
		cancellable := &CancellableCmd{
			cancel: cancel,
			done:   make(chan struct{}),
		}
		
		// Cancel any existing command with the same ID
		r.Cancel(id)
		
		// Register the new command
		r.mu.Lock()
		r.commands[id] = cancellable
		r.mu.Unlock()
		
		// Run the command
		msg := cmd(ctx)
		
		// Clean up
		r.mu.Lock()
		delete(r.commands, id)
		r.mu.Unlock()
		close(cancellable.done)
		
		return msg
	}
}

// Cancel cancels a command by ID
func (r *CancellationRegistry) Cancel(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if cancellable, exists := r.commands[id]; exists {
		cancellable.cancel()
		delete(r.commands, id)
	}
}

// CancelAll cancels all registered commands
func (r *CancellationRegistry) CancelAll() {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	for id, cancellable := range r.commands {
		cancellable.cancel()
		delete(r.commands, id)
	}
}

// IsActive checks if a command with the given ID is currently running
func (r *CancellationRegistry) IsActive(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	_, exists := r.commands[id]
	return exists
}

// Timeout creates a command that will be automatically cancelled after a duration
func Timeout(d time.Duration, cmd Cmd) Cmd {
	return func() Msg {
		ctx, cancel := context.WithTimeout(context.Background(), d)
		defer cancel()
		
		done := make(chan Msg, 1)
		
		go func() {
			done <- cmd()
		}()
		
		select {
		case msg := <-done:
			return msg
		case <-ctx.Done():
			return TimeoutMsg{Duration: d}
		}
	}
}

// TimeoutMsg is sent when a command times out
type TimeoutMsg struct {
	Duration time.Duration
}

// Debounce creates a command that will only execute after a period of inactivity
func Debounce(id string, delay time.Duration, cmd Cmd) Cmd {
	return WithCancel(id, func(ctx context.Context) Msg {
		select {
		case <-time.After(delay):
			return cmd()
		case <-ctx.Done():
			return nil
		}
	})
}

// Throttle creates a command that will execute at most once per duration
var throttleRegistry = struct {
	sync.Mutex
	lastExecution map[string]time.Time
}{
	lastExecution: make(map[string]time.Time),
}

func Throttle(id string, minInterval time.Duration, cmd Cmd) Cmd {
	return func() Msg {
		throttleRegistry.Lock()
		lastExec, exists := throttleRegistry.lastExecution[id]
		now := time.Now()
		
		if exists && now.Sub(lastExec) < minInterval {
			throttleRegistry.Unlock()
			return nil
		}
		
		throttleRegistry.lastExecution[id] = now
		throttleRegistry.Unlock()
		
		return cmd()
	}
}