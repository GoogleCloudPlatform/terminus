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
	"sync/atomic"
	"testing"
	"time"
)

func TestWithCancel(t *testing.T) {
	registry := NewCancellationRegistry()
	
	executed := false
	cancelled := false
	
	cmd := registry.WithCancel("test", func(ctx context.Context) Msg {
		select {
		case <-time.After(100 * time.Millisecond):
			executed = true
			return nil
		case <-ctx.Done():
			cancelled = true
			return nil
		}
	})
	
	// Start the command in a goroutine
	go cmd()
	
	// Give it a moment to start
	time.Sleep(10 * time.Millisecond)
	
	// Cancel it
	registry.Cancel("test")
	
	// Wait for completion
	time.Sleep(50 * time.Millisecond)
	
	if executed {
		t.Error("Command should have been cancelled before execution")
	}
	
	if !cancelled {
		t.Error("Command should have detected cancellation")
	}
}

func TestCancelNonExistent(t *testing.T) {
	registry := NewCancellationRegistry()
	
	// Should not panic
	registry.Cancel("non-existent")
}

func TestIsActive(t *testing.T) {
	registry := NewCancellationRegistry()
	
	if registry.IsActive("test") {
		t.Error("Non-existent command should not be active")
	}
	
	// Start a long-running command
	started := make(chan struct{})
	cmd := registry.WithCancel("test", func(ctx context.Context) Msg {
		close(started)
		<-ctx.Done()
		return nil
	})
	
	go cmd()
	<-started // Wait for command to start
	
	if !registry.IsActive("test") {
		t.Error("Running command should be active")
	}
	
	registry.Cancel("test")
	time.Sleep(10 * time.Millisecond)
	
	if registry.IsActive("test") {
		t.Error("Cancelled command should not be active")
	}
}

func TestCancelAll(t *testing.T) {
	registry := NewCancellationRegistry()
	
	var cancelled1, cancelled2 atomic.Bool
	
	cmd1 := registry.WithCancel("cmd1", func(ctx context.Context) Msg {
		<-ctx.Done()
		cancelled1.Store(true)
		return nil
	})
	
	cmd2 := registry.WithCancel("cmd2", func(ctx context.Context) Msg {
		<-ctx.Done()
		cancelled2.Store(true)
		return nil
	})
	
	// Start both commands
	go cmd1()
	go cmd2()
	
	time.Sleep(10 * time.Millisecond)
	
	// Cancel all
	registry.CancelAll()
	
	time.Sleep(50 * time.Millisecond)
	
	if !cancelled1.Load() {
		t.Error("First command should have been cancelled")
	}
	
	if !cancelled2.Load() {
		t.Error("Second command should have been cancelled")
	}
}

func TestTimeout(t *testing.T) {
	t.Run("Command completes before timeout", func(t *testing.T) {
		cmd := Timeout(100*time.Millisecond, func() Msg {
			time.Sleep(10 * time.Millisecond)
			return "completed"
		})
		
		msg := cmd()
		
		if _, ok := msg.(TimeoutMsg); ok {
			t.Error("Command should have completed before timeout")
		}
		
		if msg != "completed" {
			t.Errorf("Expected 'completed', got %v", msg)
		}
	})
	
	t.Run("Command times out", func(t *testing.T) {
		cmd := Timeout(10*time.Millisecond, func() Msg {
			time.Sleep(100 * time.Millisecond)
			return "completed"
		})
		
		msg := cmd()
		
		timeoutMsg, ok := msg.(TimeoutMsg)
		if !ok {
			t.Fatal("Expected TimeoutMsg")
		}
		
		if timeoutMsg.Duration != 10*time.Millisecond {
			t.Errorf("Expected duration 10ms, got %v", timeoutMsg.Duration)
		}
	})
}

func TestDebounce(t *testing.T) {
	var executionCount atomic.Int32
	
	createCmd := func() Cmd {
		return Debounce("test", 50*time.Millisecond, func() Msg {
			executionCount.Add(1)
			return nil
		})
	}
	
	// Call debounce multiple times rapidly
	go createCmd()()
	time.Sleep(10 * time.Millisecond)
	go createCmd()()
	time.Sleep(10 * time.Millisecond)
	go createCmd()()
	
	// Wait for debounce period to expire
	time.Sleep(100 * time.Millisecond)
	
	// Should only execute once
	if count := executionCount.Load(); count != 1 {
		t.Errorf("Expected 1 execution, got %d", count)
	}
}

func TestThrottle(t *testing.T) {
	var executionCount atomic.Int32
	
	cmd := func() Cmd {
		return Throttle("test", 50*time.Millisecond, func() Msg {
			executionCount.Add(1)
			return nil
		})
	}
	
	// First call should execute
	cmd()()
	firstCount := executionCount.Load()
	if firstCount != 1 {
		t.Errorf("Expected 1 execution, got %d", firstCount)
	}
	
	// Immediate second call should be throttled
	cmd()()
	if count := executionCount.Load(); count != 1 {
		t.Errorf("Expected 1 execution (throttled), got %d", count)
	}
	
	// Wait for throttle period
	time.Sleep(60 * time.Millisecond)
	
	// Next call should execute
	cmd()()
	if count := executionCount.Load(); count != 2 {
		t.Errorf("Expected 2 executions after throttle period, got %d", count)
	}
}

func TestGlobalRegistry(t *testing.T) {
	// Test global functions use the global registry
	executed := false
	
	cmd := WithCancel("global-test", func(ctx context.Context) Msg {
		select {
		case <-time.After(100 * time.Millisecond):
			executed = true
		case <-ctx.Done():
			return nil
		}
		return nil
	})
	
	go cmd()
	time.Sleep(10 * time.Millisecond)
	
	Cancel("global-test")
	time.Sleep(50 * time.Millisecond)
	
	if executed {
		t.Error("Global command should have been cancelled")
	}
}