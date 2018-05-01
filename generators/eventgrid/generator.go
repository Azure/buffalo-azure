package eventgrid

import (
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
)

//go:generate go run ./builder/builder.go -o ./static_templates.go ./templates

// Generator will parse an existing `buffalo.App` and add the relevant code
// to make that application be ready for being subscribed to an Event Grid Topic.
type Generator struct {
}

// Run executes the Generator's main purpose, of extending a Buffalo application
// to listen for Event Grid Events.
func (g *Generator) Run() error {
	name, err := ioutil.TempDir("", "buffalo-azure_templates")
	if err != nil {
		return err
	}
	defer func() {
		go os.RemoveAll(name)
	}()

	err = staticTemplates.Rehydrate(name)
	if err != nil {
		return err
	}

	var templateFiles token.FileSet
	_, err = parser.ParseDir(&templateFiles, name, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	return nil
}
