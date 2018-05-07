package eventgrid_test

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"testing"
	"time"

	"github.com/Azure/buffalo-azure/generators/eventgrid"
)

const defaultTestTimeout = 20 * time.Second

func TestGenerator_WriteSubscriberFile(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()

	generatedFile := bytes.NewBuffer([]byte{})

	errResults := make(chan error)
	go func(err chan<- error) {
		g := eventgrid.Generator{}
		err <- g.WriteSubscriberFile(generatedFile, "ingress")
	}(errResults)

	select {
	case <-ctx.Done():
		if err := ctx.Err(); err != nil {
			t.Error(err)
			return
		}
	case err := <-errResults:
		if err != nil {
			t.Error(err)
			return
		}
	}

	handle, err := ioutil.TempFile("", "buffalo-azure_test")
	if err != nil {
		t.Error(err)
		return
	}
	defer handle.Close()

	_, err = io.Copy(handle, generatedFile)
	if err != nil {
		return
	}
	t.Logf("copied generated file to: %s", handle.Name())
}
