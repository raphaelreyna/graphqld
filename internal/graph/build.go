package graph

import (
	"github.com/graphql-go/graphql"
	"github.com/raphaelreyna/graphqld/internal/scan"
)

func (g *Graph) Build() error {
	// build object definition for root query object
	{
		def, err := g.buildObjectDefinitionForTypeObject(g.Dir, "query")
		if err != nil {
			return err
		}

		g.Query = *def
	}

	// keep building referenced types as long as we have any
	{
		if g.objDefs == nil {
			g.objDefs = make(map[string]*scan.ObjectDefinition)
		}

		var (
			count                   = len(g.typeReferences)
			processedTypeReferences = make([]typeReference, 0)
		)
		for 0 < count {
			var tr = g.typeReferences[0]

			def, err := g.buildObjectDefinitionForTypeObject(tr.referencingDir, tr.referencedType)
			if err != nil {
				return err
			}

			g.objDefs[def.ObjectConf.Name] = def

			// put this type reference into the pile of processed ones
			// and remove it from the ones we still need to work on
			processedTypeReferences = append(processedTypeReferences, tr)
			g.typeReferences = g.typeReferences[1:]
			count = len(g.typeReferences)
		}

		// lets get our type references back
		g.typeReferences = processedTypeReferences
	}

	// now that we have all of the type object definitions, we need to instantiate them
	{
		if len(g.tm) == 0 {
			g.tm = make(typeObjectMap)
		}

		for name := range g.uninstantiatedTypes {
			def, ok := g.objDefs[name]
			if !ok {
				panic("could not find object definition")
			}

			g.tm[name] = graphql.NewObject(def.ObjectConf)

			delete(g.uninstantiatedTypes, name)
		}
	}

	// now that we instantiated all of our type objects, we need to make sure that
	// pointers pointing to intermediary type objects are set to point to the "real" type object
	{
		for _, tr := range g.typeReferences {
			switch referer := tr.referer.(type) {
			case *graphql.Field:
				switch tr.typeWrapper {
				case twList:
					referer.Type = graphql.NewList(g.tm[tr.referencedType])
				case twNonNull:
					referer.Type = graphql.NewNonNull(g.tm[tr.referencedType])
				case twNone:
					referer.Type = g.tm[tr.referencedType]
				}
			case *graphql.ArgumentConfig:
				switch tr.typeWrapper {
				case twList:
					referer.Type = graphql.NewList(g.tm[tr.referencedType])
				case twNonNull:
					referer.Type = graphql.NewNonNull(g.tm[tr.referencedType])
				case twNone:
					referer.Type = g.tm[tr.referencedType]
				}
			}
		}
	}

	// finally we create a resolver for each field that needs one
	{
		if err := g.Query.SetResolvers(); err != nil {
			return err
		}
		for _, objDef := range g.objDefs {
			if err := objDef.SetResolvers(); err != nil {
				return err
			}
		}
	}

	return nil
}
