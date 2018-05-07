package actions

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Azure/buffalo-azure/sdk/eventgrid"
	"github.com/gobuffalo/buffalo"
)

// MyEventGridTopicSubscriber gathers responds to all Requests sent to a particular endpoint.
type MyEventGridTopicSubscriber struct {
	eventgrid.Subscriber
}

// NewMyEventGridTopicSubscriber instantiates MyEventGridTopicSubscriber for use in a `buffalo.App`.
func NewMyEventGridTopicSubscriber(parent eventgrid.Subscriber) (created *MyEventGridTopicSubscriber) {
	dispatcher := eventgrid.NewTypeDispatchSubscriber(parent)

	created = &MyEventGridTopicSubscriber{
		Subscriber: dispatcher,
	}

	return
}

// ReceiveMyType will respond to an `eventgrid.Event` carrying a serialized `MyType` as its payload.
func (s *MyEventGridTopicSubscriber) ReceiveMyType(c buffalo.Context, e eventgrid.Event) error {
	var payload MyType
	if err := json.Unmarshal(e.Data, &payload); err != nil {
		return c.Error(http.StatusBadRequest, errors.New("unable to unmarshal request data"))
	}

	// Replace the code below with your logic
	return nil
}
