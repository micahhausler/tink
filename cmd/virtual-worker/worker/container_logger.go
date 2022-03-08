package worker

import (
	"context"

	"github.com/tinkerbell/tink/cmd/tink-worker/worker"
)

type emptyLogger struct{}

// compile-time type check.
var _ worker.LogCapturer = &emptyLogger{}

func (l *emptyLogger) CaptureLogs(_ context.Context, _ string) {}

// NewEmptyContainerLogger returns an no-op container logger.
func NewEmptyContainerLogger() worker.LogCapturer {
	return &emptyLogger{}
}
