package eventgrid

import (
	"github.com/gobuffalo/buffalo"
)

// RegisterSubscriber updates a `buffalo.App` to route requests to a particular
// subscriber.
// This method is the spiritual equivalent of `App.Resource`:
// https://godoc.org/github.com/gobuffalo/buffalo#App.Resource
func RegisterSubscriber(app *buffalo.App, route string, s Subscriber) *buffalo.App {
	group := app.Group(route)

	route = "/"

	group.POST(route, SubscriptionValidationMiddleware(s.Receive))
	group.GET(route, s.List)
	group.GET(route+"{event_id}", s.Show)

	return group
}
