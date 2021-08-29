package scan

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/raphaelreyna/graphqld/internal/objdef"
)

func Scan(parent string, f File) (*FileContents, error) {
	switch x := f.(type) {
	case execFile:
		fields, err := x.Fields()
		if err != nil {
			return nil, err
		}

		return &FileContents{
			Path:       f.Path(),
			ParentName: parent,
			Fields:     fields,
		}, nil
	case *GraphQLFile:
		if err := x.scan(); err != nil {
			return nil, err
		}

		fields, _ := x.Fields()
		inputs, _ := x.Input()

		return &FileContents{
			Path:       x.Path(),
			ParentName: parent,
			Fields:     fields,
			Input:      inputs,
		}, nil
	}

	return nil, fmt.Errorf("invalid file type: %T", f)
}

func ScanForType(dir, parent, typeName string) (*objdef.ObjectDefinition, error) {
	defer func() {
		panic("unimplemented")
	}()
	return nil, nil
}

// FieldOutput is an abstraction of the partial definitions
// that graphqld obtains from scanning its root directory.
//
// For executables, this is obtained by running the executable with the `--cggi-fields`.
type FieldOutput struct {
	Raw       string
	Name      string
	Type      ast.Type
	Arguments []*ast.InputValueDefinition
}

type FileContents struct {
	Path       string
	ParentName string
	Fields     []*FieldOutput
	Input      *ast.InputObjectDefinition
}

type File interface {
	IsExec() bool
	Path() string
	Fields() ([]*FieldOutput, error)
}

func NewFile(root string, info fs.FileInfo) (File, error) {
	if isUserExec(info) {
		return execFile{
			path: filepath.Join(root, info.Name()),
		}, nil
	}

	ext := filepath.Ext(info.Name())
	if ext == ".gql" || ext == ".graphql" {
		return &GraphQLFile{
			path: filepath.Join(root, info.Name()),
		}, nil
	}

	return nil, nil
}

func isUserExec(info fs.FileInfo) bool {
	return info.Mode().Perm()&0100 != 0 && !info.IsDir()
}
