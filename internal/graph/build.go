package graph

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/graphql-go/graphql"
	"github.com/raphaelreyna/graphqld/internal/intermediary"
	"github.com/raphaelreyna/graphqld/internal/objdef"
	"github.com/raphaelreyna/graphqld/internal/scan"
	"github.com/rs/zerolog/log"
)

var ErrorNoRoots = errors.New("no root query or mutation directories found")

func (g *Graph) Build() error {
	// build object definition for root query object
	{
		if g.inputConfs == nil {
			g.inputConfs = make(map[string]*graphql.InputObjectConfig)
		}

		def, err := g.buildObjectDefinitionForTypeObject(g.Dir, "Query")
		if err != nil && !errors.Is(err, ErrTypeHasNoDir) {
			log.Info().Err(err).
				Str("dir", g.Dir).
				Msg("error building root query")
		} else {
			g.Query = def
		}

		def, err = g.buildObjectDefinitionForTypeObject(g.Dir, "Mutation")
		if err != nil && !errors.Is(err, ErrTypeHasNoDir) {
			log.Info().Err(err).
				Str("dir", g.Dir).
				Msg("error building root mutation")
		} else {
			g.Mutation = def
		}

		if g.Mutation == nil && g.Query == nil {
			return ErrorNoRoots
		}
	}

	// keep building referenced types as long as we have any
	{
		if g.objDefs == nil {
			g.objDefs = make(map[string]*objdef.ObjectDefinition)
		}

		var (
			count                   = len(g.typeReferences)
			processedTypeReferences = make([]*typeReference, 0)
		)
		for 0 < count {
			var tr = g.typeReferences[0]

			def, err := g.buildObjectDefinitionForTypeObject(tr.referencingDir, tr.referencedType)
			if err != nil {
				return err
			}

			g.objDefs[def.ObjectConf.Name] = def

			// put this type reference into the pile of processed ones
			// and remove it from the ones we still need to work on
			processedTypeReferences = append(processedTypeReferences, tr)
			g.typeReferences = g.typeReferences[1:]
			count = len(g.typeReferences)
		}

		// lets get our type references back
		g.typeReferences = processedTypeReferences
	}

	// Read all of the referenced input configs from the root dir
	{
		var (
			count                    = len(g.inputReferences)
			processedInputReferences = make([]*inputReference, 0)
		)
		for 0 < count {
			var (
				ir = g.inputReferences[0]

				dir       = ir.dir()
				filePath  = filepath.Join(dir, ir.referencedInput+".graphql")
				inputName = ir.referencedInput

				key = ir.key(inputName)

				ic = g.inputConfs[key]
			)

			// put this type reference into the pile of processed ones
			// and remove it from the ones we still need to work on
			processedInputReferences = append(processedInputReferences, ir)
			g.inputReferences = g.inputReferences[1:]

			// Grab contents from file system
			info, err := os.Stat(filePath)
			if err != nil {
				return err
			}
			file, err := scan.NewFile(dir, info)
			if err != nil {
				return err
			}
			contents, err := scan.Scan(inputName, file)
			if err != nil {
				return err
			}

			if ic == nil {
				ic = &graphql.InputObjectConfig{
					Name: inputName,
				}
				g.inputConfs[key] = ic
			}

			func(ic *graphql.InputObjectConfig) {
				fm := make(graphql.InputObjectConfigFieldMap)
				ic.Fields = fm

				for _, iField := range contents.Input.Fields {
					fm[iField.Name.Value] = &graphql.InputObjectFieldConfig{
						Type: g.gqlInputFromType(&inputReference{
							referencingDir:       ir.referencingDir,
							referencingType:      ir.referencingType,
							referencingFieldName: ir.referencingFieldName,
							referencingArgName:   ir.referencingArgName,
							referencedInput:      ir.referencedInput,
							referer:              ic,
						}, iField.Type),
					}
				}
			}(ic)

			count = len(g.inputReferences)
		}

		// lets get our input references back
		g.inputReferences = processedInputReferences
	}

	// now that we have all of the type object definitions, we need to instantiate the input objects
	{
		if g.im == nil {
			g.im = make(map[string]graphql.Input)
		}

		for name := range g.uninstantiatedInputs {
			conf, ok := g.inputConfs[name]
			if !ok {
				panic(fmt.Sprintf("could not find input config for %s", name))
			}

			g.im[name] = graphql.NewInputObject(*conf)

			delete(g.uninstantiatedInputs, name)
		}
	}

	// now that we instantiated all of our input objects, we need to make sure that
	// pointers pointing to intermediary input objects are set to point to the "real" type object
	{
		for _, ir := range g.inputReferences {
			key := ir.key(ir.referencedInput)
			switch referer := ir.referer.(type) {
			case *graphql.Field:
				switch ir.inputWrapper {
				case twList:
					referer.Type = graphql.NewList(g.im[key])
				case twNonNull:
					referer.Type = graphql.NewNonNull(g.im[key])
				case twNone:
					referer.Type = g.im[key]
				}
			case *graphql.ArgumentConfig:
				switch ir.inputWrapper {
				case twList:
					referer.Type = graphql.NewList(g.im[key])
				case twNonNull:
					referer.Type = graphql.NewNonNull(g.im[key])
				case twNone:
					referer.Type = g.im[key]
				}
			case *graphql.InputObjectConfig:
				fm := referer.Fields.(graphql.InputObjectConfigFieldMap)
				for _, f := range fm {
					if !intermediary.IsIntermediary(f.Type) {
						continue
					}
					switch ir.inputWrapper {
					case twList:
						f.Type = graphql.NewList(g.im[key])
					case twNonNull:
						f.Type = graphql.NewNonNull(g.im[key])
					case twNone:
						f.Type = g.im[key]
					}
				}
			}
		}
	}

	// now that we have all of the type object definitions, we need to instantiate them
	{
		if len(g.tm) == 0 {
			g.tm = make(map[string]*graphql.Object)
		}

		for name := range g.uninstantiatedTypes {
			def, ok := g.objDefs[name]
			if !ok {
				panic("could not find object definition")
			}

			g.tm[name] = graphql.NewObject(def.ObjectConf)

			delete(g.uninstantiatedTypes, name)
		}
	}

	// now that we instantiated all of our type objects, we need to make sure that
	// pointers pointing to intermediary type objects are set to point to the "real" type object
	{
		for _, tr := range g.typeReferences {
			switch referer := tr.referer.(type) {
			case *graphql.Field:
				switch tr.typeWrapper {
				case twList:
					referer.Type = graphql.NewList(g.tm[tr.referencedType])
				case twNonNull:
					referer.Type = graphql.NewNonNull(g.tm[tr.referencedType])
				case twNone:
					referer.Type = g.tm[tr.referencedType]
				}
			case *graphql.ArgumentConfig:
				switch tr.typeWrapper {
				case twList:
					referer.Type = graphql.NewList(g.tm[tr.referencedType])
				case twNonNull:
					referer.Type = graphql.NewNonNull(g.tm[tr.referencedType])
				case twNone:
					referer.Type = g.tm[tr.referencedType]
				}
			}
		}
	}

	// finally we create a resolver for each field that needs one
	{
		if g.Query != nil {
			if err := g.Query.SetResolvers(g.Dir, g.ResolverWD); err != nil {
				return err
			}
		}

		if g.Mutation != nil {
			if err := g.Mutation.SetResolvers(g.Dir, g.ResolverWD); err != nil {
				return err
			}
		}

		for _, objDef := range g.objDefs {
			if err := objDef.SetResolvers(g.Dir, g.ResolverWD); err != nil {
				return err
			}
		}
	}

	return nil
}
