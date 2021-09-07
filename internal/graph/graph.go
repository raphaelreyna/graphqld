package graph

import (
	"errors"

	"github.com/graphql-go/graphql"
	"github.com/raphaelreyna/graphqld/internal/graph/resolver"
)

var ErrorNoRoots = errors.New("no root query or mutation directories found")

type Graph struct {
	DocumentRoot string
	ResolverDir  string

	definitions map[string]interface{}

	resolverPaths map[string]map[string]string

	enums   map[string]*graphql.Enum
	inputs  map[string]*graphql.InputObject
	objects map[string]*graphql.Object

	Query    *graphql.Object
	Mutation *graphql.Object
}

func (g *Graph) Build() error {
	if err := g.scanForDefinitions(); err != nil {
		return err
	}

	if err := g.instantiateEnums(); err != nil {
		return err
	}

	if err := g.instantiateInputs(); err != nil {
		return err
	}

	if err := g.instantiateObjects(); err != nil {
		return err
	}

	if q := g.objects["Query"]; 0 < len(q.Fields()) {
		g.Query = q
	}
	if m := g.objects["Mutation"]; 0 < len(m.Fields()) {
		g.Mutation = m
	}

	if g.Query == nil && g.Mutation == nil {
		return ErrorNoRoots
	}

	for objName, resolverPaths := range g.resolverPaths {
		var fields = g.objects[objName].Fields()

		for fieldName, resolverPath := range resolverPaths {
			var field = fields[fieldName]
			resolver, err := resolver.NewFieldResolveFn(resolverPath, g.ResolverDir, field)
			if err != nil {
				return err
			}

			field.Resolve = *resolver
		}
	}

	g.definitions = nil
	g.resolverPaths = nil
	g.objects = nil
	g.resolverPaths = nil
	g.inputs = nil
	g.enums = nil

	return nil
}
