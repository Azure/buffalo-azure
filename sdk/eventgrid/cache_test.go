package eventgrid_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/Azure/buffalo-azure/sdk/eventgrid"
)

func ExampleCache() {
	myCache := &eventgrid.Cache{}

	myCache.Add(eventgrid.Event{
		EventType: "Contoso.Buffalo.CacheProd",
	})
	myCache.Add(eventgrid.Event{
		EventType: "Microsoft.Storage.BlobCreated",
	})

	fmt.Println(myCache.List())

	myCache.Clear()
	fmt.Println(myCache.List())
	// Output:
	// [{   [] Microsoft.Storage.BlobCreated 0001-01-01 00:00:00 +0000 UTC  } {   [] Contoso.Buffalo.CacheProd 0001-01-01 00:00:00 +0000 UTC  }]
	// []
}

func ExampleCache_SetTTL() {
	myCache := &eventgrid.Cache{}
	myCache.SetTTL(time.Second)

	myCache.Add(eventgrid.Event{
		EventType: "Microsoft.Storage.BlobCreated",
	})
	fmt.Println(len(myCache.List()))

	<-time.After(2 * time.Second)
	fmt.Println(len(myCache.List()))

	// Output:
	// 1
	// 0
}

func ExampleCache_SetMaxDepth() {
	myCache := &eventgrid.Cache{}
	myCache.SetMaxDepth(2)

	fmt.Println(len(myCache.List()))
	myCache.Add(eventgrid.Event{})
	fmt.Println(len(myCache.List()))
	myCache.Add(eventgrid.Event{})
	fmt.Println(len(myCache.List()))
	myCache.Add(eventgrid.Event{})
	fmt.Println(len(myCache.List()))

	// Output:
	// 0
	// 1
	// 2
	// 2
}

func TestCache_SetMaxDepth(t *testing.T) {

	myCache := &eventgrid.Cache{}
	myCache.SetMaxDepth(2)

	output := &bytes.Buffer{}

	fmt.Fprint(output, len(myCache.List()))
	myCache.Add(eventgrid.Event{})
	fmt.Fprint(output, len(myCache.List()))
	myCache.Add(eventgrid.Event{})
	fmt.Fprint(output, len(myCache.List()))
	myCache.Add(eventgrid.Event{})
	fmt.Fprint(output, len(myCache.List()))
	fmt.Fprint(output, len(myCache.List()))
	fmt.Fprint(output, len(myCache.List()))

	const want = "012222"
	if got := output.String(); got != want {
		t.Logf("got: %q want: %q", got, want)
		t.Fail()
	}
}
