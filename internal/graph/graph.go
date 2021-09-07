package graph

import (
	"errors"

	"github.com/graphql-go/graphql"
	"github.com/raphaelreyna/graphqld/internal/graph/resolver"
)

var ErrorNoRoots = errors.New("no root query or mutation directories found")

type definitions map[string]interface{}
type resolverPaths map[string]map[string]string
type enums map[string]*graphql.Enum
type inputs map[string]*graphql.InputObject
type objects map[string]*graphql.Object

type Graph struct {
	DocumentRoot string
	ResolverDir  string

	Query    *graphql.Object
	Mutation *graphql.Object
}

func (g *Graph) Build() error {
	definitions, resolverPaths, err := g.scanForDefinitions()
	if err != nil {
		return err
	}

	enums, err := g.instantiateEnums(definitions)
	if err != nil {
		return err
	}

	inputs, err := g.instantiateInputs(definitions, enums)
	if err != nil {
		return err
	}

	objects, err := g.instantiateObjects(definitions, enums, inputs)
	if err != nil {
		return err
	}

	if q := objects["Query"]; 0 < len(q.Fields()) {
		g.Query = q
	}
	if m := objects["Mutation"]; 0 < len(m.Fields()) {
		g.Mutation = m
	}

	if g.Query == nil && g.Mutation == nil {
		return ErrorNoRoots
	}

	for objName, resolverPaths := range resolverPaths {
		var fields = objects[objName].Fields()

		for fieldName, resolverPath := range resolverPaths {
			var field = fields[fieldName]
			resolver, err := resolver.NewFieldResolveFn(resolverPath, g.ResolverDir, field)
			if err != nil {
				return err
			}

			field.Resolve = *resolver
		}
	}

	return nil
}
