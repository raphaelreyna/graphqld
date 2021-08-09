package main

import "github.com/graphql-go/graphql"

type interfacer interface {
	Interfaces() []*graphql.Interface
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

func NewTypeMap(typeFielder ...namerFielder) (typeObjectMap, error) {
	var (
		argc = len(typeFielder)
		tm   = make(typeObjectMap, argc)
	)

	// Allocate space for graphql type objects
	for _, tf := range typeFielder {
		tm[tf.TypeName()] = new(graphql.Object)
	}

	// Handle wiring
	var rectifiedFields = make(map[string]graphql.Fields, argc)
	for _, t := range typeFielder {
		// New field objects are created every time (TypeFielder).TypeFields is called.
		// Since we're about to modify the field objects, we need to capture them.
		rectifiedFields[t.TypeName()] = t.TypeFields()

		for _, field := range rectifiedFields[t.TypeName()] {
			if t, ok := field.Type.(_type); ok {
				field.Type = tm[t.Name()]
			} else if t, ok := field.Type.(_listType); ok {
				field.Type = graphql.NewList(tm[t.Name()])
			}
		}
	}

	// Instantiate all of type objects in the type map
	for _, t := range typeFielder {
		var (
			// since we're using thunks we need to capture t so
			// the underlying value doesn't shift around in the for loop
			t = t

			name = t.TypeName()
		)

		objConf := graphql.ObjectConfig{
			Name: name,
			Fields: graphql.FieldsThunk(func() graphql.Fields {
				return rectifiedFields[name]
			}),
		}

		if t, ok := t.(interfacer); ok {
			objConf.Interfaces = graphql.InterfacesThunk(t.Interfaces)
		}

		// We replace the underlying graphql.Object value without changing the pointer
		// since the wiring done in the for loop above depends on where things are in the heap.
		*tm[name] = *graphql.NewObject(objConf)
	}

	return tm, nil
}
