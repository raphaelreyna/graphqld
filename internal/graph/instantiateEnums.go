package graph

import (
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

func (g *Graph) instantiateEnums(defs definitions) (enums, error) {
	var enums = make(enums)

	for k, v := range defs {
		var (
			parts   = strings.Split(k, "::")
			defType = parts[0]
			name    = parts[1]
		)

		if defType != "enum" {
			continue
		}

		var (
			enum   = v.(*ast.EnumDefinition)
			values = make(graphql.EnumValueConfigMap)
		)

		for idx, valueDef := range enum.Values {
			var (
				name        = valueDef.Name.Value
				description string
			)

			if d := valueDef.Description; d != nil {
				description = d.Value
			}

			values[name] = &graphql.EnumValueConfig{
				Value:       idx,
				Description: description,
			}
		}

		var description string
		if d := enum.Description; d != nil {
			description = d.Value
		}
		enums[name] = graphql.NewEnum(graphql.EnumConfig{
			Name:        name,
			Values:      values,
			Description: description,
		})

		delete(defs, k)
	}

	return enums, nil
}
