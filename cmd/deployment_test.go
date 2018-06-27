package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func TestDeploymentParameters_UnmarshalFile(t *testing.T) {
	handle, err := os.Open("./testdata/parameters1.json")
	if err != nil {
		t.Error(err)
		return
	}
	defer handle.Close()

	subject := NewDeploymentParameters()
	dec := json.NewDecoder(handle)
	dec.Decode(subject)

	if got, want := len(subject.Parameters), 5; got != want {
		t.Logf("Number Parameters:\n\tgot: %d want: %d", got, want)
		t.Fail()
	}

	expected := map[string]string{
		"database":                   "postgresql",
		"databaseAdministratorLogin": "buffaloAdmin",
		"databaseName":               "buffalo30_development",
		"imageName":                  "marstr/quickstart:latest",
		"name":                       "buffalo-app-uqqq1nfno1",
	}

	for k, v := range expected {
		if val, ok := subject.Parameters[k]; ok {
			if v != val.Value.(string) {
				t.Logf("got: %q want: %q", val, v)
				t.Fail()
			}
		} else {
			t.Logf("didn't find expected parameter %q", k)
			t.Fail()
		}
	}
}

func TestDeploymentParameters_MarshalRoundTrip(t *testing.T) {
	subject := NewDeploymentParameters()

	subject.Parameters["foo"] = DeploymentParameter{"bar"}
	subject.Parameters["bar"] = DeploymentParameter{2}

	buf := bytes.NewBuffer([]byte{})

	enc := json.NewEncoder(buf)
	if err := enc.Encode(subject); err != nil {
		t.Error(err)
		return
	}

	rehydrated := NewDeploymentParameters()
	dec := json.NewDecoder(buf)
	if err := dec.Decode(rehydrated); err != nil {
		t.Error(err)
		return
	}

	if rehydrated.ContentVersion != subject.ContentVersion {
		t.Logf("Unexpected Content Version:\n\tgot:  %q\n\twant: %q", rehydrated.ContentVersion, subject.ContentVersion)
		t.Fail()
	}

	if rehydrated.Schema != subject.Schema {
		t.Logf("Unexpected Content Version:\n\tgot:  %q\n\twant: %q", rehydrated.Schema, subject.Schema)
		t.Fail()
	}

	if len(rehydrated.Parameters) != len(subject.Parameters) {
		t.Logf("parameter count:\n\tgot: %d want: %d", len(rehydrated.Parameters), len(subject.Parameters))
		t.Fail()
	}

	for k, v := range rehydrated.Parameters {
		if orig, ok := subject.Parameters[k]; ok {
			t.Logf("Found %q -> %v : %v", k, orig, v)
			delete(subject.Parameters, k)
		} else {
			t.Logf("unexpected parameter: %q included with value: %v", k, v)
			t.Fail()
		}
	}

	for k := range subject.Parameters {
		t.Logf("Missing parameter: %q", k)
		t.Fail()
	}
}
