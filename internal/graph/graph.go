package graph

import (
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/raphaelreyna/graphqld/internal/objdef"
)

type Graph struct {
	Dir string

	tm                  map[string]*graphql.Object
	uninstantiatedTypes map[string]interface{}
	typeReferences      []*typeReference

	im                   map[string]graphql.Input
	uninstantiatedInputs map[string]interface{}
	inputReferences      []*inputReference

	objDefs    map[string]*objdef.ObjectDefinition
	inputConfs map[string]*graphql.InputObjectConfig

	Query objdef.ObjectDefinition
}

type typeWrapper uint

const (
	twNone = iota
	twNonNull
	twList
)

type typeReference struct {
	referencingDir       string
	referencingType      string
	referencingFieldName string
	referer              interface{}
	referencedType       string
	typeWrapper          typeWrapper
}

type inputWrapper uint

const (
	iwNone = iota
	iwNonNull
	iwList
)

type inputReference struct {
	referencingDir       string
	referencingType      string
	referencingFieldName string
	referencingArgName   string
	referer              interface{}
	referencedInput      string
	inputWrapper         inputWrapper
}

func (ir *inputReference) key(name string) string {
	return fmt.Sprintf(
		"@%s::%s::%s::%s",
		ir.referencingType, ir.referencingFieldName,
		ir.referencingArgName, name,
	)
}
