package worker

import (
	"context"
	"time"

	bufwork "github.com/gobuffalo/buffalo/worker"
)

// Receiver encapsulates the methods that must be implemented to do processing created by a
// publish to a Worker's queue.
type Receiver interface {
	Start(context.Context) error
	Stop() error
	Register(string, bufwork.Handler) error
}

// Publisher declares all of the functionality a type must have to be able to enqueue Jobs for
// a Buffalo worker to start processing.
type Publisher interface {
	Perform(bufwork.Job) error
	PerformAt(bufwork.Job, time.Time) error
	PerformIn(bufwork.Job, time.Duration) error
}
