// Package eventgrid aims to provide a shim for easily working with Event Grid
// from a gobuffalo application. It will frequently be utilized by the code
// generated for extending your app to receive Event Grid Events.
package eventgrid

import (
	"path"

	"github.com/gobuffalo/buffalo"
)

// EventTypeWildcard is a special-case value that can be used when subscribing
// to an EventGrid topic.
const EventTypeWildcard = "all"

// App extends the functionality of a normal buffalo.App with actions
// specific to Event Grid. Specifically, it seeks to allow quick and easy
// register a Group of actions for processing and reasoning.
type App buffalo.App

// Subscriber creates a group of mappings (*buffalo.App) between
// a Subscriber interface implementation and the appropriate REST
// paths.
func (a *App) Subscriber(p string, s Subscriber) *buffalo.App {
	g := (*buffalo.App)(a).Group(p)
	p = "/"

	g.POST(p, s.Receive)
	if a.Env == "development" {
		g.GET(p, s.List)
		g.GET(path.Join(p, "new"), s.New)
	}

	return g
}
