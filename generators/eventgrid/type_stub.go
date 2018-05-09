package eventgrid

import (
	"errors"
	"reflect"
	"strings"
)

// TypeStub fulfills the reflect.Type interface, but only knows the fully qualified
// name of a type. All other details will panic upon use.
type TypeStub struct {
	reflect.Type
	pkgName  string
	typeName string
}

// NewTypeStub creates a new reflect.Type stub based on a package and a type.
func NewTypeStub(packagePath string, typeName string) (*TypeStub, error) {
	return &TypeStub{
		pkgName:  packagePath,
		typeName: typeName,
	}, nil
}

// NewTypeStubIdentifier creates a new reflect.Type stub based on a fully-qualified
// Go type name. The expected format of the identifier is:
// <package path>.<type name>
//
// For example:
// ```
// github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid.StorageBlobCreatedEventData
// ```
func NewTypeStubIdentifier(identifier string) (*TypeStub, error) {
	last := strings.LastIndex(identifier, ".")
	if last < 0 {
		return nil, errors.New("no type found")
	}

	return &TypeStub{
		pkgName:  identifier[:last],
		typeName: identifier[last+1:],
	}, nil
}

// Name fetches the type's name within the package.
func (t *TypeStub) Name() string {
	return t.typeName
}

// PkgPath fetches the package's unique identifier.
func (t *TypeStub) PkgPath() string {
	return t.pkgName
}
