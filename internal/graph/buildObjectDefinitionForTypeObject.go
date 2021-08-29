package graph

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/raphaelreyna/graphqld/internal/intermediary"
	"github.com/raphaelreyna/graphqld/internal/objdef"
	"github.com/raphaelreyna/graphqld/internal/scan"
)

var (
	ErrTypeHasNoDir = errors.New("directory not found for type")
)

// buildObjectDefinitionForTypeObject looks for a directory named name in dir and scans it
// for files that graphqld can use to build up the graph for the type named name.
func (g *Graph) buildObjectDefinitionForTypeObject(dir, name string) (*objdef.ObjectDefinition, error) {
	var (
		fields    = []string{}
		gqlFields = graphql.Fields{}

		scriptsDir = filepath.Join(dir, name)

		definition = objdef.ObjectDefinition{
			ResolverPaths: make(map[string]string),
			ObjectConf: graphql.ObjectConfig{
				Name: name,
				Fields: graphql.FieldsThunk(func() graphql.Fields {
					return gqlFields
				}),
			},
		}
	)

	if (name == "query" || name == "Query") && g.Dir == dir {
		name = "query"
		scriptsDir = dir
	}

	dirEntries, err := os.ReadDir(scriptsDir)
	if err != nil {
		return nil, fmt.Errorf(
			"error building type object %s referenced from %s: %w",
			name, dir, ErrTypeHasNoDir,
		)
	}
	for _, dirEntry := range dirEntries {
		info, err := dirEntry.Info()
		if err != nil {
			return nil, fmt.Errorf(
				"error reading info for %s: %w",
				dirEntry.Name(), err,
			)
		}

		if info.IsDir() {
			continue
		}

		var execPath = filepath.Join(scriptsDir, dirEntry.Name())

		file, err := scan.NewFile(scriptsDir, info)
		if err != nil {
			return nil, fmt.Errorf(
				"error %s: %w",
				execPath, err,
			)
		}
		if file == nil {
			continue
		}

		fieldsOutput, err := scan.Scan(name, file)
		if err != nil {
			if errors.Is(err, scan.ErrNoFields) {
				continue
			}

			return nil, fmt.Errorf(
				"error reading fields from %s: %w",
				execPath, err,
			)
		}

		for _, fieldOutput := range fieldsOutput.Fields {
			gqlField := graphql.Field{
				Name: fieldOutput.Name,
			}
			gqlField.Type = g.gqlOutputFromType(
				&gqlField, scriptsDir, name,
				fieldOutput.Name, fieldOutput.Type,
			)

			if args := fieldOutput.Arguments; 0 < len(args) {
				var arguments = make(graphql.FieldConfigArgument)

				for _, arg := range args {
					var (
						argConf   graphql.ArgumentConfig
						reference = inputReference{
							referencingDir:       dir,
							referencingType:      name,
							referencingFieldName: fieldOutput.Name,
							referencingArgName:   arg.Name.Value,
							referer:              &argConf,
						}
					)

					if arg.Description != nil {
						argConf.Description = arg.Description.Value
					}
					if arg.DefaultValue != nil {
						argConf.DefaultValue = arg.DefaultValue.GetValue()
					}
					argConf.Type = g.gqlInputFromType(&reference, arg.Type)

					if intermediary.IsIntermediary(argConf.Type) {
						inputName := argConf.Type.Name()
						g.inputConfs[reference.key(inputName)] = &graphql.InputObjectConfig{
							Name: inputName,
						}
					}

					arguments[arg.Name.Value] = &argConf
				}

				gqlField.Args = arguments
			}

			gqlFields[fieldOutput.Name] = &gqlField

			if file.IsExec() {
				definition.ResolverPaths[fieldOutput.Name] = execPath
			}
		}
	}

	definition.DefinitionString = fmt.Sprintf(
		"type %s {\n\t%s\n}",
		name, strings.Join(fields, "\n\t"),
	)

	return &definition, nil
}
