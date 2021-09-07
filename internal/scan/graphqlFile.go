package scan

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
)

type GraphqlFile struct {
	Dir, Name string

	Objects    []*ast.ObjectDefinition
	Inputs     []*ast.InputObjectDefinition
	Enums      []*ast.EnumDefinition
	Interfaces []*ast.InterfaceDefinition
}

func (gf *GraphqlFile) Path() string {
	return filepath.Join(gf.Dir, gf.Name+".graphql")
}

func (gf *GraphqlFile) Scan() error {
	var path = gf.Path()

	file, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	defer func() {
		if file != nil {
			file.Close()
		}
	}()
	if err != nil {
		return nil
	}

	filename := file.Name()
	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf(
			"error reading file %s: %w",
			filename, err,
		)
	}

	parsedOutput, err := parser.Parse(parser.ParseParams{
		Source: string(data),
	})
	if err != nil {
		return fmt.Errorf(
			"error parsing fields returned by %s: %w",
			filename, err,
		)
	}

	if gf.Objects == nil {
		gf.Objects = []*ast.ObjectDefinition{}
	}
	if gf.Inputs == nil {
		gf.Inputs = []*ast.InputObjectDefinition{}
	}
	if gf.Enums == nil {
		gf.Enums = []*ast.EnumDefinition{}
	}
	if gf.Interfaces == nil {
		gf.Interfaces = []*ast.InterfaceDefinition{}
	}

	for _, def := range parsedOutput.Definitions {
		switch x := def.(type) {
		case *ast.ObjectDefinition:
			gf.Objects = append(gf.Objects, x)
		case *ast.InputObjectDefinition:
			gf.Inputs = append(gf.Inputs, x)
		case *ast.EnumDefinition:
			gf.Enums = append(gf.Enums, x)
		case *ast.InterfaceDefinition:
			gf.Interfaces = append(gf.Interfaces, x)
		default:
			return fmt.Errorf("unsupported definition type: %T %v", def, def)
		}
	}

	return nil
}
