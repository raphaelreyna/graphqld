package scan

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/textproto"
	"os/exec"
	"strconv"

	"github.com/graphql-go/graphql"
)

type ObjectDefinition struct {
	ResolverPaths    map[string]string
	DefinitionString string
	ObjectConf       graphql.ObjectConfig
}

// SetResolvers shall only be called after type objects have been reified.
// SetResolvers creates resolvers reshape graphql arguments into a cli args list, call the appropriate resolver script,
// and and reshape the output into whats expected at the graphql layer.
func (od *ObjectDefinition) SetResolvers() error {
	var (
		fields = od.ObjectConf.Fields.(graphql.FieldsThunk)()
	)

	for name, field := range fields {
		var (
			takesArgs = 0 < len(field.Args)
			field     = field
			name      = name
		)

		if _, exists := od.ResolverPaths[name]; !exists {
			continue
		}

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
			if p.Source != nil {
				source, err := json.Marshal(p.Source)
				if err != nil {
					return nil, err
				}

				cmd.Stdin = bytes.NewReader(source)
			}
			output, err := cmd.Output()
			if err != nil {
				exitErr, ok := err.(*exec.ExitError)
				if !ok {
					return nil, err
				}
				return nil, errors.New(string(exitErr.Stderr))
			}

			parts := bytes.SplitN(output, []byte("\n\n"), 2)
			if len(parts) == 2 {
				output = parts[1]
				tpReader := textproto.NewReader(
					bufio.NewReader(
						bytes.NewReader(parts[0]),
					),
				)

				header, err := tpReader.ReadMIMEHeader()
				if err != nil && !errors.Is(err, io.EOF) {
					return nil, err
				}

				fmt.Printf("found header: %+v\n", header)
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
				case graphql.Boolean:
					switch x := string(output); x {
					case "True":
						fallthrough
					case "true":
						return true, nil
					case "False":
						fallthrough
					case "false":
						return false, nil
					default:
						panic(fmt.Sprintf("non bool output: %s", x))
					}
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
