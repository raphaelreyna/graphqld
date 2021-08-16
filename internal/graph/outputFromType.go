package graph

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/raphaelreyna/graphqld/internal/intermediary"
)

func (g *Graph) gqlOutputFromType(referer interface{}, referencingDir, referencingTypeName, referencingFieldName string, t ast.Type) graphql.Output {
	var standardBasicScalarFromNamedType = func(named *ast.Named) *graphql.Scalar {
		switch named.Name.Value {
		case "String":
			return graphql.String
		case "Int":
			return graphql.Int
		case "Boolean":
			return graphql.Boolean
		default:
			return nil
		}
	}

	if g.uninstantiatedTypes == nil {
		g.uninstantiatedTypes = make(map[string]interface{})
	}
	if g.typeReferences == nil {
		g.typeReferences = make([]typeReference, 0)
	}

	switch x := t.(type) {
	case *ast.NonNull:
		named, ok := x.Type.(*ast.Named)
		if !ok {
			panic("received nonnamed")
		}

		if scalar := standardBasicScalarFromNamedType(named); scalar != nil {
			return graphql.NewNonNull(scalar)
		}

		to := intermediary.NonNullType{
			TypeName: named.Name.Value,
		}
		if _, exists := g.uninstantiatedTypes[to.TypeName]; !exists {
			g.uninstantiatedTypes[to.TypeName] = to
		}
		if referencingFieldName != "" && referencingTypeName != "" {
			g.typeReferences = append(g.typeReferences, typeReference{
				referencingDir:       referencingDir,
				referencingType:      referencingTypeName,
				referer:              referer,
				referencingFieldName: referencingFieldName,
				referencedType:       to.TypeName,
				typeWrapper:          twNonNull,
			})
		}
		return to
	case *ast.List:
		named, ok := x.Type.(*ast.Named)
		if !ok {
			panic("received nonnamed")
		}

		if scalar := standardBasicScalarFromNamedType(named); scalar != nil {
			return graphql.NewList(scalar)
		}

		to := intermediary.ListType{
			TypeName: named.Name.Value,
		}
		if _, exists := g.uninstantiatedTypes[to.TypeName]; !exists {
			g.uninstantiatedTypes[to.TypeName] = to
		}

		if referencingFieldName != "" && referencingTypeName != "" {
			g.typeReferences = append(g.typeReferences, typeReference{
				referencingDir:       referencingDir,
				referencingType:      referencingTypeName,
				referer:              referer,
				referencingFieldName: referencingFieldName,
				referencedType:       to.TypeName,
				typeWrapper:          twList,
			})
		}
		return to
	case *ast.Named:
		if scalar := standardBasicScalarFromNamedType(x); scalar != nil {
			return scalar
		}

		to := intermediary.Type{
			TypeName: x.Name.Value,
		}
		if _, exists := g.uninstantiatedTypes[to.TypeName]; !exists {
			g.uninstantiatedTypes[to.TypeName] = to
		}

		if referencingFieldName != "" && referencingTypeName != "" {
			g.typeReferences = append(g.typeReferences, typeReference{
				referencingDir:       referencingDir,
				referencingType:      referencingTypeName,
				referer:              referer,
				referencingFieldName: referencingFieldName,
				referencedType:       to.TypeName,
				typeWrapper:          twNone,
			})
		}

		return to
	}
	return nil
}
