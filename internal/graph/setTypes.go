package graph

import "github.com/graphql-go/graphql"

func (g *Graph) SetTypes() {
	for tr := range g.typeReferences {
		if tr.referenceringType != "Query" {
			panic("unsupported")
		}

		field := g.Query.
			ObjectConf.
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
