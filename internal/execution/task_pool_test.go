package execution_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kgory/kirmaphor/internal/execution"
	"github.com/google/uuid"
)

func TestTaskPoolRunsJob(t *testing.T) {
	var ran atomic.Int32
	pool := execution.NewTaskPool(2)
	pool.Start()
	defer pool.Stop()

	for i := 0; i < 3; i++ {
		pool.Enqueue(execution.TaskRequest{
			TaskID: uuid.New(),
			Run: func(ctx context.Context) {
				ran.Add(1)
			},
		})
	}

	time.Sleep(200 * time.Millisecond)
	if ran.Load() != 3 {
		t.Fatalf("expected 3 tasks run, got %d", ran.Load())
	}
}

func TestTaskPoolRespectsConcurrencyLimit(t *testing.T) {
	var concurrent atomic.Int32
	var maxSeen atomic.Int32

	pool := execution.NewTaskPool(2)
	pool.Start()
	defer pool.Stop()

	for i := 0; i < 5; i++ {
		pool.Enqueue(execution.TaskRequest{
			TaskID: uuid.New(),
			Run: func(ctx context.Context) {
				c := concurrent.Add(1)
				if c > maxSeen.Load() {
					maxSeen.Store(c)
				}
				time.Sleep(50 * time.Millisecond)
				concurrent.Add(-1)
			},
		})
	}

	time.Sleep(500 * time.Millisecond)
	if maxSeen.Load() > 2 {
		t.Fatalf("concurrency exceeded limit: max seen = %d", maxSeen.Load())
	}
}
