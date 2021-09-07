package resolver

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/textproto"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/graphql-go/graphql"
	"github.com/raphaelreyna/graphqld/internal/transport/http"
)

func NewFieldResolveFn(path, wd string, field *graphql.FieldDefinition) (*graphql.FieldResolveFn, error) {
	var (
		takesArgs  = 0 < len(field.Args)
		scriptName = filepath.Base(path)
	)

	parseOutput, err := newOutputParser(field.Type)
	if err != nil {
		return nil, fmt.Errorf(
			"NewFieldResolveFn:: error creating output parser for field %s: %w",
			field.Name, err,
		)
	}

	var f = func(p graphql.ResolveParams) (interface{}, error) {
		var (
			args      = make([]string, 0)
			namedArgs = make(map[string]*graphql.Argument)
		)

		for _, arg := range field.Args {
			namedArgs[arg.Name()] = arg
		}

		if takesArgs {
			for name, arg := range p.Args {
				argType := namedArgs[name].Type
				argStr, err := argStringFromValue(argType, name, arg)
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

		env := http.GetEnv(p.Context)
		env = append(env,
			"SCRIPT_NAME="+scriptName,
			"SCRIPT_FILENAME="+path,
		)
		cmd.Env = env

		if ctxFile := http.GetCtxFile(p.Context); ctxFile != nil {
			cmd.ExtraFiles = []*os.File{ctxFile}
		}

		if wd != "" {
			cmd.Dir = wd
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

			h := http.GetWHeader(p.Context)
			for k := range header {
				h.Add(k, header.Get(k))
			}
		}

		return parseOutput(output)
	}
	var ff = graphql.FieldResolveFn(f)
	return &ff, nil
}
