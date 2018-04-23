package eventgrid

import (
	"fmt"
	"strings"

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
	return
}

// Bind ties together an EventType string
func (s *TypeDispatchSubscriber) Bind(eventType string, handler EventHandler) *TypeDispatchSubscriber {
	if s.normalizeTypeCase {
		eventType = strings.ToUpper(eventType)
	}
	s.bindings[eventType] = handler
	return s
}

// Receive is `buffalo.Handler` which is called when
func (s TypeDispatchSubscriber) Receive(c buffalo.Context) (err error) {
	var event Event

	err = c.Bind(&event)
	if err != nil {
		return
	}

	if handler, ok := s.Handler(event.EventType); ok {
		err = handler(c, event)
	} else if handler, ok = s.Handler(EventTypeWildcard); ok {
		err = handler(c, event)
	} else {
		err = fmt.Errorf("no Handler found for type %q", event.EventType)
	}

	return
}

// Handler gets the EventHandler meant to process a particular Event Grid Event Type.
func (s TypeDispatchSubscriber) Handler(eventType string) (handler EventHandler, ok bool) {
	if s.normalizeTypeCase {
		eventType = strings.ToUpper(eventType)
	}
	handler, ok = s.bindings[eventType]
	return
}
