package scan

import (
	"io/fs"
	"path/filepath"

	"github.com/graphql-go/graphql/language/ast"
)

func Scan(parent string, f File) (*FileFields, error) {
	fields, err := f.Fields()
	if err != nil {
		return nil, err
	}

	return &FileFields{
		Path:       f.Path(),
		ParentName: parent,
		Fields:     fields,
	}, nil
}

func ScanForType(dir, parent, typeName string) (*ObjectDefinition, error) {
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

type FileFields struct {
	Path       string
	ParentName string
	Fields     []*FieldOutput
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

	return graphqlFile{
		path: filepath.Join(root, info.Name()),
	}, nil
}

func isUserExec(info fs.FileInfo) bool {
	return info.Mode().Perm()&0100 != 0 && !info.IsDir()
}
