package eventgrid

import (
	"net/http"
	"sync"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/pkg/errors"
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

// Context extends `buffalo.Context` to ease communication between a Request Handler
// and an Event Grid Topic.
type Context struct {
	buffalo.Context
	resp  *ResponseWriter
	data  map[string]interface{}
	flash buffalo.Flash
}

// NewContext initializes a new `eventgrid.Context`.
func NewContext(parent buffalo.Context) (created *Context) {
	created = &Context{
		Context: parent,
		resp:    NewResponseWriter(),
		data:    make(map[string]interface{}, len(parent.Data())),
	}

	for k, v := range parent.Data() {
		created.data[k] = v
	}

	return
}

// Response fulfills Buffalo's requirement to allow folks to write a response,
// but it actually just throws away anything you write to it.
func (c *Context) Response() http.ResponseWriter {
	return c.resp
}

// ResponseHasFailure indicates whether or not any Status Codes not indicating
// success to an Event Grid Topic were published to this Context's `ResponseWriter`.
func (c *Context) ResponseHasFailure() bool {
	return c.resp.HasFailure()
}

func (c *Context) Error(status int, err error) error {
	c.resp.WriteHeader(status)
	if logger := c.Logger(); logger != nil {
		logger.Error(err)
	}
	return errors.WithStack(err)
}

// Render discards the body that is populated by the renderer, but takes the status
// into consideration for how to communicate success or failue to an Event Grid Topic.
func (c *Context) Render(status int, r render.Renderer) error {
	c.resp.WriteHeader(status)
	return r.Render(c.Response(), c.Data())
}

// Flash fetches an unused instance of Flash.
func (c *Context) Flash() *buffalo.Flash {
	return &c.flash
}

// Redirect informs the Event Grid Topic that an Event was unable to be handled.
func (c *Context) Redirect(status int, url string, args ...interface{}) error {
	c.resp.WriteHeader(status)
	return nil
}

// ResponseWriter looks like an `http.ResponseWriter`, but
type ResponseWriter struct {
	sync.RWMutex
	failureSeen bool
	header      http.Header
}

// NewResponseWriter initializes a ResponseWriter which will merge the responses of
// several Event Grid Handlers.
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

// SetFailure indicates that a Status Code outside of ones an Event Grid Topic
// accepts as meaning not to retry was present in one of Handlers writing to this
// ResponseWriter.
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
