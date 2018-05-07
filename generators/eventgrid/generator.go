package eventgrid

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gobuffalo/buffalo/meta"
	"github.com/markbates/inflect"
)

//go:generate go run ./builder/builder.go -o ./static_templates.go ./templates

// Generator will parse an existing `buffalo.App` and add the relevant code
// to make that application be ready for being subscribed to an Event Grid Topic.
type Generator struct{}

// Run executes the Generator's main purpose, of extending a Buffalo application
// to listen for Event Grid Events.
func (g *Generator) Run(app meta.App, name inflect.Name, types ...string) error {
	subscriberFile, err := os.Create(filepath.Join(app.ActionsPkg, fmt.Sprintf("%s.go", name.File())))
	if err != nil {
		return err
	}
	defer subscriberFile.Close()

	return g.WriteSubscriberFile(subscriberFile, "ingress")
}

// WriteSubscriberFile prints the contents of an EventGridSubscriber with the specified name
func (g *Generator) WriteSubscriberFile(output io.Writer, name inflect.Name) error {
	updatedFiles := token.NewFileSet()
	fast, err := g.loadSubscriberAST()
	if err != nil {
		return err
	}

	name.Camel()

	return printer.Fprint(output, updatedFiles, fast)
}

func (g *Generator) getSubscriberConstructor(template *ast.File, name inflect.Name) (*ast.FuncDecl, error) {
	const constructorName = "NewMyEventGridTopicSubscriber"
	for _, fd := range template.Decls {
		if asFunc, ok := fd.(*ast.FuncDecl); ok {
			if asFunc.Name.Name == constructorName {
				return &ast.FuncDecl{
					Name: &ast.Ident{Name: fmt.Sprintf("New%sSubscriber", name.Camel())},
					Body: asFunc.Body,
					Type: asFunc.Type,
					Recv: asFunc.Recv,
				}, nil
			}
		}
	}
	return nil, errors.New(`constructor "` + constructorName + `" not found in template`)
}

func (g *Generator) getBindingCall(constructor *ast.FuncDecl, typeIdentifier string, typeName inflect.Name) (*ast.CallExpr, error) {
	// TODO https://github.com/Azure/buffalo-azure/issues/21
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "dispatcher"},
			Sel: &ast.Ident{Name: "Bind"},
		},
		Args: []ast.Expr{
			&ast.BasicLit{
				Kind:  token.STRING,
				Value: fmt.Sprintf("%q", typeIdentifier),
			},
			&ast.SelectorExpr{
				X:   &ast.Ident{Name: "created"},
				Sel: &ast.Ident{Name: fmt.Sprintf("Receive%s", typeName.Camel())},
			},
		},
	}, nil
}

func (g *Generator) loadSubscriberAST() (*ast.File, error) {
	name, err := ioutil.TempDir("", "buffalo-azure_templates")
	if err != nil {
		return nil, err
	}
	defer func() {
		go os.RemoveAll(name)
	}()

	err = staticTemplates.Rehydrate(name)
	if err != nil {
		return nil, err
	}

	actionsLocation := filepath.Join(name, "templates", "actions")

	var templateFiles token.FileSet
	pkgs, err := parser.ParseDir(&templateFiles, actionsLocation, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	return pkgs["actions"].Files[filepath.Join(actionsLocation, "eventgrid_name.go")], nil
}
