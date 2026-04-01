package execution

import "context"

// Job is the interface for an executable unit of work.
// Implementations: LocalJob (subprocess), RemoteJob (future).
type Job interface {
	// Run executes the job. It blocks until completion or ctx is cancelled.
	// Output lines are sent to the provided output channel.
	Run(ctx context.Context, output chan<- string) error
	// Kill terminates a running job.
	Kill()
	// IsKilled returns true if the job was killed.
	IsKilled() bool
}
