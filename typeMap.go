package main

import (
	"github.com/graphql-go/graphql"
)

type namerFielder interface {
	TypeName() string
	TypeFields() graphql.Fields
}

// typeObjectMap maps type names to their type object singletons.
type typeObjectMap map[string]*graphql.Object

func (tm typeObjectMap) TypeOf(tf namerFielder) *graphql.Object {
	return tm[tf.TypeName()]
}
