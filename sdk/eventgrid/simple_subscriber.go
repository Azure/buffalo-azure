package eventgrid

import (
	"github.com/gobuffalo/buffalo"
)

// SimpleSubscriber only fulfills the "Receive" portion of the Subscriber interface.
// It is equivalent to creating a `TypeDispatchSubscriber` but only binding an `EventHandler`
// to the `EventTypeWildcard` type.
type SimpleSubscriber struct {
	Subscriber
	EventHandler
}

// Receive unmarshals the body of the request as an Event Grid Event, and hands it to the
// EventHandler for further processing.
func (s SimpleSubscriber) Receive(c buffalo.Context) (err error) {
	var event Event

	err = c.Bind(&event)
	if err != nil {
		return
	}

	return s.EventHandler(c, event)
}
