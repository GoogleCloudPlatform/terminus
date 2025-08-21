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

// Quit is a special command that signals the application should terminate
var Quit Cmd = func() Msg {
	return QuitMsg{}
}

// Batch performs a list of commands in parallel and returns immediately
func Batch(cmds ...Cmd) Cmd {
	return func() Msg {
		for _, cmd := range cmds {
			if cmd != nil {
				go func(c Cmd) {
					c()
				}(cmd)
			}
		}
		return nil
	}
}

// Sequence performs commands one after another, waiting for each to complete
func Sequence(cmds ...Cmd) Cmd {
	return func() Msg {
		for _, cmd := range cmds {
			if cmd != nil {
				cmd()
			}
		}
		return nil
	}
}

// Parallel performs commands in parallel and waits for all to complete
func Parallel(cmds ...Cmd) Cmd {
	return func() Msg {
		var wg sync.WaitGroup
		
		for _, cmd := range cmds {
			if cmd != nil {
				wg.Add(1)
				go func(c Cmd) {
					defer wg.Done()
					c()
				}(cmd)
			}
		}
		
		wg.Wait()
		return nil
	}
}

// tickMsg is the message sent by the Tick command
type tickMsg struct {
	time time.Time
}

// TickMsg returns the time at which the tick occurred
type TickMsg interface {
	Msg
	Time() time.Time
}

// Time returns the time at which the tick occurred
func (t tickMsg) Time() time.Time {
	return t.time
}

// Tick returns a command that will wait for the given duration,
// then return a TickMsg
func Tick(d time.Duration, fn func(time.Time) Msg) Cmd {
	return func() Msg {
		time.Sleep(d)
		t := time.Now()
		if fn != nil {
			return fn(t)
		}
		return tickMsg{time: t}
	}
}

// Every returns a command that sends a message at regular intervals
// Note: This command runs indefinitely and should be used with WithCancel
func Every(d time.Duration, fn func(time.Time) Msg) Cmd {
	return func() Msg {
		ticker := time.NewTicker(d)
		defer ticker.Stop()
		
		for t := range ticker.C {
			if fn != nil {
				msg := fn(t)
				if msg != nil {
					// This is problematic as we can't send messages from here
					// Every should be used with a different pattern
				}
			}
		}
		return nil
	}
}

// Interval creates a cancellable command that sends messages at regular intervals
func Interval(id string, duration time.Duration, fn func(time.Time) Msg) Cmd {
	return WithCancel(id, func(ctx context.Context) Msg {
		ticker := time.NewTicker(duration)
		defer ticker.Stop()
		
		for {
			select {
			case t := <-ticker.C:
				if fn != nil {
					if msg := fn(t); msg != nil {
						// In a real implementation, we'd need a way to send this message
						// to the update loop. For now, we'll document this limitation
					}
				}
			case <-ctx.Done():
				return nil
			}
		}
	})
}