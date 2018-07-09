package common

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"path"
	"sort"
	"strings"
)

// PackageSpecifier is a string that represents the name that will be used to refer to
// exported functions and variables from a package.
type PackageSpecifier string

// PackagePath is a string that refers to the location of Go package.
type PackagePath string

// ImportBag captures all of the imports in a Go source file, and attempts
// to ease the process of working with them.
type ImportBag struct {
	bySpec     map[PackageSpecifier]PackagePath
	blankIdent map[PackagePath]struct{}
	localIdent map[PackagePath]struct{}
}

// NewImportBag instantiates an empty ImportBag.
func NewImportBag() *ImportBag {
	return &ImportBag{
		bySpec:     make(map[PackageSpecifier]PackagePath),
		blankIdent: make(map[PackagePath]struct{}),
		localIdent: make(map[PackagePath]struct{}),
	}
}

// NewImportBagFromFile reads a Go source file, finds all imports,
// and returns them as an instantiated ImportBag.
func NewImportBagFromFile(filepath string) (*ImportBag, error) {
	f, err := parser.ParseFile(token.NewFileSet(), filepath, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}

	ib := NewImportBag()

	for _, spec := range f.Imports {
		pkgPath := PackagePath(strings.Trim(spec.Path.Value, `"`))
		if spec.Name == nil {
			ib.AddImport(pkgPath)
		} else {
			ib.AddImportWithSpecifier(pkgPath, PackageSpecifier(spec.Name.Name))
		}
	}

	return ib, nil
}

// AddImport includes a package, and returns the name that was selected to
// be the specified for working with this path. It first attempts to use the
// package name as the specifier. Should that cause a conflict, it determines
// a unique name to be used as the specifier.
//
// If the path provided has already been imported, the existing name for it
// is returned, but err is non-nil.
func (ib *ImportBag) AddImport(pkgPath PackagePath) PackageSpecifier {
	spec, err := FindSpecifier(pkgPath)
	if err != nil {
		spec = "unknown"
	}
	specLen := len(spec)
	suffix := uint(1)
	for {
		err = ib.AddImportWithSpecifier(pkgPath, spec)
		if err == nil {
			break
		} else if err != ErrDuplicateImport {
			panic(err)
		}
		spec = PackageSpecifier(fmt.Sprintf("%s%d", spec[:specLen], suffix))
		suffix++
	}

	return spec
}

// ErrDuplicateImport is the error that will be returned when two packages are both requested
// to be imported using the same specifier.
var ErrDuplicateImport = errors.New("specifier already in use in ImportBag")

// ErrMultipleLocalImport is the error that will be returned when the same package has been imported
// to the specifer "." more than once.
var ErrMultipleLocalImport = errors.New("package already imported into the local namespace")

// AddImportWithSpecifier will add an import with a given name. If it would lead
// to conflicting package specifiers, it returns an error.
func (ib *ImportBag) AddImportWithSpecifier(pkgPath PackagePath, specifier PackageSpecifier) error {
	if specifier == "_" {
		ib.blankIdent[pkgPath] = struct{}{}
		return nil
	}

	if specifier == "." {
		if _, ok := ib.localIdent[pkgPath]; ok {
			return ErrMultipleLocalImport
		}
		ib.localIdent[pkgPath] = struct{}{}
		return nil
	}

	if impPath, ok := ib.bySpec[specifier]; ok && pkgPath != impPath {
		return ErrDuplicateImport
	}

	ib.bySpec[specifier] = pkgPath
	return nil
}

// FindSpecifier finds the specifier assocatied with a particular package.
//
// If the package was not imported, the empty string and false are returned.
//
// If multiple specifiers are assigned to the package, one is returned at
// random.
//
// If the same package is imported with a named specifier, and the blank
// identifier, the name is returned.
func (ib ImportBag) FindSpecifier(pkgPath PackagePath) (PackageSpecifier, bool) {
	for k, v := range ib.bySpec {
		if v == pkgPath {
			return k, true
		}
	}

	if _, ok := ib.blankIdent[pkgPath]; ok {
		return "_", true
	}

	if _, ok := ib.localIdent[pkgPath]; ok {
		return ".", true
	}

	return "", false
}

// List returns each import statement as a slice of strings sorted alphabetically by
// their import paths.
func (ib *ImportBag) List() []string {
	specs := ib.ListAsImportSpec()
	retval := make([]string, len(specs))

	builder := bytes.NewBuffer([]byte{})

	for i, s := range specs {
		if s.Name != nil {
			builder.WriteString(s.Name.Name)
			builder.WriteRune(' ')
		}
		builder.WriteString(s.Path.Value)
		retval[i] = builder.String()
		builder.Reset()
	}
	return retval
}

// ListAsImportSpec returns the imports from the ImportBag as a slice of ImportSpecs
// sorted alphabetically by their import paths.
func (ib *ImportBag) ListAsImportSpec() []*ast.ImportSpec {
	retval := make([]*ast.ImportSpec, 0, len(ib.bySpec)+len(ib.localIdent)+len(ib.blankIdent))

	getLit := func(pkgPath PackagePath) *ast.BasicLit {
		return &ast.BasicLit{
			Kind:  token.STRING,
			Value: fmt.Sprintf("%q", string(pkgPath)),
		}
	}

	for k, v := range ib.bySpec {
		var name *ast.Ident

		if path.Base(string(v)) != string(k) {
			name = ast.NewIdent(string(k))
		}

		retval = append(retval, &ast.ImportSpec{
			Name: name,
			Path: getLit(v),
		})
	}

	for s := range ib.localIdent {
		retval = append(retval, &ast.ImportSpec{
			Name: ast.NewIdent("."),
			Path: getLit(s),
		})
	}

	for s := range ib.blankIdent {
		retval = append(retval, &ast.ImportSpec{
			Name: ast.NewIdent("_"),
			Path: getLit(s),
		})
	}

	sort.Slice(retval, func(i, j int) bool {
		return strings.Compare(retval[i].Path.Value, retval[j].Path.Value) < 0
	})

	return retval
}

var impFinder = importer.Default()

// FindSpecifier finds the name of a package by loading it in from GOPATH
// or a vendor folder.
func FindSpecifier(pkgPath PackagePath) (PackageSpecifier, error) {
	var pkg *types.Package
	var err error

	if cast, ok := impFinder.(types.ImporterFrom); ok {
		pkg, err = cast.ImportFrom(string(pkgPath), "", 0)
	} else {
		pkg, err = impFinder.Import(string(pkgPath))
	}

	if err != nil {
		return "", err
	}
	return PackageSpecifier(pkg.Name()), nil
}
