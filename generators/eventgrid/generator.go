package eventgrid

import (
	"io/ioutil"
	"os"
)

//go:generate go run ./builder/builder.go -o ./static_templates.go ./templates

type Generator struct {
}

func Run() error {
	name, err := ioutil.TempDir("", "buffalo-azure_templates")
	if err != nil {
		return err
	}
	defer os.RemoveAll(name)

	err = staticTemplates.Rehydrate(name)
	if err != nil {
		return err
	}
	return nil
}
