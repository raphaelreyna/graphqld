package graph

import (
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

func (g *Graph) instantiateInputs(defs definitions, enums enums) (inputs, error) {
	var (
		unknowns = make([]*Unknown, 0)
		inputs   = make(map[string]*graphql.InputObject)
	)

	for k, v := range defs {
		var (
			parts   = strings.Split(k, "::")
			defType = parts[0]
			name    = parts[1]
		)

		if defType != "input" {
			continue
		}

		var (
			inputDef = v.(*ast.InputObjectDefinition)
			fields   = make(graphql.InputObjectConfigFieldMap)
		)

		for _, field := range inputDef.Fields {
			var (
				fieldConf graphql.InputObjectFieldConfig

				u *Unknown
			)

			if d := field.Description; d != nil {
				fieldConf.Description = d.Value
			}
			if d := field.DefaultValue; d != nil {
				fieldConf.DefaultValue = d.GetValue()
			}

			fieldConf.Type, u = NewType(field.Type, &fieldConf)
			if u != nil {
				unknowns = append(unknowns, u)
			}

			fields[field.Name.Value] = &fieldConf
		}

		var description string
		if d := inputDef.Description; d != nil {
			description = d.Value
		}

		inputs[name] = graphql.NewInputObject(graphql.InputObjectConfig{
			Name:        name,
			Fields:      fields,
			Description: description,
		})

		delete(defs, k)
	}

	for _, u := range unknowns {
		var (
			referencer     = u.Referencer.(*graphql.InputObjectFieldConfig)
			referencedName = u.Name()
		)

		if referenced, ok := enums[referencedName]; ok {
			referencer.Type = u.ModifyType(referenced)
			continue
		}

		if referenced, ok := inputs[referencedName]; ok {
			referencer.Type = u.ModifyType(referenced)
		}
	}

	return inputs, nil
}
