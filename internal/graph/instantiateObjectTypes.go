package graph

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
)

func (g *Graph) InstantiateTypesObjects() error {
	if len(g.uninstantiatedTypes) == 0 {
		return nil
	}

	if len(g.tm) == 0 {
		g.tm = make(typeObjectMap)
	}

	for ut := range g.uninstantiatedTypes {
		dirEntries, err := os.ReadDir(g.Dir)
		if err != nil {
			return err
		}

		var (
			utDirPath string
		)
		for _, de := range dirEntries {
			if de.Name() == ut {
				utDirPath = filepath.Join(g.Dir, de.Name())
				break
			}
		}

		// check for schema file
		file, err := os.OpenFile(
			filepath.Join(utDirPath, ut+".graphql"),
			os.O_RDONLY, os.ModePerm,
		)
		if err != nil {
			return err
		}
		defer file.Close()
		data, err := io.ReadAll(file)
		if err != nil {
			return fmt.Errorf(
				"error reading file %s: %w",
				file.Name(), err,
			)
		}

		// parse
		parsedSchema, err := parser.Parse(parser.ParseParams{
			Source: string(data),
		})
		if err != nil {
			return err
		}

		// make sure we're dealing with a single object / type definition
		if len(parsedSchema.Definitions) != 1 {
			fmt.Printf("%+v\n", parsedSchema)
			return fmt.Errorf(
				"error parsing %s: expected 1 definition",
				file.Name(),
			)
		}
		objDef, ok := parsedSchema.Definitions[0].(*ast.ObjectDefinition)
		if !ok {
			return fmt.Errorf(
				"error parsing fields returned by %s: no object definition found",
				file.Name(),
			)
		}

		// make sure the names match
		if name := objDef.Name.Value; name != ut {
			return fmt.Errorf(
				"error parsing %s: mismatched type name, expected: %s found: %s",
				file.Name(), ut, name,
			)
		}

		fields := graphql.Fields{}
		for _, field := range objDef.Fields {
			name := field.Name.Value
			fields[name] = &graphql.Field{
				Name: name,
				Type: g.gqlOutputFromType("", name, field.Type),
			}
		}

		g.tm[ut] = graphql.NewObject(graphql.ObjectConfig{
			Name: ut,
			Fields: graphql.FieldsThunk(func() graphql.Fields {
				return fields
			}),
		})

		delete(g.uninstantiatedTypes, ut)
	}

	return nil
}
