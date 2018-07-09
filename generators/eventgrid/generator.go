package eventgrid

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/gobuffalo/buffalo/generators"
	"github.com/gobuffalo/buffalo/meta"
	"github.com/gobuffalo/makr"
	"github.com/markbates/inflect"

	"github.com/Azure/buffalo-azure/generators/common"
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
		PkgPath string
		PkgSpec common.PackageSpecifier
	}
	flatTypes := make([]TypeMapping, 0, len(types))

	ib := common.NewImportBag()
	ib.AddImport("encoding/json")
	ib.AddImport("errors")
	ib.AddImport("net/http")
	ib.AddImportWithSpecifier("github.com/Azure/buffalo-azure/sdk/eventgrid", "eg")
	ib.AddImport("github.com/gobuffalo/buffalo")

	for i, n := range types {
		flatTypes = append(flatTypes, TypeMapping{
			Identifier: i,
			PkgPath:    n.PkgPath(),
			PkgSpec:    ib.AddImport(common.PackagePath(n.PkgPath())),
			Name:       inflect.Name(n.Name()),
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
	d["imports"] = ib.List()

	return g.Run(app.Root, d)
}

func existingImports(filepath string) (retval map[string]string, err error) {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, filepath, nil, parser.ImportsOnly)
	if err != nil {
		return
	}

	retval = make(map[string]string, len(f.Imports))
	for _, imp := range f.Imports {
		pkgPath := strings.Trim(imp.Path.Value, `"`)

		if imp.Name == nil {
			impFinder := importer.Default()
			var pkg *types.Package

			// This downcasting protects this code in the case that it is being called from a compiler
			// other than the official Google Go compiler.
			if cast, ok := impFinder.(types.ImporterFrom); ok {
				pkg, err = cast.ImportFrom(pkgPath, "", 0)
				if err != nil {
					return
				}
				imp.Name = &ast.Ident{Name: pkg.Name()}
			} else {
				pkg, err = impFinder.Import(pkgPath)
				if err != nil {
					return
				}
				imp.Name = &ast.Ident{Name: pkg.Name()}
			}
		}
		retval[imp.Name.Name] = pkgPath
	}
	return
}
