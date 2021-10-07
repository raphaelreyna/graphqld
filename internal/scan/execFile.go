package scan

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/raphaelreyna/graphqld/internal/config"
)

var ErrNotAResolver = errors.New("not a resolver")

type ExecFile struct {
	Dir, Name, Ext string

	ObjectName string
	Fields     []*ast.FieldDefinition
}

func (ef *ExecFile) Path() string {
	return filepath.Join(ef.Dir, ef.Name+ef.Ext)
}

func (ef *ExecFile) Scan() error {
	var (
		path         = ef.Path()
		fieldStrings []string
	)
	// populate fieldStrings
	{
		cmd := exec.Command(path, "--graphqld-fields")

		if user := config.Config.User; user != nil {
			cmd.SysProcAttr = &syscall.SysProcAttr{
				Credential: &syscall.Credential{
					Uid: user.Uid,
					Gid: user.Gid,
				},
			}

			cmd.Env = append(cmd.Env,
				"USER="+user.Name,
				"USERNAME="+user.Name,
				"LOGNAME="+user.Name,
			)

			if user.HomeDir != "" {
				cmd.Env = append(cmd.Env, "HOME="+user.HomeDir)
			}
		}

		schemaBytes, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf(
				"error executing %s --graphqld-fields: %w",
				path, ErrNotAResolver,
			)
		}

		if err := json.Unmarshal(schemaBytes, &fieldStrings); err != nil {
			return fmt.Errorf(
				"error parsing json output of %s --graphqld-fields: %w",
				path, err,
			)
		}
	}

	// parse field strings
	{
		parsedOutput, err := parser.Parse(parser.ParseParams{
			Source: fmt.Sprintf(
				"type Query {\n\t%s\n}",
				strings.Join(fieldStrings, "\n\t"),
			),
		})
		if err != nil {
			return fmt.Errorf(
				"error parsing fields returned by %s: %w",
				path, err,
			)
		}

		if len(parsedOutput.Definitions) != 1 {
			return fmt.Errorf(
				"error parsing fields returned by %s: expected 1 definition",
				path,
			)
		}

		objDef, ok := parsedOutput.Definitions[0].(*ast.ObjectDefinition)
		if !ok {
			return fmt.Errorf(
				"error parsing fields returned by %s: no object definition found",
				path,
			)
		}

		ef.Fields = objDef.Fields
	}

	return nil
}
