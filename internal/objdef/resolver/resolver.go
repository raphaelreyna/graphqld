package resolver

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/textproto"
	"os/exec"

	"github.com/graphql-go/graphql"
	"github.com/raphaelreyna/graphqld/internal/transport/http"
)

func NewFieldResolveFn(path string, field *graphql.Field) (graphql.FieldResolveFn, error) {
	var (
		takesArgs = 0 < len(field.Args)
	)

	parseOutput, err := newOutputParser(field.Type)
	if err != nil {
		return nil, fmt.Errorf(
			"NewFieldResolveFn:: error creating output parser for field %s: %w",
			field.Name, err,
		)
	}

	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			args = []string{}
		)

		if takesArgs {
			for name, arg := range p.Args {
				argInfo := field.Args[name]
				argStr, err := argStringFromValue(argInfo, name, arg)
				if err != nil {
					return nil, err
				}

				args = append(args, "--"+name, argStr)
			}
		}

		cmd := exec.Command(path, args...)
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

			h := http.GetHeader(p.Context)
			for k := range header {
				h.Add(k, header.Get(k))
			}
		}

		return parseOutput(output)
	}, nil
}
