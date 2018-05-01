package eventgrid_test

import (
	"context"
	"testing"
	"time"

	"github.com/Azure/buffalo-azure/generators/eventgrid"
)

func TestGenerator_Run(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	errResults := make(chan error)
	go func(err chan<- error) {
		g := eventgrid.Generator{}
		err <- g.Run()
		close(err)
	}(errResults)

	select {
	case <-ctx.Done():
		if err := ctx.Err(); err != nil {
			t.Error(err)
		}
	case err := <-errResults:
		if err != nil {
			t.Error(err)
		}
	}
}
