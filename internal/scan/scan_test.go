package scan

import (
	"errors"
	"testing"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/matryer/is"
)

type testFile struct {
	path   string
	fields []*FieldOutput
	err    error
}

func (tf testFile) Path() string {
	return tf.path
}

func (tf testFile) Fields() ([]*FieldOutput, error) {
	return tf.fields, tf.err
}

func TestScan(t *testing.T) {
	is := is.New(t)

	tf := testFile{
		path: "path",
		fields: []*FieldOutput{
			{
				Raw:  "raw1",
				Name: "name1",
				Type: &ast.Named{},
			},
			{
				Raw:  "raw2",
				Name: "name2",
				Type: &ast.NonNull{},
				Arguments: []*ast.InputValueDefinition{
					{
						Name: &ast.Name{
							Value: "arg1",
						},
						Type: &ast.Named{},
					},
				},
			},
		},
	}

	ff, err := Scan("parent", tf)
	is.NoErr(err)
	is.True(ff != nil)
	//lint:ignore SA5011 checked by test
	is.Equal(ff.ParentName, "parent")
	//lint:ignore SA5011 checked by test
	is.Equal(ff.Path, "path")
	//lint:ignore SA5011 checked by test
	is.Equal(len(ff.Fields), 2)
	//lint:ignore SA5011 checked by test
	is.Equal(ff.Fields[0], tf.fields[0])
	//lint:ignore SA5011 checked by test
	is.Equal(ff.Fields[1], tf.fields[1])
}

func TestScan_Error(t *testing.T) {
	is := is.New(t)

	tf := testFile{
		path: "path",
		err:  errors.New("error"),
	}

	ff, err := Scan("parent", tf)
	is.True(errors.Is(err, tf.err))
	is.True(ff == nil)
}
