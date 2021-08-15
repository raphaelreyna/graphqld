package scan

import (
	"github.com/graphql-go/graphql/language/ast"
)

// FieldOutput is an abstraction of the partial definitions
// that graphqld obtains from scanning its root directory.
//
// For executables, this is obtained by running the executable with the `--cggi-fields`.
type FieldOutput struct {
	Raw       string
	Name      string
	Type      ast.Type
	Arguments []*ast.InputValueDefinition
}

func newFieldOutput(fd *ast.FieldDefinition) FieldOutput {
	return FieldOutput{
		Name:      fd.Name.Value,
		Type:      fd.Type,
		Arguments: fd.Arguments,
	}
}

type FileFields struct {
	Path       string
	ParentName string
	Fields     []*FieldOutput
}
