package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
)

type graph struct {
	rootDir string

	tm                  typeObjectMap
	uninstantiatedTypes map[string]interface{}
	typeReferences      map[typeReference]struct{}

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
				Type: g.gqlOutputFromType("Query", fieldOutput.name, fieldOutput.gqlType),
			}

			if args := fieldOutput.arguments; 0 < len(args) {
				arguments := graphql.FieldConfigArgument{}

				for _, arg := range args {
					arguments[arg.Name.Value] = &graphql.ArgumentConfig{
						Type: g.gqlOutputFromType("Query", fieldOutput.name, arg.Type),
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

func (g *graph) gqlOutputFromType(referencingTypeName, referencingFieldName string, t ast.Type) graphql.Output {
	var standardScalarFromNamedType = func(named *ast.Named) *graphql.Scalar {
		switch named.Name.Value {
		case "String":
			return graphql.String
		case "Int":
			return graphql.Int
		default:
			return nil
		}
	}

	if g.uninstantiatedTypes == nil {
		g.uninstantiatedTypes = make(map[string]interface{})
	}
	if g.typeReferences == nil {
		g.typeReferences = make(map[typeReference]struct{})
	}

	switch x := t.(type) {
	case *ast.NonNull:
		named, ok := x.Type.(*ast.Named)
		if !ok {
			panic("received nonnamed")
		}

		if scalar := standardScalarFromNamedType(named); scalar != nil {
			return graphql.NewNonNull(scalar)
		}

		to := _nonNullType{named.Name.Value}
		if _, exists := g.uninstantiatedTypes[to.name]; !exists {
			g.uninstantiatedTypes[to.name] = to
		}
		if referencingFieldName != "" && referencingTypeName != "" {
			g.typeReferences[typeReference{
				referenceringType: referencingTypeName,
				referencingField:  referencingFieldName,
				referencedType:    to.name,
				typeWrapper:       twNonNull,
			}] = struct{}{}
		}
		return to
	case *ast.List:
		named, ok := x.Type.(*ast.Named)
		if !ok {
			panic("received nonnamed")
		}

		if scalar := standardScalarFromNamedType(named); scalar != nil {
			return graphql.NewList(scalar)
		}

		to := _listType{named.Name.Value}
		if _, exists := g.uninstantiatedTypes[to.name]; !exists {
			g.uninstantiatedTypes[to.name] = to
		}

		if referencingFieldName != "" && referencingTypeName != "" {
			g.typeReferences[typeReference{
				referenceringType: referencingTypeName,
				referencingField:  referencingFieldName,
				referencedType:    to.name,
				typeWrapper:       twList,
			}] = struct{}{}
		}
		return to
	case *ast.Named:
		if scalar := standardScalarFromNamedType(x); scalar != nil {
			return scalar
		}

		to := _type{x.Name.Value}
		if _, exists := g.uninstantiatedTypes[to.name]; !exists {
			g.uninstantiatedTypes[to.name] = to
		}
		if referencingFieldName != "" && referencingTypeName != "" {
			g.typeReferences[typeReference{
				referenceringType: referencingTypeName,
				referencingField:  referencingFieldName,
				referencedType:    to.name,
				typeWrapper:       twNone,
			}] = struct{}{}
		}

		return to
	}
	return nil
}

func isUserExec(info fs.FileInfo) bool {
	return info.Mode().Perm()&0100 != 0 && !info.IsDir()
}

func (g *graph) instantiateTypesObjects() error {
	if len(g.uninstantiatedTypes) == 0 {
		return nil
	}

	if len(g.tm) == 0 {
		g.tm = make(typeObjectMap)
	}

	for ut, _ := range g.uninstantiatedTypes {
		dirEntries, err := os.ReadDir(g.rootDir)
		if err != nil {
			return err
		}

		var (
			utDirPath string
		)
		for _, de := range dirEntries {
			if de.Name() == ut {
				utDirPath = filepath.Join(g.rootDir, de.Name())
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

func (g *graph) setTypes() {
	for tr := range g.typeReferences {
		if tr.referenceringType != "Query" {
			panic("unsupported")
		}

		field := g.rootQuery.
			objectConf.
			Fields.(graphql.FieldsThunk)()[tr.referencingField]

		switch tr.typeWrapper {
		case twList:
			field.Type = graphql.NewList(g.tm[tr.referencedType])
		case twNonNull:
			field.Type = graphql.NewNonNull(g.tm[tr.referencedType])
		case twNone:
			field.Type = g.tm[tr.referencedType]
		}
	}
}
