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
)

// CommandProcessor manages concurrent execution of commands
type CommandProcessor struct {
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	workerCount int
	cmdQueue  chan Cmd
	msgSender func(Msg)
}

// NewCommandProcessor creates a new command processor with the specified number of workers
func NewCommandProcessor(workerCount int, msgSender func(Msg)) *CommandProcessor {
	ctx, cancel := context.WithCancel(context.Background())
	return &CommandProcessor{
		ctx:         ctx,
		cancel:      cancel,
		workerCount: workerCount,
		cmdQueue:    make(chan Cmd, 100),
		msgSender:   msgSender,
	}
}

// Start begins processing commands
func (p *CommandProcessor) Start() {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

// Stop gracefully shuts down the processor
func (p *CommandProcessor) Stop() {
	p.cancel()
	close(p.cmdQueue)
	p.wg.Wait()
}

// Execute queues a command for execution
func (p *CommandProcessor) Execute(cmd Cmd) {
	if cmd == nil {
		return
	}
	
	select {
	case p.cmdQueue <- cmd:
	case <-p.ctx.Done():
	}
}

// worker processes commands from the queue
func (p *CommandProcessor) worker() {
	defer p.wg.Done()
	
	for {
		select {
		case cmd, ok := <-p.cmdQueue:
			if !ok {
				return
			}
			
			// Execute the command
			if msg := cmd(); msg != nil && p.msgSender != nil {
				p.msgSender(msg)
			}
			
		case <-p.ctx.Done():
			return
		}
	}
}