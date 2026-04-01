package execution_test

import (
	"testing"
	"time"
	"github.com/kgory/kirmaphor/internal/execution"
)

func TestLogWriterBatchesLines(t *testing.T) {
	var flushed [][]string
	flush := func(lines []string) error {
		flushed = append(flushed, append([]string{}, lines...))
		return nil
	}

	lw := execution.NewLogWriter(flush, 100*time.Millisecond)
	lw.Write("line 1")
	lw.Write("line 2")
	lw.Write("line 3")

	// Wait for flush interval
	time.Sleep(200 * time.Millisecond)
	lw.Close()

	if len(flushed) == 0 {
		t.Fatal("expected at least one flush")
	}
	total := 0
	for _, batch := range flushed {
		total += len(batch)
	}
	if total != 3 {
		t.Fatalf("expected 3 lines total, got %d", total)
	}
}

func TestLogWriterFlushOnClose(t *testing.T) {
	var flushed [][]string
	flush := func(lines []string) error {
		flushed = append(flushed, append([]string{}, lines...))
		return nil
	}

	lw := execution.NewLogWriter(flush, 10*time.Second) // long interval
	lw.Write("only line")
	lw.Close() // should flush immediately

	if len(flushed) == 0 || flushed[0][0] != "only line" {
		t.Fatal("expected line to be flushed on Close")
	}
}
