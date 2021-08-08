package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
)

type fieldOutput struct {
	raw       string
	name      string
	gqlType   ast.Type
	arguments []*ast.InputValueDefinition
}

func newFieldOutput(fd *ast.FieldDefinition) fieldOutput {
	return fieldOutput{
		name:      fd.Name.Value,
		gqlType:   fd.Type,
		arguments: fd.Arguments,
	}
}

type fieldsOutput struct {
	path       string
	parentName string
	fields     []*fieldOutput
}

func newFieldsOutput(path, parent string) (*fieldsOutput, error) {
	var (
		fieldStrings []string

		fo = fieldsOutput{
			path:       path,
			parentName: parent,
		}
	)

	// populate fieldStrings
	{
		cmd := exec.Command(path, "--cggi-fields")
		schemaBytes, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf(
				"error executing %s --cggi-fields: %w",
				path, err,
			)
		}

		if err := json.Unmarshal(schemaBytes, &fieldStrings); err != nil {
			return nil, fmt.Errorf(
				"error parsing json output of %s --cggi-fields: %w",
				path, err,
			)
		}
	}

	fo.fields = make([]*fieldOutput, len(fieldStrings))

	// parse field strings
	var (
		fields []*ast.FieldDefinition
	)
	{
		parsedOutput, err := parser.Parse(parser.ParseParams{
			Source: fmt.Sprintf(
				"type Query {\n\t%s\n}",
				strings.Join(fieldStrings, "\n\t"),
			),
		})
		if err != nil {
			return nil, fmt.Errorf(
				"error parsing fields returned by %s: %w",
				path, err,
			)
		}

		if len(parsedOutput.Definitions) != 1 {
			return nil, fmt.Errorf(
				"error parsing fields returned by %s: expected 1 definition",
				path,
			)
		}

		objDef, ok := parsedOutput.Definitions[0].(*ast.ObjectDefinition)
		if !ok {
			return nil, fmt.Errorf(
				"error parsing fields returned by %s: no object definition found",
				path,
			)
		}

		fields = objDef.Fields
	}

	for idx, field := range fields {
		output := newFieldOutput(field)
		output.raw = fieldStrings[idx]

		fo.fields[idx] = &output
	}

	return &fo, nil
}
