package execution

import (
	"context"
	"log"
	"sync"

	"github.com/google/uuid"
)

// TaskRequest is a unit of work submitted to the TaskPool.
type TaskRequest struct {
	TaskID uuid.UUID
	Run    func(ctx context.Context)
}

// TaskPool manages a bounded pool of concurrent task runners.
// Pattern adopted from semaphoreui/semaphore: TaskPool.go.
// Uses a buffered channel as the queue and a semaphore channel for concurrency.
type TaskPool struct {
	maxConcurrent int
	queue         chan TaskRequest
	sem           chan struct{}
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewTaskPool(maxConcurrent int) *TaskPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &TaskPool{
		maxConcurrent: maxConcurrent,
		queue:         make(chan TaskRequest, 500), // buffered — Semaphore uses 500
		sem:           make(chan struct{}, maxConcurrent),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start begins processing enqueued tasks.
func (p *TaskPool) Start() {
	go p.dispatch()
}

// Stop cancels the context and waits for running tasks to complete.
func (p *TaskPool) Stop() {
	p.cancel()
	p.wg.Wait()
}

// Enqueue adds a task to the queue. Non-blocking if queue not full.
func (p *TaskPool) Enqueue(req TaskRequest) {
	select {
	case p.queue <- req:
	default:
		log.Printf("taskpool: queue full, dropping task %s", req.TaskID)
	}
}

func (p *TaskPool) dispatch() {
	for {
		select {
		case <-p.ctx.Done():
			return
		case req := <-p.queue:
			p.sem <- struct{}{} // acquire slot
			p.wg.Add(1)
			go func(r TaskRequest) {
				defer func() {
					<-p.sem // release slot
					p.wg.Done()
					if rec := recover(); rec != nil {
						log.Printf("taskpool: panic in task %s: %v", r.TaskID, rec)
					}
				}()
				r.Run(p.ctx)
			}(req)
		}
	}
}
