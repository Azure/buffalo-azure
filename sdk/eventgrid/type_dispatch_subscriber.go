package eventgrid

import (
	"encoding/json"
	"fmt"
	"io"
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
func (s TypeDispatchSubscriber) Receive(c buffalo.Context) (err error) {
	var events []Event
	err = json.NewDecoder(c.Request().Body).Decode(&events)
	if err != nil && err != io.EOF {
		return
	}
	for _, event := range events {
		if handler, ok := s.Handler(event.EventType); ok {
			err = handler(c, event)
		} else if handler, ok = s.Handler(EventTypeWildcard); ok {
			err = handler(c, event)
		} else {
			err = fmt.Errorf("no Handler found for type %q", event.EventType)
		}
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
