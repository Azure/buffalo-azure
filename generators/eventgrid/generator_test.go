package eventgrid

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/Azure/buffalo-azure/sdk/eventgrid"
	"github.com/gobuffalo/buffalo/meta"
	"os/exec"
	"context"
	"time"
	"path/filepath"
	"path"
	"io"
)

func TestGenerator_Run(t *testing.T) {
	const bufCmd = "buffalo"
	const appName = "gentest"
	if _, err := exec.LookPath(bufCmd); err != nil {
		t.Skipf("%s not found on system", bufCmd)
		return
	}

	subject := Generator{}

	testLoc := path.Join(os.Getenv("GOPATH"), "src")

	loc, err := ioutil.TempDir(testLoc, "buffalo-azure_eventgrid_test")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.RemoveAll(loc)
	t.Log("Output Location: ", loc)

	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Minute)
	defer cancel()

	var outHandle, errHandle io.Writer
	outHandle, err = os.Create(path.Join(loc, "buffalo_stdout.txt"))
	if err != nil {
		t.Logf("not able to harness %s stdout", bufCmd)
		outHandle = ioutil.Discard
	}
	errHandle, err = os.Create(path.Join(loc, "buffalo_stderr.txt"))
	if err != nil {
		t.Logf("not able to harness %s stderr", bufCmd)
		errHandle = ioutil.Discard
	}

	bufCreater := exec.CommandContext(ctx, bufCmd, "new", appName)
	bufCreater.Dir = loc
	bufCreater.Stdout = outHandle
	bufCreater.Stderr = errHandle
	if err := bufCreater.Run(); err != nil {
		t.Error(err)
		return
	}

	fakeApp := meta.App{
		Root:       filepath.Join(loc, appName),
		ActionsPkg: "github.com/marstr/musicvotes/actions",
	}

	faux := eventgrid.SubscriptionValidationRequest{}

	if err = subject.Run(fakeApp, "ingress", map[string]reflect.Type{
		"Microsoft.EventGrid.SubscriptionValidation": reflect.TypeOf(faux),
	}); err != nil {
		t.Error(err)
		return
	}
}
