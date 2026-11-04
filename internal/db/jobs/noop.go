package jobs

import (
	"context"

	"github.com/riverqueue/river"
)

// NoopArgs is a placeholder job used to satisfy River's requirement that at
// least one worker be registered. It performs no work and should never be
// inserted into the queue in normal operation.
type NoopArgs struct{}

func (NoopArgs) Kind() string { return "noop" }

// NoopWorker processes placeholder noop jobs.
type NoopWorker struct {
	river.WorkerDefaults[NoopArgs]
}

func (w *NoopWorker) Work(_ context.Context, _ *river.Job[NoopArgs]) error {
	return nil
}
