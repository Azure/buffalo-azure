package eventgrid

import (
	"reflect"
	"testing"
)

func TestTypeStub_IsType(t *testing.T) {
	subject := &TypeStub{}
	m := reflect.TypeOf((*reflect.Type)(nil)).Elem()

	if !reflect.TypeOf(subject).Implements(m) {
		t.Log("TypeStub does not implement reflect.Type")
		t.Fail()
	}
}
