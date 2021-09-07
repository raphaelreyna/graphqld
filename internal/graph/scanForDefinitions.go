package graph

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/raphaelreyna/graphqld/internal/scan"
)

func (g *Graph) scanForDefinitions() error {
	if g.definitions == nil {
		g.definitions = make(map[string]interface{})
		g.resolverPaths = make(map[string]map[string]string)
	}

	return filepath.WalkDir(g.DocumentRoot, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		var file scan.File
		{
			info, err := d.Info()
			if err != nil {
				return err
			}

			if file = scan.NewFile(path, info); file == nil {
				return nil
			}

			if err := file.Scan(); err != nil {
				if !errors.Is(err, scan.ErrNotAResolver) {
					return nil
				}

				return err
			}
		}

		switch file := file.(type) {
		case *scan.ExecFile:
			for _, field := range file.Fields {
				var key = fmt.Sprintf("field::%s:%s", file.ObjectName, file.Name)

				g.definitions[key] = field

				paths, ok := g.resolverPaths[file.ObjectName]
				if !ok {
					paths = make(map[string]string)
					g.resolverPaths[file.ObjectName] = paths
				}

				paths[file.Name] = file.Path()
			}
		case *scan.GraphqlFile:
			for _, obj := range file.Objects {
				g.definitions["object::"+obj.Name.Value] = obj
			}

			for _, input := range file.Inputs {
				g.definitions["input::"+input.Name.Value] = input
			}

			for _, enum := range file.Enums {
				g.definitions["enum::"+enum.Name.Value] = enum
			}

			for _, iface := range file.Interfaces {
				g.definitions["iface::"+iface.Name.Value] = iface
			}
		}

		return nil
	})
}
