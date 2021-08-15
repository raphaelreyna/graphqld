package scan

import "github.com/graphql-go/graphql"

type ObjectDefinition struct {
	ResolverPaths    map[string]string
	DefinitionString string
	ObjectConf       graphql.ObjectConfig
}
