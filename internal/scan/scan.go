package scan

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
)

var (
	ErrNoFields = errors.New("file does not contain any object definition with fields")
	ErrNoInputs = errors.New("file does not contain any object definition with fields")
)

type File interface {
	Path() string
	Scan() error
}

func NewFile(path string, info fs.FileInfo) File {
	var (
		dir  = filepath.Dir(path)
		name = filepath.Base(path)
		ext  = filepath.Ext(name)

		isUserExec = func(info fs.FileInfo) bool {
			var (
				perms = info.Mode().Perm()

				isUserExec  = perms&0100 != 0
				isGroupExec = perms&0010 != 0
				isOtherExec = perms&0001 != 0
			)

			return (isUserExec || isGroupExec || isOtherExec) && !info.IsDir()
		}
	)
	name = strings.TrimSuffix(name, ext)

	if isUserExec(info) {
		return &ExecFile{
			Dir:  dir,
			Name: name,
			Ext:  ext,

			// The object name is the name of the directory this exec file is in
			ObjectName: filepath.Base(dir),
		}
	}

	if ext == ".graphql" {
		return &GraphqlFile{
			Dir:  dir,
			Name: name,
		}
	}

	return nil
}
