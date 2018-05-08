package eventgrid

import (
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/gobuffalo/buffalo/meta"
	"github.com/markbates/inflect"

	"github.com/Azure/buffalo-azure/sdk/eventgrid"
)

func TestGenerator_Run(t *testing.T) {
	subject := Generator{}

	loc, err := ioutil.TempDir("", "buffalo-azure_eventgrid_test")
	if err != nil {
		t.Error(err)
		return
	}
	err = os.MkdirAll(path.Join(loc, "actions"), os.ModePerm)
	if err != nil {
		t.Error(err)
		return
	}
	defer os.RemoveAll(loc)

	t.Log("Output Location: ", loc)

	fakeApp := meta.App{
		Root:       loc,
		ActionsPkg: "actions",
	}

	faux := eventgrid.SubscriptionValidationRequest{}

	if err = subject.Run(fakeApp, inflect.Name("ingress"), map[string]reflect.Type{
		"Microsoft.EventGrid.SubscriptionValidation": reflect.TypeOf(faux),
	}); err != nil {
		t.Error(err)
		return
	}
}
