package eventgrid

import (
	"net/http"
	"sync"

	"github.com/gobuffalo/buffalo"
)

// SuccessStatusCodes returns an unordered list of HTTP Status Codes
// that should be considered having been handled correctly. Event Grid
// Topics will retry on any HTTP Status Code that is not in this list.
func SuccessStatusCodes() map[int]struct{} {
	return successStatusCodes
}

var successStatusCodes = map[int]struct{}{
	http.StatusOK:      struct{}{},
	http.StatusCreated: struct{}{},
}

type Context struct {
	buffalo.Context
	*ResponseWriter
}

// NewContext initializes a new `eventgrid.Context`.
func NewContext(parent buffalo.Context) *Context {
	return &Context{
		Context:        parent,
		ResponseWriter: NewResponseWriter(),
	}
}

// Response fulfills Buffalo's requirement to allow folks to write a response,
// but it actually just throws away anything you write to it.
func (c *Context) Response() http.ResponseWriter {
	return c.ResponseWriter
}

func (c *Context) Error(status int, err error) error {
	c.WriteHeader(status)
	return c.Context.Error(status, err)
}

// ResponseWriter looks like an `http.ResponseWriter`, but
type ResponseWriter struct {
	sync.RWMutex
	failureSeen bool
	header      http.Header
}

func NewResponseWriter() *ResponseWriter {
	return &ResponseWriter{
		failureSeen: false,
		header:      make(http.Header),
	}
}

// Header gets the Headers associated with this Response writer.
func (w *ResponseWriter) Header() http.Header {
	w.RLock()
	defer w.RUnlock()

	return w.header
}

// Write takes a message to write to a Response, and does nothing with it.
func (w *ResponseWriter) Write(x []byte) (int, error) {
	// Because the body of the response is thrown away, we don't need to
	// put a gaurd around it.
	return len(x), nil
}

// HasFailure evaluates whether or not any Status Headers have been written
// to this Context that are not in the result of calling `SuccessStatusCodes`.
func (w *ResponseWriter) HasFailure() bool {
	w.RLock()
	defer w.RUnlock()

	return w.failureSeen
}

func (w *ResponseWriter) SetFailure() {
	w.Lock()
	defer w.Unlock()

	w.failureSeen = true
}

// WriteHeader takes an HTTP Status Code and informs the `Context` as to whether or
// not there was an error processing it.
func (w *ResponseWriter) WriteHeader(s int) {
	// Because there is only a "set" operation, there is no need to
	// combat the defacto race-condition present here.
	// Because both `HasFailure` and `SetFailure` are protected,
	// this function also will not trigger the `go` tool's race condition
	// detector.
	if _, ok := SuccessStatusCodes()[s]; !w.HasFailure() && !ok {
		w.SetFailure()
	}
}
