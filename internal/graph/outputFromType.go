package graph

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/raphaelreyna/graphqld/internal/intermediary"
)

func (g *Graph) gqlOutputFromType(referencingTypeName, referencingFieldName string, t ast.Type) graphql.Output {
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

		to := intermediary.NonNullType{named.Name.Value}
		if _, exists := g.uninstantiatedTypes[to.TypeName]; !exists {
			g.uninstantiatedTypes[to.TypeName] = to
		}
		if referencingFieldName != "" && referencingTypeName != "" {
			g.typeReferences[typeReference{
				referenceringType: referencingTypeName,
				referencingField:  referencingFieldName,
				referencedType:    to.TypeName,
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

		to := intermediary.ListType{named.Name.Value}
		if _, exists := g.uninstantiatedTypes[to.TypeName]; !exists {
			g.uninstantiatedTypes[to.TypeName] = to
		}

		if referencingFieldName != "" && referencingTypeName != "" {
			g.typeReferences[typeReference{
				referenceringType: referencingTypeName,
				referencingField:  referencingFieldName,
				referencedType:    to.TypeName,
				typeWrapper:       twList,
			}] = struct{}{}
		}
		return to
	case *ast.Named:
		if scalar := standardScalarFromNamedType(x); scalar != nil {
			return scalar
		}

		to := intermediary.Type{x.Name.Value}
		if _, exists := g.uninstantiatedTypes[to.TypeName]; !exists {
			g.uninstantiatedTypes[to.TypeName] = to
		}
		if referencingFieldName != "" && referencingTypeName != "" {
			g.typeReferences[typeReference{
				referenceringType: referencingTypeName,
				referencingField:  referencingFieldName,
				referencedType:    to.TypeName,
				typeWrapper:       twNone,
			}] = struct{}{}
		}

		return to
	}
	return nil
}
