package resolver

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/graphql-go/graphql"
)

type outputParser func(data []byte) (interface{}, error)

func newOutputParser(outputType graphql.Output) (outputParser, error) {
	switch x := outputType.(type) {
	case *graphql.NonNull:
		switch x := x.OfType.(type) {
		case *graphql.Scalar:
			switch x {
			case graphql.ID:
				fallthrough
			case graphql.String:
				return func(data []byte) (interface{}, error) {
					return string(data), nil
				}, nil
			case graphql.Int:
				return func(data []byte) (interface{}, error) {
					x, err := strconv.Atoi(string(data))
					if err != nil {
						return nil, err
					}
					return x, nil
				}, nil
			case graphql.Float:
				return func(data []byte) (interface{}, error) {
					x, err := strconv.ParseFloat(string(data), 64)
					if err != nil {
						return nil, err
					}
					return x, nil
				}, nil
			case graphql.Boolean:
				return func(data []byte) (interface{}, error) {
					switch string(data) {
					case "True":
						fallthrough
					case "true":
						return true, nil
					case "False":
						fallthrough
					case "false":
						return false, nil
					default:
						return nil, fmt.Errorf(
							"non bool output: %s",
							string(data),
						)
					}
				}, nil
			case graphql.DateTime:
				return func(data []byte) (interface{}, error) {
					t, err := time.Parse(time.RFC3339, string(data))
					if err != nil {
						return nil, err
					}
					return t, nil
				}, nil
			default:
				return nil, fmt.Errorf("unsupported return type: %T", x)
			}
		case *graphql.Object:
			return func(data []byte) (interface{}, error) {
				var jsonOutput interface{}
				if err := json.Unmarshal(data, &jsonOutput); err != nil {
					return nil, err
				}
				return jsonOutput, nil
			}, nil
		default:
			return nil, fmt.Errorf("invalid output type: %T %+v", outputType, outputType)
		}
	case *graphql.List:
		panic("TODO")
	case *graphql.Scalar:
		switch x {
		case graphql.ID:
			fallthrough
		case graphql.String:
			return func(data []byte) (interface{}, error) {
				if len(data) == 0 {
					return nil, nil
				}

				return string(data), nil
			}, nil
		case graphql.Int:
			return func(data []byte) (interface{}, error) {
				if len(data) == 0 {
					return nil, nil
				}

				x, err := strconv.Atoi(string(data))
				if err != nil {
					return nil, err
				}
				return x, nil
			}, nil
		case graphql.Float:
			return func(data []byte) (interface{}, error) {
				if len(data) == 0 {
					return nil, nil
				}

				x, err := strconv.ParseFloat(string(data), 64)
				if err != nil {
					return nil, err
				}
				return x, nil
			}, nil
		case graphql.Boolean:
			return func(data []byte) (interface{}, error) {
				if len(data) == 0 {
					return nil, nil
				}

				switch string(data) {
				case "True":
					fallthrough
				case "true":
					return true, nil
				case "False":
					fallthrough
				case "false":
					return false, nil
				default:
					return nil, fmt.Errorf("non bool output: %s",
						string(data),
					)
				}
			}, nil
		case graphql.DateTime:
			return func(data []byte) (interface{}, error) {
				if len(data) == 0 {
					return nil, nil
				}

				t, err := time.Parse(time.RFC3339, string(data))
				if err != nil {
					return nil, err
				}
				return t, nil
			}, nil
		default:
			return nil, fmt.Errorf("unsupported return type: %T", x)
		}
	case *graphql.Object:
		return func(data []byte) (interface{}, error) {
			if len(data) == 0 {
				return nil, nil
			}

			var jsonOutput interface{}
			if err := json.Unmarshal(data, &jsonOutput); err != nil {
				return nil, err
			}
			return jsonOutput, nil
		}, nil
	default:
		return nil, fmt.Errorf(
			"unsupported return type: %T",
			x,
		)
	}
}

func argStringFromValue(argConf *graphql.ArgumentConfig, name string, v interface{}) (string, error) {
	switch x := argConf.Type.(type) {
	case *graphql.NonNull:
		switch x := x.OfType.(type) {
		case *graphql.Scalar:
			switch x {
			case graphql.ID:
				fallthrough
			case graphql.String:
				return v.(string), nil
			case graphql.Int:
				fallthrough
			case graphql.Float:
				fallthrough
			case graphql.Boolean:
				fallthrough
			case graphql.DateTime:
				return fmt.Sprintf("%v", v), nil
			}
		case *graphql.Object:
			data, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return string(data), nil
		}
	case *graphql.List:
		panic("TODO")
	case *graphql.Scalar:
		switch argConf.Type {
		case graphql.ID:
			fallthrough
		case graphql.String:
			return v.(string), nil
		case graphql.Int:
			fallthrough
		case graphql.Float:
			fallthrough
		case graphql.Boolean:
			fallthrough
		case graphql.DateTime:
			return fmt.Sprintf("%v", v), nil
		default:
			data, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return string(data), nil
		}
	}

	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
