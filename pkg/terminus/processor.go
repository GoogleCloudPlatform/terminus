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