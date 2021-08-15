package scan

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/matryer/is"
)

type testExec string

func (te testExec) realize() (string, error) {
	var script = `#!/bin/sh
if [ "$1" = '--cggi-fields' ]; then
	echo '%s'
else
	exit 1	
fi
`

	script = fmt.Sprintf(script, te)

	file, err := ioutil.TempFile("", "*.sh")
	if err != nil {
		return "", nil
	}
	defer file.Close()

	if _, err := file.WriteString(script); err != nil {
		return "", err
	}

	if err := os.Chmod(file.Name(), 0500); err != nil {
		os.Remove(file.Name())
		return "", err
	}

	return file.Name(), nil
}

func TestExecFile_Path(t *testing.T) {
	var (
		is = is.New(t)

		fieldsJSON testExec = `["field1: String!", "field2(arg1: Int): Int"]`
		ef         execFile

		err error
	)

	{
		ef.path, err = fieldsJSON.realize()
		defer os.Remove(ef.path)
		is.NoErr(err)
	}

	is.Equal(ef.Path(), ef.path)
}

func TestExecFile_Fields(t *testing.T) {
	var (
		is = is.New(t)

		fieldsJSON testExec = `["field1: String!", "field2(arg1: Int): Int"]`
		ef         execFile

		err error
	)

	{
		ef.path, err = fieldsJSON.realize()
		defer os.Remove(ef.path)
		is.NoErr(err)
	}

	fields, err := ef.Fields()
	is.NoErr(err)

	is.Equal(len(fields), 2)

	{
		field := fields[0]

		is.Equal(field.Name, "field1")
		is.Equal(field.Type.String(), "NonNull")
		nonNull, ok := field.Type.(*ast.NonNull)
		is.True(ok)
		is.Equal(nonNull.Type.String(), "Named")
		named, ok := nonNull.Type.(*ast.Named)
		is.True(ok)
		is.Equal(named.Name.Value, "String")
	}
	{
		field := fields[1]

		is.Equal(field.Name, "field2")
		is.Equal(field.Type.String(), "Named")
		named, ok := field.Type.(*ast.Named)
		is.True(ok)
		is.Equal(named.Name.Value, "Int")

		is.Equal(len(field.Arguments), 1)
		arg := field.Arguments[0]
		is.Equal(arg.Name.Value, "arg1")
		named, ok = arg.Type.(*ast.Named)
		is.True(ok)
		is.Equal(named.Name.Value, "Int")
	}
}
