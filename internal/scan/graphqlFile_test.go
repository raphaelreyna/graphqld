package scan

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/matryer/is"
)

func TestGraphQLFile_Path(t *testing.T) {
	var (
		is = is.New(t)

		gf graphqlFile
	)

	gf.path = "path"

	is.Equal(gf.Path(), gf.path)
}

func TestGraphQL_Fields(t *testing.T) {
	var (
		is = is.New(t)

		gf graphqlFile

		gql = `
type TestType {
	field1: String!
	field2(arg1: Int): Int
}
		`
	)

	file, err := ioutil.TempFile("", "*.graphql")
	defer func() {
		if file != nil {
			file.Close()
			os.Remove(file.Name())
		}
	}()
	is.NoErr(err)

	_, err = file.WriteString(gql)
	is.NoErr(err)

	gf.path = file.Name()

	fields, err := gf.Fields()
	is.NoErr(err)
	is.True(fields != nil)
	is.Equal(len(fields), 2)
	{
		//lint:ignore SA5011 checked by test
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
		//lint:ignore SA5011 checked by test
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
