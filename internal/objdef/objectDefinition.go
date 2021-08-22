package objdef

import (
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/raphaelreyna/graphqld/internal/objdef/resolver"
)

type ObjectDefinition struct {
	ResolverPaths    map[string]string
	DefinitionString string
	ObjectConf       graphql.ObjectConfig
}

// SetResolvers shall only be called after type objects have been reified.
// SetResolvers creates resolvers reshape graphql arguments into a cli args list, call the appropriate resolver script,
// and and reshape the output into whats expected at the graphql layer.
func (od *ObjectDefinition) SetResolvers() error {
	var fields = od.ObjectConf.Fields.(graphql.FieldsThunk)()

	for name, field := range fields {
		if _, exists := od.ResolverPaths[name]; !exists {
			continue
		}

		resolverFn, err := resolver.NewFieldResolveFn(od.ResolverPaths[name], field)
		if err != nil {
			return fmt.Errorf(
				"(*ObjectDefintion).SetResolvers:: error creating resolver for field %s: %w",
				name, err,
			)
		}

		field.Resolve = resolverFn
	}
	return nil
}
