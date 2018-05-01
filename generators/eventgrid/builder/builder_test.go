package main

import (
	"context"
	"io/ioutil"
	"testing"
	"time"
)

func TestReadFiles(t *testing.T) {
	discoveredFiles := make(chan file)

	currentDirectory, err := ioutil.ReadDir(".")
	if err != nil {
		t.Error(err)
		return
	}

	expected := map[string]struct{}{}

	for _, item := range currentDirectory {
		expected[item.Name()] = struct{}{}
	}

	finishedCheckingExpected := make(chan struct{})
	go func(ctx context.Context, discoveredFiles <-chan file) {
		defer close(finishedCheckingExpected)
		for {
			select {
			case discovered, ok := <-discoveredFiles:
				if !ok {
					return
				}
				if _, ok = expected[discovered.path]; ok {
					delete(expected, discovered.path)
				} else {
					t.Logf("unexpected file: %q", discovered.path)
					t.Fail()
				}
			case <-ctx.Done():
				return
			}
		}
	}(context.Background(), discoveredFiles)

	if err := readFiles(context.Background(), ".", discoveredFiles); err != nil {
		t.Error(err)
		return
	}
	close(discoveredFiles)

	<-finishedCheckingExpected

	for unfound := range expected {
		t.Logf("expected %q, but it was not found", unfound)
		t.Fail()
	}
}

func TestReadFiles_Cancel(t *testing.T) {
	testControl, testCancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer testCancel()

	ctx, cancel := context.WithCancel(testControl)
	discoveredFiles := make(chan file)
	defer close(discoveredFiles)

	finished := make(chan error)
	go func(ctx context.Context, err chan<- error) {
		err <- readFiles(ctx, ".", discoveredFiles)
	}(ctx, finished)
	cancel()

	select {
	case <-testControl.Done():
		t.Log("test timed out")
		t.Fail()
	case err := <-finished:
		if err != context.Canceled {
			t.Logf("got: %v wanted: %v", err, context.Canceled)
			t.Fail()
		}
	}
}
