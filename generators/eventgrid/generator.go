package eventgrid

import (
	"fmt"
	"path"
	"path/filepath"
	"reflect"
	"sort"

	"github.com/gobuffalo/buffalo/meta"
	"github.com/gobuffalo/makr"
	"github.com/markbates/inflect"
)

//go:generate go run ./builder/builder.go -o ./static_templates.go ./templates

// Generator will parse an existing `buffalo.App` and add the relevant code
// to make that application be ready for being subscribed to an Event Grid Topic.
type Generator struct{}

// Run executes the Generator's main purpose, of extending a Buffalo application
// to listen for Event Grid Events.
func (eg *Generator) Run(app meta.App, name inflect.Name, types map[string]reflect.Type) error {
	type TypeMapping struct {
		Identifier string
		inflect.Name
		PackageSpecifier string
	}
	flatTypes := make([]TypeMapping, 0, len(types))

	for i, n := range types {
		flatTypes = append(flatTypes, TypeMapping{
			Identifier:       i,
			PackageSpecifier: path.Base(n.PkgPath()),
			Name:             inflect.Name(n.Name()),
		})
	}

	// I <3 determinism
	sort.Slice(flatTypes, func(i, j int) bool {
		return flatTypes[i].Identifier < flatTypes[j].Identifier
	})

	eventgridFilepath := filepath.Join(app.ActionsPkg, fmt.Sprintf("%s.go", name.File()))
	g := makr.New()
	defer g.Fmt(app.Root)

	g.Add(makr.NewFile(eventgridFilepath, string(staticTemplates["templates/actions/eventgrid_name.go.tmpl"])))
	d := make(makr.Data)
	d["name"] = name
	d["types"] = flatTypes

	return g.Run(app.Root, d)
}
