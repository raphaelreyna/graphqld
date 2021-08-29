package scan

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
)

type GraphQLFile struct {
	path    string
	objects []*ast.ObjectDefinition
	inputs  []*ast.InputObjectDefinition
}

func (gf *GraphQLFile) Path() string {
	return gf.path
}

var (
	ErrNoFields = errors.New("file does not contain any object definition with fields")
	ErrNoInputs = errors.New("file does not contain any object definition with fields")
)

func (gf *GraphQLFile) Fields() ([]*FieldOutput, error) {
	if gf.objects == nil {
		if err := gf.scan(); err != nil {
			return nil, err
		}
	}

	if len(gf.objects) == 0 {
		return nil, nil
	}

	fields := []*FieldOutput{}
	for _, field := range gf.objects[0].Fields {
		output := FieldOutput{
			Name:      field.Name.Value,
			Type:      field.Type,
			Arguments: field.Arguments,
		}

		fields = append(fields, &output)
	}

	return fields, nil
}

func (gf *GraphQLFile) Input() (*ast.InputObjectDefinition, error) {
	if gf.inputs == nil {
		if err := gf.scan(); err != nil {
			return nil, err
		}
	}

	if len(gf.inputs) == 0 {
		return nil, nil
	}

	return gf.inputs[0], nil
}

func (*GraphQLFile) IsExec() bool {
	return false
}

func (gf *GraphQLFile) scan() error {
	file, err := os.OpenFile(gf.path, os.O_RDONLY, os.ModePerm)
	defer func() {
		if file != nil {
			file.Close()
		}
	}()
	if err != nil {
		return nil
	}

	filename := file.Name()
	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf(
			"error reading file %s: %w",
			filename, err,
		)
	}

	parsedOutput, err := parser.Parse(parser.ParseParams{
		Source: string(data),
	})
	if err != nil {
		return fmt.Errorf(
			"error parsing fields returned by %s: %w",
			filename, err,
		)
	}

	if gf.objects == nil {
		gf.objects = []*ast.ObjectDefinition{}
	}
	if gf.inputs == nil {
		gf.inputs = []*ast.InputObjectDefinition{}
	}

	for _, def := range parsedOutput.Definitions {
		switch x := def.(type) {
		case *ast.ObjectDefinition:
			gf.objects = append(gf.objects, x)
		case *ast.InputObjectDefinition:
			gf.inputs = append(gf.inputs, x)
		default:
			return fmt.Errorf("unsupported definition type: %T %v", def, def)
		}
	}

	return nil
}
