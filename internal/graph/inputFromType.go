package graph

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/raphaelreyna/graphqld/internal/intermediary"
)

func (g *Graph) gqlInputFromType(reference *inputReference, t ast.Type) graphql.Input {
	var standardBasicScalarFromNamedType = func(named *ast.Named) *graphql.Scalar {
		switch named.Name.Value {
		case "String":
			return graphql.String
		case "Int":
			return graphql.Int
		case "Float":
			return graphql.Float
		case "Boolean":
			return graphql.Boolean
		case "ID":
			return graphql.ID
		case "DateTime":
			return graphql.DateTime
		default:
			return nil
		}
	}

	if g.uninstantiatedInputs == nil {
		g.uninstantiatedInputs = make(map[string]interface{})
	}
	if g.inputReferences == nil {
		g.inputReferences = make([]*inputReference, 0)
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

		ii := intermediary.NonNullInput{
			InputName: named.Name.Value,
		}
		if _, exists := g.uninstantiatedInputs[reference.key(ii.InputName)]; !exists {
			g.uninstantiatedInputs[reference.key(ii.InputName)] = ii
		}

		if reference.referencingFieldName != "" && reference.referencingType != "" {
			reference.referencedInput = ii.InputName
			reference.inputWrapper = iwNonNull
			g.inputReferences = append(g.inputReferences, reference)
		}

		return ii
	case *ast.List:
		named, ok := x.Type.(*ast.Named)
		if !ok {
			panic("received nonnamed")
		}

		if scalar := standardBasicScalarFromNamedType(named); scalar != nil {
			return graphql.NewList(scalar)
		}

		ii := intermediary.ListInput{
			InputName: named.Name.Value,
		}
		if _, exists := g.uninstantiatedInputs[reference.key(ii.InputName)]; !exists {
			g.uninstantiatedInputs[reference.key(ii.InputName)] = ii
		}

		if reference.referencingFieldName != "" && reference.referencingType != "" {
			reference.referencedInput = ii.InputName
			reference.inputWrapper = iwList
			g.inputReferences = append(g.inputReferences, reference)
		}

		return ii
	case *ast.Named:
		if scalar := standardBasicScalarFromNamedType(x); scalar != nil {
			return scalar
		}

		ii := intermediary.Input{
			InputName: x.Name.Value,
		}
		if _, exists := g.uninstantiatedInputs[reference.key(ii.InputName)]; !exists {
			g.uninstantiatedInputs[reference.key(ii.InputName)] = ii
		}

		if reference.referencingFieldName != "" && reference.referencingType != "" {
			reference.referencedInput = ii.InputName
			reference.inputWrapper = iwNone
			g.inputReferences = append(g.inputReferences, reference)
		}

		return ii
	}
	return nil
}
