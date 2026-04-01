package execution

import (
	"sync"
	"time"
)

// LogWriter batches output lines and flushes them via the provided function
// at a regular interval or when Close is called.
// This matches Semaphore's 500ms batch pattern.
type LogWriter struct {
	flush    func(lines []string) error
	interval time.Duration
	mu       sync.Mutex
	buffer   []string
	done     chan struct{}
	once     sync.Once
}

func NewLogWriter(flush func(lines []string) error, interval time.Duration) *LogWriter {
	lw := &LogWriter{
		flush:    flush,
		interval: interval,
		done:     make(chan struct{}),
	}
	go lw.run()
	return lw
}

func (lw *LogWriter) Write(line string) {
	lw.mu.Lock()
	lw.buffer = append(lw.buffer, line)
	lw.mu.Unlock()
}

func (lw *LogWriter) flushNow() {
	lw.mu.Lock()
	if len(lw.buffer) == 0 {
		lw.mu.Unlock()
		return
	}
	lines := lw.buffer
	lw.buffer = nil
	lw.mu.Unlock()
	lw.flush(lines) // ignore error — caller can log separately
}

func (lw *LogWriter) run() {
	ticker := time.NewTicker(lw.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			lw.flushNow()
		case <-lw.done:
			lw.flushNow() // final flush
			return
		}
	}
}

// Close flushes remaining lines and stops the background goroutine.
func (lw *LogWriter) Close() {
	lw.once.Do(func() { close(lw.done) })
	// Give the goroutine time to finish final flush
	time.Sleep(10 * time.Millisecond)
}
