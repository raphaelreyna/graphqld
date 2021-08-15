package scan

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
)

type execFile struct {
	path string
}

func (ef execFile) Path() string {
	return ef.path
}

func (ef execFile) Fields() ([]*FieldOutput, error) {
	var fieldStrings []string
	// populate fieldStrings
	{
		cmd := exec.Command(ef.path, "--cggi-fields")
		schemaBytes, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf(
				"error executing %s --cggi-fields: %w",
				ef.path, err,
			)
		}

		if err := json.Unmarshal(schemaBytes, &fieldStrings); err != nil {
			return nil, fmt.Errorf(
				"error parsing json output of %s --cggi-fields: %w",
				ef.path, err,
			)
		}
	}

	fields := make([]*FieldOutput, len(fieldStrings))

	// parse field strings
	var (
		astFieldDefs []*ast.FieldDefinition
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
				ef.path, err,
			)
		}

		if len(parsedOutput.Definitions) != 1 {
			return nil, fmt.Errorf(
				"error parsing fields returned by %s: expected 1 definition",
				ef.path,
			)
		}

		objDef, ok := parsedOutput.Definitions[0].(*ast.ObjectDefinition)
		if !ok {
			return nil, fmt.Errorf(
				"error parsing fields returned by %s: no object definition found",
				ef.path,
			)
		}

		astFieldDefs = objDef.Fields
	}

	for idx, field := range astFieldDefs {
		fields[idx] = &FieldOutput{
			Name:      field.Name.Value,
			Type:      field.Type,
			Arguments: field.Arguments,
			Raw:       fieldStrings[idx],
		}
	}

	return fields, nil
}
