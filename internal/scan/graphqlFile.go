package scan

import (
	"fmt"
	"io"
	"os"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
)

type graphqlFile struct {
	path string
}

func (gf graphqlFile) Path() string {
	return gf.path
}

func (gf graphqlFile) Fields() ([]*FieldOutput, error) {
	file, err := os.OpenFile(gf.path, os.O_RDONLY, os.ModePerm)
	defer func() {
		if file != nil {
			file.Close()
		}
	}()
	if err != nil {
		return nil, nil
	}

	// parse field strings
	var (
		fileName = file.Name()

		fields = make([]*FieldOutput, 0)

		astFieldDefs []*ast.FieldDefinition
	)
	{
		data, err := io.ReadAll(file)
		if err != nil {
			return nil, fmt.Errorf(
				"error reading file %s: %w",
				fileName, err,
			)
		}
		parsedOutput, err := parser.Parse(parser.ParseParams{
			Source: string(data),
		})
		if err != nil {
			return nil, fmt.Errorf(
				"error parsing fields returned by %s: %w",
				fileName, err,
			)
		}

		if len(parsedOutput.Definitions) != 1 {
			return nil, fmt.Errorf(
				"error parsing fields returned by %s: expected 1 definition",
				fileName,
			)
		}

		objDef, ok := parsedOutput.Definitions[0].(*ast.ObjectDefinition)
		if !ok {
			return nil, fmt.Errorf(
				"error parsing fields returned by %s: no object definition found",
				fileName,
			)
		}

		astFieldDefs = objDef.Fields
	}

	for _, field := range astFieldDefs {
		output := FieldOutput{
			Name:      field.Name.Value,
			Type:      field.Type,
			Arguments: field.Arguments,
		}

		fields = append(fields, &output)
	}

	return fields, nil
}
