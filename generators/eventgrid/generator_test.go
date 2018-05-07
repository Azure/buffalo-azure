package eventgrid

import (
	"bytes"
	"fmt"
	"go/printer"
	"go/token"
	"testing"

	"github.com/markbates/inflect"
)

func TestGenerator_getSubscriberConstructor(t *testing.T) {
	subject := &Generator{}

	file, err := subject.loadSubscriberAST()
	if err != nil {
		t.Error(err)
		return
	}

	const fakeName = "testName1"

	decl, err := subject.getSubscriberConstructor(file, inflect.Name(fakeName))
	if err != nil {
		t.Error(err)
		return
	}

	if want := fmt.Sprintf("New%sSubscriber", inflect.Name(fakeName).Camel()); decl.Name.Name != want {
		t.Logf("got: %s want: %s", decl.Name.Name, want)
		t.Fail()
		return
	}

	buffer := bytes.NewBuffer([]byte{})
	empty := token.NewFileSet()
	printer.Fprint(buffer, empty, decl)

	t.Logf("Constructor as mutated:\n%s", buffer.String())
}

func TestGenerator_getBindingCall(t *testing.T) {
	g := &Generator{}

	const ident = "Microsoft.BuffaloAzure.FakeType"
	const name = "FakeType"

	result, err := g.getBindingCall(nil, ident, name)
	if err != nil {
		t.Error(err)
		return
	}

	var empty token.FileSet
	output := bytes.NewBuffer([]byte{})
	printer.Fprint(output, &empty, result)

	const want = `dispatcher.Bind("` + ident + `", created.ReceiveFakeType)`
	if got := output.String(); got != want {
		t.Logf("\ngot:\n\t%s\nwant:\n\t%s", got, want)
		t.Fail()
	}
}
