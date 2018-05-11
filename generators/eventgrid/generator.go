package eventgrid

import (
	"fmt"
	"path"
	"path/filepath"
	"reflect"
	"sort"

	"github.com/gobuffalo/buffalo/generators"
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
func (eg *Generator) Run(app meta.App, name string, types map[string]reflect.Type) error {
	iName := inflect.Name(name)
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

	eventgridFilepath := filepath.Join(path.Base(app.ActionsPkg), fmt.Sprintf("%s.go", iName.File()))
	g := makr.New()
	defer g.Fmt(app.Root)

	g.Add(makr.NewFile(eventgridFilepath, string(staticTemplates["templates/actions/eventgrid_name.go.tmpl"])))
	g.Add(&makr.Func{
		Should: func(_ makr.Data) bool { return true },
		Runner: func(root string, data makr.Data) error {
			subName := data["name"].(inflect.Name)
			registrationExpr := fmt.Sprintf(`eventgrid.RegisterSubscriber(app, "/%s", New%sSubscriber(&eventgrid.BaseSubscriber{}))`, subName.Lower(), subName.Camel())
			return generators.AddInsideAppBlock(registrationExpr)
		},
	})

	d := make(makr.Data)
	d["name"] = iName
	d["types"] = flatTypes

	return g.Run(app.Root, d)
}
