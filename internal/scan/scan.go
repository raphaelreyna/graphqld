package scan

import (
	"io/fs"
	"path/filepath"
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

type File interface {
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
