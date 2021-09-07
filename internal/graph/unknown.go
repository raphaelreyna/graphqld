package graph

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

type Unknown struct {
	name string

	ReferencedType ast.Type

	Referencer interface{}
}

func (u *Unknown) Name() string {
	return u.name
}

func (*Unknown) Description() string {
	defer func() {
		panic("(Unknown)Description should never be called")
	}()

	return ""
}

func (*Unknown) String() string {
	defer func() {
		panic("(Unknown)String should never be called")
	}()

	return ""
}

func (*Unknown) Error() error {
	defer func() {
		panic("(Unknown)Error should never be called")
	}()

	return nil
}

func NewType(t ast.Type, referencer interface{}) (graphql.Type, *Unknown) {
	var (
		u *Unknown

		handleList    func(x *ast.List) graphql.Type
		handleNonNull func(x *ast.NonNull) graphql.Type

		standardBasicScalarFromNamedType = func(named *ast.Named) *graphql.Scalar {
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

		handleNamed = func(x *ast.Named) graphql.Type {
			if scalar := standardBasicScalarFromNamedType(x); scalar != nil {
				return scalar
			}

			u = &Unknown{
				name:           x.Name.Value,
				ReferencedType: x,
				Referencer:     referencer,
			}

			return u
		}

		_typeSwitch = func(x ast.Type) graphql.Type {
			switch x := x.(type) {
			case *ast.List:
				return handleList(x)
			case *ast.NonNull:
				return handleNonNull(x)
			case *ast.Named:
				return handleNamed(x)
			default:
				return nil
			}
		}
	)

	{
		handleList = func(x *ast.List) graphql.Type {
			return graphql.NewList(
				_typeSwitch(x.Type),
			)
		}

		handleNonNull = func(x *ast.NonNull) graphql.Type {
			return graphql.NewNonNull(
				_typeSwitch(x.Type),
			)
		}

	}

	switch x := t.(type) {
	case *ast.NonNull:
		return handleNonNull(x), u
	case *ast.List:
		return handleList(x), u
	case *ast.Named:
		return handleNamed(x), u
	}

	return nil, nil
}

func (u *Unknown) ModifyType(t graphql.Type) graphql.Type {
	var (
		handleList    func(x *ast.List, t graphql.Type) graphql.Type
		handleNonNull func(x *ast.NonNull, t graphql.Type) graphql.Type

		handleNamed = func(x *ast.Named, t graphql.Type) graphql.Type {
			return t
		}

		_typeSwitch = func(x ast.Type, t graphql.Type) graphql.Type {
			switch x := x.(type) {
			case *ast.List:
				return handleList(x, t)
			case *ast.NonNull:
				return handleNonNull(x, t)
			case *ast.Named:
				return handleNamed(x, t)
			default:
				return nil
			}
		}
	)

	{
		handleList = func(x *ast.List, t graphql.Type) graphql.Type {
			return graphql.NewList(
				_typeSwitch(x.Type, t),
			)
		}

		handleNonNull = func(x *ast.NonNull, t graphql.Type) graphql.Type {
			return graphql.NewNonNull(
				_typeSwitch(x.Type, t),
			)
		}
	}

	switch x := u.ReferencedType.(type) {
	case *ast.NonNull:
		return handleNonNull(x, t)
	case *ast.List:
		return handleList(x, t)
	case *ast.Named:
		return handleNamed(x, t)
	}

	return nil
}
