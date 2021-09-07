package graph

import (
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

func (g *Graph) instantiateObjects(defs definitions, enums enums, inputs inputs) (objects, error) {
	var (
		unknowns = make([]*Unknown, 0)

		fieldsMap = map[string]graphql.Fields{
			"Query":    make(graphql.Fields),
			"Mutation": make(graphql.Fields),
		}

		objects = objects{
			"Query": graphql.NewObject(graphql.ObjectConfig{
				Name: "Query",
				Fields: graphql.FieldsThunk(func() graphql.Fields {
					return fieldsMap["Query"]
				}),
			}),
			"Mutation": graphql.NewObject(graphql.ObjectConfig{
				Name: "Mutation",
				Fields: graphql.FieldsThunk(func() graphql.Fields {
					return fieldsMap["Mutation"]
				}),
			}),
		}
	)

	for k, v := range defs {
		var (
			parts   = strings.Split(k, "::")
			defType = parts[0]
			name    = parts[1]
		)

		if defType != "object" {
			continue
		}

		var (
			objDef = v.(*ast.ObjectDefinition)
			fields = make(graphql.Fields)
		)

		for _, field := range objDef.Fields {
			var (
				fieldConf = graphql.Field{
					Name: name,
				}

				u *Unknown
			)

			if d := field.Description; d != nil {
				fieldConf.Description = d.Value
			}

			fieldConf.Type, u = NewType(field.Type, &fieldConf)
			if u != nil {
				unknowns = append(unknowns, u)
			}

			var argConfs = make(graphql.FieldConfigArgument)
			for _, arg := range field.Arguments {
				var (
					argConf graphql.ArgumentConfig
					u       *Unknown
				)

				if d := arg.Description; d != nil {
					argConf.Description = d.Value
				}

				if d := arg.DefaultValue; d != nil {
					argConf.DefaultValue = d.GetValue()
				}

				argConf.Type, u = NewType(arg.Type, &argConf)
				if u != nil {
					unknowns = append(unknowns, u)
				}

				argConfs[arg.Name.Value] = &argConf
			}

			fieldConf.Args = argConfs

			fields[field.Name.Value] = &fieldConf
		}

		var description string
		if d := objDef.Description; d != nil {
			description = d.Value
		}

		objects[name] = graphql.NewObject(graphql.ObjectConfig{
			Name: name,
			Fields: graphql.FieldsThunk(func() graphql.Fields {
				return fields
			}),
			Description: description,
		})

		fieldsMap[name] = fields

		delete(defs, k)
	}

	for k, v := range defs {
		var (
			parts = strings.Split(k, "::")

			defType = parts[0]
			name    = parts[1]
			objName string
		)

		if defType != "field" {
			continue
		}

		parts = strings.Split(name, ":")
		objName = parts[0]
		name = parts[1]

		var obj, ok = objects[objName]
		if !ok {
			delete(defs, k)
			continue
		}

		var (
			u         *Unknown
			fieldDef  = v.(*ast.FieldDefinition)
			fieldConf = graphql.Field{
				Name: name,
			}
		)

		fieldConf.Type, u = NewType(fieldDef.Type, &fieldConf)
		if u != nil {
			unknowns = append(unknowns, u)
		}

		var argConfs = make(graphql.FieldConfigArgument)
		for _, arg := range fieldDef.Arguments {
			var (
				argConf graphql.ArgumentConfig
				u       *Unknown
			)

			if d := arg.Description; d != nil {
				argConf.Description = d.Value
			}

			if d := arg.DefaultValue; d != nil {
				argConf.DefaultValue = d.GetValue()
			}

			argConf.Type, u = NewType(arg.Type, &argConf)
			if u != nil {
				unknowns = append(unknowns, u)
			}

			argConfs[arg.Name.Value] = &argConf
		}

		if len(argConfs) != 0 {
			fieldConf.Args = argConfs
		}

		fields := fieldsMap[obj.Name()]
		fields[name] = &fieldConf

		delete(defs, k)
	}

	for _, u := range unknowns {
		switch referencer := u.Referencer.(type) {
		case *graphql.Field:
			var referencedName = u.Name()

			if referenced, ok := enums[referencedName]; ok {
				referencer.Type = u.ModifyType(referenced)
				continue
			}

			if referenced, ok := inputs[referencedName]; ok {
				referencer.Type = u.ModifyType(referenced)
				continue
			}

			if referenced, ok := objects[referencedName]; ok {
				referencer.Type = u.ModifyType(referenced)
			}
		case *graphql.ArgumentConfig:
			var referencedName = u.Name()

			if referenced, ok := enums[referencedName]; ok {
				referencer.Type = u.ModifyType(referenced)
				continue
			}

			if referenced, ok := inputs[referencedName]; ok {
				referencer.Type = u.ModifyType(referenced)
				continue
			}

			if referenced, ok := objects[referencedName]; ok {
				referencer.Type = u.ModifyType(referenced)
			}
		}
	}

	return objects, nil
}
