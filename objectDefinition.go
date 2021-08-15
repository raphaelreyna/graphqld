package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/graphql-go/graphql"
	"github.com/raphaelreyna/graphqld/internal/scan"
)

func setResolvers(od *scan.ObjectDefinition) error {
	var (
		fields = od.ObjectConf.Fields.(graphql.FieldsThunk)()
	)

	for name, field := range fields {
		var (
			takesArgs = 0 < len(field.Args)
			field     = field
			name      = name
		)

		field.Resolve = graphql.FieldResolveFn(func(p graphql.ResolveParams) (interface{}, error) {
			var (
				args = []string{}
			)

			if takesArgs {
				for name, arg := range p.Args {
					argInfo := field.Args[name]
					switch argInfo.Type.(type) {
					case *graphql.NonNull:
						x := argInfo.Type.(*graphql.NonNull)
						switch x.OfType {
						case graphql.String:
							args = append(args, "--"+name, arg.(string))
						default:
							panic(fmt.Sprintf("unsupported arg type: %T %+v", field.Args[name].Type, field.Args[name].Type))
						}
					case *graphql.Scalar:
						switch field.Args[name].Type {
						case graphql.String:
							args = append(args, "--"+name, arg.(string))
						default:
							panic(fmt.Sprintf("unsupported arg type: %T %+v", field.Args[name].Type, field.Args[name].Type))
						}
					}
				}
			}

			cmd := exec.Command(od.ResolverPaths[name], args...)
			output, err := cmd.Output()
			if err != nil {
				return nil, err
			}

			switch x := field.Type.(type) {
			case *graphql.NonNull:
				switch x.OfType {
				case graphql.String:
					return string(output), nil
				case graphql.Int:
					x, err := strconv.Atoi(string(output))
					if err != nil {
						return nil, err
					}

					return x, nil
				default:
					switch x := x.OfType.(type) {
					case *graphql.Object:
						var jsonOutput interface{}
						if err := json.Unmarshal(output, &jsonOutput); err != nil {
							return nil, err
						}
						return jsonOutput, nil
					default:
						panic(fmt.Sprintf(
							"unsupported return type: %+T",
							x,
						))
					}
				}
			case *graphql.Scalar:
				switch x {
				case graphql.String:
					return string(output), nil
				case graphql.Int:
					x, err := strconv.Atoi(string(output))
					if err != nil {
						return nil, err
					}

					return x, nil
				default:
					panic("unsupported return type")
				}
			case *graphql.Object:
				var jsonOutput interface{}
				if err := json.Unmarshal(output, &jsonOutput); err != nil {
					return nil, err
				}
				return jsonOutput, nil
			default:
				panic(fmt.Sprintf(
					"unsupported return type: %T",
					x,
				))
			}
		})
	}
	return nil
}
