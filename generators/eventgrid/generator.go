package eventgrid

import (
	"go/parser"
	"go/token"
)

type Generator struct{}

func Run() error {
	files := &token.FileSet{}

	parser.ParseDir(files)
}
