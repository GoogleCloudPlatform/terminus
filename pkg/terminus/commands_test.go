package terminus

import (
	"sync"
	"testing"
	"time"
)

func TestQuitCommand(t *testing.T) {
	msg := Quit()
	if _, ok := msg.(QuitMsg); !ok {
		t.Error("Quit command should return QuitMsg")
	}
}

func TestBatchCommand(t *testing.T) {
	// For now, we'll test that Batch returns immediately with nil
	// The actual command execution happens asynchronously
	cmd1 := func() Msg { return nil }
	cmd2 := func() Msg { return nil }
	cmd3 := func() Msg { return nil }
	
	batch := Batch(cmd1, cmd2, cmd3)
	msg := batch()
	
	if msg != nil {
		t.Error("Batch should return nil message")
	}
}

func TestBatchWithNilCommands(t *testing.T) {
	// Test that Batch handles nil commands gracefully
	cmd := func() Msg { return nil }
	
	batch := Batch(nil, cmd, nil)
	msg := batch()
	
	if msg != nil {
		t.Error("Batch should return nil message even with nil commands")
	}
}

func TestTickCommand(t *testing.T) {
	start := time.Now()
	duration := 50 * time.Millisecond
	
	cmd := Tick(duration, nil)
	msg := cmd()
	
	elapsed := time.Since(start)
	
	tickMsg, ok := msg.(tickMsg)
	if !ok {
		t.Fatal("Tick should return tickMsg when fn is nil")
	}
	
	if elapsed < duration {
		t.Errorf("Tick returned too early: %v < %v", elapsed, duration)
	}
	
	if tickMsg.Time().Before(start) {
		t.Error("Tick time should be after start time")
	}
}

func TestTickCommandWithFunction(t *testing.T) {
	duration := 50 * time.Millisecond
	called := false
	
	customMsg := struct{ Msg }{}
	
	cmd := Tick(duration, func(t time.Time) Msg {
		called = true
		return customMsg
	})
	
	msg := cmd()
	
	if !called {
		t.Error("Tick function should have been called")
	}
	
	if msg != customMsg {
		t.Error("Tick should return the message from the provided function")
	}
}

func TestTickMsgInterface(t *testing.T) {
	now := time.Now()
	msg := tickMsg{time: now}
	
	// Test that tickMsg implements TickMsg interface
	var tickMsgInterface TickMsg = msg
	
	if tickMsgInterface.Time() != now {
		t.Error("TickMsg.Time() should return the correct time")
	}
}

func TestSequence(t *testing.T) {
	var order []int
	var mu sync.Mutex
	
	cmd1 := func() Msg {
		mu.Lock()
		order = append(order, 1)
		mu.Unlock()
		time.Sleep(20 * time.Millisecond)
		return nil
	}
	
	cmd2 := func() Msg {
		mu.Lock()
		order = append(order, 2)
		mu.Unlock()
		time.Sleep(20 * time.Millisecond)
		return nil
	}
	
	cmd3 := func() Msg {
		mu.Lock()
		order = append(order, 3)
		mu.Unlock()
		return nil
	}
	
	seq := Sequence(cmd1, cmd2, cmd3)
	seq()
	
	mu.Lock()
	defer mu.Unlock()
	
	if len(order) != 3 {
		t.Fatalf("Expected 3 commands to execute, got %d", len(order))
	}
	
	// Check they executed in order
	for i, v := range order {
		if v != i+1 {
			t.Errorf("Expected command %d at position %d, got %d", i+1, i, v)
		}
	}
}

func TestParallel(t *testing.T) {
	var completed sync.WaitGroup
	completed.Add(3)
	
	start := time.Now()
	
	cmd1 := func() Msg {
		time.Sleep(50 * time.Millisecond)
		completed.Done()
		return nil
	}
	
	cmd2 := func() Msg {
		time.Sleep(50 * time.Millisecond)
		completed.Done()
		return nil
	}
	
	cmd3 := func() Msg {
		time.Sleep(50 * time.Millisecond)
		completed.Done()
		return nil
	}
	
	parallel := Parallel(cmd1, cmd2, cmd3)
	
	done := make(chan struct{})
	go func() {
		parallel()
		close(done)
	}()
	
	// Wait for parallel to complete
	<-done
	elapsed := time.Since(start)
	
	// All commands should have completed
	completed.Wait()
	
	// If they ran in parallel, total time should be ~50ms, not 150ms
	if elapsed > 100*time.Millisecond {
		t.Errorf("Commands appear to have run sequentially: %v", elapsed)
	}
}

func TestSequenceWithNil(t *testing.T) {
	executed := false
	
	cmd := func() Msg {
		executed = true
		return nil
	}
	
	seq := Sequence(nil, cmd, nil)
	seq()
	
	if !executed {
		t.Error("Non-nil command should have executed")
	}
}

func TestParallelWithNil(t *testing.T) {
	executed := false
	done := make(chan struct{})
	
	cmd := func() Msg {
		executed = true
		close(done)
		return nil
	}
	
	parallel := Parallel(nil, cmd, nil)
	parallel()
	
	<-done
	
	if !executed {
		t.Error("Non-nil command should have executed")
	}
}