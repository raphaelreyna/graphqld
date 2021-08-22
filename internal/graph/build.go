package graph

import (
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/raphaelreyna/graphqld/internal/objdef"
)

func (g *Graph) Build() error {
	// build object definition for root query object
	{
		if g.inputConfs == nil {
			g.inputConfs = make(map[string]*graphql.InputObjectConfig)
		}

		def, err := g.buildObjectDefinitionForTypeObject(g.Dir, "query")
		if err != nil {
			return err
		}

		g.Query = *def
	}

	// keep building referenced types as long as we have any
	{
		if g.objDefs == nil {
			g.objDefs = make(map[string]*objdef.ObjectDefinition)
		}

		var (
			count                   = len(g.typeReferences)
			processedTypeReferences = make([]*typeReference, 0)
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

	// now that we have all of the type object definitions, we need to instantiate the input objects
	{
		if g.im == nil {
			g.im = make(map[string]graphql.Input)
		}

		for name := range g.uninstantiatedInputs {
			conf, ok := g.inputConfs[name]
			if !ok {
				panic("could not find input config")
			}

			g.im[name] = graphql.NewInputObject(*conf)

			delete(g.uninstantiatedInputs, name)
		}
	}

	// now that we instantiated all of our input objects, we need to make sure that
	// pointers pointing to intermediary input objects are set to point to the "real" type object
	{
		for _, ir := range g.inputReferences {
			key := ir.key(ir.referencedInput)
			switch referer := ir.referer.(type) {
			case *graphql.Field:
				switch ir.inputWrapper {
				case twList:
					referer.Type = graphql.NewList(g.im[key])
				case twNonNull:
					referer.Type = graphql.NewNonNull(g.im[key])
				case twNone:
					referer.Type = g.im[key]
				}
			case *graphql.ArgumentConfig:
				switch ir.inputWrapper {
				case twList:
					referer.Type = graphql.NewList(g.im[key])
				case twNonNull:
					referer.Type = graphql.NewNonNull(g.im[key])
				case twNone:
					referer.Type = g.im[key]
				}
			}
		}
	}

	// now that we have all of the type object definitions, we need to instantiate them
	{
		if len(g.tm) == 0 {
			g.tm = make(map[string]*graphql.Object)
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

	fmt.Printf("instantiated objects")

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
