package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/raphaelreyna/graphqld/internal/scan"
)

func (g *Graph) SynthesizeRootQueryConf() error {
	var (
		fields    = []string{}
		gqlFields = graphql.Fields{}

		rootDir = g.Dir

		definition = scan.ObjectDefinition{
			ResolverPaths: map[string]string{},
			ObjectConf: graphql.ObjectConfig{
				Name: "Query",
				Fields: graphql.FieldsThunk(func() graphql.Fields {
					return gqlFields
				}),
			},
		}
	)

	dirEntries, err := os.ReadDir(rootDir)
	if err != nil {
		return fmt.Errorf(
			"error reading dir entries in %s: %w",
			rootDir, err,
		)
	}

	for _, dirEntry := range dirEntries {
		info, err := dirEntry.Info()
		if err != nil {
			return fmt.Errorf(
				"error reading info for %s: %w",
				dirEntry.Name(), err,
			)
		}

		if info.IsDir() {
			continue
		}

		var execPath = filepath.Join(g.Dir, dirEntry.Name())

		file, err := scan.NewFile(rootDir, info)
		if err != nil {
			return fmt.Errorf(
				"error %s: %w",
				execPath, err,
			)
		}

		fieldsOutput, err := scan.Scan("query", file)
		if err != nil {
			return fmt.Errorf(
				"error reading fields from %s: %w",
				execPath, err,
			)
		}

		for _, fieldOutput := range fieldsOutput.Fields {
			gqlField := graphql.Field{
				Name: fieldOutput.Name,
				Type: g.gqlOutputFromType("Query", fieldOutput.Name, fieldOutput.Type),
			}

			if args := fieldOutput.Arguments; 0 < len(args) {
				arguments := graphql.FieldConfigArgument{}

				for _, arg := range args {
					arguments[arg.Name.Value] = &graphql.ArgumentConfig{
						Type: g.gqlOutputFromType("Query", fieldOutput.Name, arg.Type),
					}
				}

				gqlField.Args = arguments
			}

			gqlFields[fieldOutput.Name] = &gqlField

			definition.ResolverPaths[fieldOutput.Name] = execPath
		}
	}

	definition.DefinitionString = fmt.Sprintf("type Query {\n\t%s\n}", strings.Join(fields, "\n\t"))

	g.Query = definition

	return nil
}
