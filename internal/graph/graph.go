package graph

import (
	"github.com/graphql-go/graphql"
	"github.com/raphaelreyna/graphqld/internal/scan"
)

type Graph struct {
	Dir string

	tm                  typeObjectMap
	uninstantiatedTypes map[string]interface{}
	typeReferences      map[typeReference]struct{}

	Query scan.ObjectDefinition
}

type namerFielder interface {
	TypeName() string
	TypeFields() graphql.Fields
}

// typeObjectMap maps type names to their type object singletons.
type typeObjectMap map[string]*graphql.Object

func (tm typeObjectMap) TypeOf(tf namerFielder) *graphql.Object {
	return tm[tf.TypeName()]
}

type typeWrapper uint

const (
	twNone = iota
	twNonNull
	twList
)

type typeReference struct {
	referenceringType string
	referencingField  string
	referencedType    string
	typeWrapper       typeWrapper
}
