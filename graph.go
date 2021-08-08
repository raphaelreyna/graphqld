package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

type graph struct {
	rootDir string

	rootQuery objectDefinition
}

func (g *graph) synthesizeRootQueryConf() error {
	var (
		fields    = []string{}
		gqlFields = graphql.Fields{}

		rootDir = g.rootDir

		definition = objectDefinition{
			resolverPaths: map[string]string{},
			objectConf: graphql.ObjectConfig{
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

		if !isUserExec(info) {
			continue
		}

		var (
			execPath = filepath.Join(g.rootDir, dirEntry.Name())
		)
		fieldsOutput, err := newFieldsOutput(execPath, "Query")
		if err != nil {
			return fmt.Errorf(
				"error reading fields from %s: %w",
				execPath, err,
			)
		}

		for _, fieldOutput := range fieldsOutput.fields {
			gqlField := graphql.Field{
				Name: fieldOutput.name,
				Type: g.gqlOutputFromType(fieldOutput.gqlType),
			}

			if args := fieldOutput.arguments; 0 < len(args) {
				arguments := graphql.FieldConfigArgument{}

				for _, arg := range args {
					arguments[arg.Name.Value] = &graphql.ArgumentConfig{
						Type: g.gqlOutputFromType(arg.Type),
					}
				}

				gqlField.Args = arguments
			}

			gqlFields[fieldOutput.name] = &gqlField

			definition.resolverPaths[fieldOutput.name] = execPath
		}

	}

	definition.definitionString = fmt.Sprintf("type Query {\n\t%s\n}", strings.Join(fields, "\n\t"))

	g.rootQuery = definition

	return nil
}

func (g *graph) gqlOutputFromType(t ast.Type) graphql.Output {
	var scalarFromNamedType = func(named *ast.Named) *graphql.Scalar {
		switch named.Name.Value {
		case "String":
			return graphql.String
		case "Int":
			return graphql.Int
		default:
			return nil
		}
	}

	switch x := t.(type) {
	case *ast.NonNull:
		named, ok := x.Type.(*ast.Named)
		if !ok {
			panic("received nonnamed")
		}

		if scalar := scalarFromNamedType(named); scalar != nil {
			return graphql.NewNonNull(scalar)
		}

		return _nonNullType{named.Name.Value}
	case *ast.List:
		named, ok := x.Type.(*ast.Named)
		if !ok {
			panic("received nonnamed")
		}

		if scalar := scalarFromNamedType(named); scalar != nil {
			return graphql.NewList(scalar)
		}

		return _listType{named.Name.Value}
	case *ast.Named:
		if scalar := scalarFromNamedType(x); scalar != nil {
			return scalar
		}

		return _type{x.Name.Value}
	}

	return nil
}

func isUserExec(info fs.FileInfo) bool {
	return info.Mode().Perm()&0100 != 0 && !info.IsDir()
}
