package eventgrid

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gobuffalo/buffalo"
)

// TypeDispatchSubscriber offers an indirection for calling a function when
// an Event Grid Event has a particular value for the property `eventType`.
// While the `EventHandler` interface does not itself has
type TypeDispatchSubscriber struct {
	Subscriber
	bindings          map[string]EventHandler
	normalizeTypeCase bool
}

// NewTypeDispatchSubscriber initializes a new empty TypeDispathSubscriber.
func NewTypeDispatchSubscriber(parent Subscriber) (created *TypeDispatchSubscriber) {
	created = &TypeDispatchSubscriber{
		Subscriber: parent,
		bindings:   make(map[string]EventHandler),
	}

	created.Bind("Microsoft.EventGrid.SubscriptionValidationEvent", ReceiveSubscriptionValidationRequest)

	return
}

// Bind ties together an Event Type identifier string and a function that knows how to handle it.
func (s *TypeDispatchSubscriber) Bind(eventType string, handler EventHandler) *TypeDispatchSubscriber {
	s.bindings[s.NormalizeEventType(eventType)] = handler
	return s
}

// Unbind removes the mapping between an Event Type string and the associated EventHandler, if
// such a mapping exists.
func (s *TypeDispatchSubscriber) Unbind(eventType string) *TypeDispatchSubscriber {
	delete(s.bindings, s.NormalizeEventType(eventType))
	return s
}

// NormalizeEventType applies casing rules
func (s TypeDispatchSubscriber) NormalizeEventType(eventType string) string {
	if s.normalizeTypeCase {
		eventType = strings.ToUpper(eventType)
	}
	return eventType
}

// Receive is `buffalo.Handler` which is called when
func (s TypeDispatchSubscriber) Receive(c buffalo.Context) error {
	var events []Event

	if err := c.Bind(&events); err != nil {
		return c.Error(http.StatusBadRequest, err)
	}

	ctx := NewContext(c)
	var wg sync.WaitGroup
	for _, event := range events {
		wg.Add(1)
		go func() {
			if handler, ok := s.Handler(event.EventType); ok {
				handler(ctx, event)
			} else if handler, ok = s.Handler(EventTypeWildcard); ok {
				handler(ctx, event)
			} else {
				ctx.Error(http.StatusBadRequest, fmt.Errorf("no Handler found for type %q", event.EventType))
			}
			wg.Done()
		}()
	}
	wg.Wait()

	if ctx.HasFailure() {
		return c.Error(http.StatusInternalServerError, errors.New("at least one handler failed to process an event in this batch"))
	}
	c.Response().WriteHeader(http.StatusOK)
	return nil
}

// Handler gets the EventHandler meant to process a particular Event Grid Event Type.
func (s TypeDispatchSubscriber) Handler(eventType string) (handler EventHandler, ok bool) {
	if s.normalizeTypeCase {
		eventType = strings.ToUpper(eventType)
	}
	handler, ok = s.bindings[eventType]
	return
}
