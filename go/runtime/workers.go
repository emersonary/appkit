package runtime

import "context"

// RegisterWorker returns a background context cancelled during Application.Shutdown.
// Use it from block wiring or WireServices for long-running goroutines.
// The returned CancelFunc is optional; shutdown cancels every registered worker.
func (a *Application[T]) RegisterWorker() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	a.workerCancels = append(a.workerCancels, cancel)
	return ctx, cancel
}

func (a *Application[T]) stopWorkers() {
	for _, cancel := range a.workerCancels {
		if cancel != nil {
			cancel()
		}
	}
	a.workerCancels = nil
}
