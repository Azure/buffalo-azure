package eventgrid_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/gobuffalo/buffalo"

	"github.com/Azure/buffalo-azure/sdk/eventgrid"
)

func ExampleContext() {
	ctx := eventgrid.NewContext(&buffalo.DefaultContext{})

	var wg sync.WaitGroup

	succeed := func(c buffalo.Context) error {
		defer wg.Done()

		c.Response().WriteHeader(http.StatusOK)

		return nil
	}

	fail := func(c buffalo.Context) error {
		defer wg.Done()
		return c.Error(http.StatusInternalServerError, errors.New("unknown error"))
	}

	wg.Add(3)
	go succeed(ctx)
	go fail(ctx)
	go succeed(ctx)
	wg.Wait()

	fmt.Println(ctx.HasFailure())

	// Output: true
}

type MockContext struct {
	buffalo.Context
	request *http.Request
	*MockResponseWriter
}

func NewMockContext(req *http.Request) *MockContext {
	return &MockContext{
		Context:            &buffalo.DefaultContext{},
		request:            req,
		MockResponseWriter: NewMockResponseWriter(),
	}
}

func (c MockContext) Request() *http.Request {
	return c.request
}

func (c MockContext) Response() http.ResponseWriter {
	return c.MockResponseWriter
}

func (c MockContext) Bind(payload interface{}) error {
	return json.NewDecoder(c.Request().Body).Decode(payload)
}

type MockResponseWriter struct {
	status int
	header http.Header
	body   *bytes.Buffer
}

func NewMockResponseWriter() *MockResponseWriter {
	return &MockResponseWriter{
		body:   bytes.NewBuffer([]byte{}),
		header: make(http.Header),
	}
}

func (w *MockResponseWriter) Header() http.Header {
	return w.header
}

func (w *MockResponseWriter) Write(d []byte) (int, error) {
	return w.body.Write(d)
}

func (w *MockResponseWriter) WriteHeader(s int) {
	w.status = s
}

func (w *MockResponseWriter) Body() io.Reader {
	return w.body
}

func (w *MockResponseWriter) Status() int {
	return w.status
}
