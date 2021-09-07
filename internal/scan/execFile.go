package scan

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
)

var ErrNotAResolver = errors.New("not a resolver")

type ExecFile struct {
	Dir, Name, Ext string

	ObjectName string
	Fields     []*ast.FieldDefinition
}

func (ef *ExecFile) Path() string {
	return filepath.Join(ef.Dir, ef.Name+ef.Ext)
}

func (ef *ExecFile) Scan() error {
	var (
		path         = ef.Path()
		fieldStrings []string
	)
	// populate fieldStrings
	{
		cmd := exec.Command(path, "--cggi-fields")
		schemaBytes, err := cmd.Output()
		if err != nil {
			return fmt.Errorf(
				"error executing %s --cggi-fields: %w",
				path, ErrNotAResolver,
			)
		}

		if err := json.Unmarshal(schemaBytes, &fieldStrings); err != nil {
			return fmt.Errorf(
				"error parsing json output of %s --cggi-fields: %w",
				path, err,
			)
		}
	}

	// parse field strings
	{
		parsedOutput, err := parser.Parse(parser.ParseParams{
			Source: fmt.Sprintf(
				"type Query {\n\t%s\n}",
				strings.Join(fieldStrings, "\n\t"),
			),
		})
		if err != nil {
			return fmt.Errorf(
				"error parsing fields returned by %s: %w",
				path, err,
			)
		}

		if len(parsedOutput.Definitions) != 1 {
			return fmt.Errorf(
				"error parsing fields returned by %s: expected 1 definition",
				path,
			)
		}

		objDef, ok := parsedOutput.Definitions[0].(*ast.ObjectDefinition)
		if !ok {
			return fmt.Errorf(
				"error parsing fields returned by %s: no object definition found",
				path,
			)
		}

		ef.Fields = objDef.Fields
	}

	return nil
}
