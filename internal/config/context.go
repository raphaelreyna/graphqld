package config

import (
	"path/filepath"

	"github.com/rs/zerolog/log"
)

type Context struct {
	ExecPath string
	TmpDir   string
	Context  interface{}
}

func contextFromMap(m map[interface{}]interface{}) *Context {
	var c = Context{}
	c.ExecPath, _ = m["execPath"].(string)
	if c.ExecPath == "" {
		// this key is all lowercase when its coming from the root of the config file
		// (idk why)
		c.ExecPath, _ = m["execpath"].(string)
	}
	c.TmpDir, _ = m["tmpdir"].(string)

	if !filepath.IsAbs(c.ExecPath) && c.ExecPath != "" {
		path, err := filepath.Abs(c.ExecPath)
		if err != nil {
			log.Fatal().Err(err).
				Str("path", c.ExecPath).
				Msg("unable to compute absolute rooth path")
		}
		c.ExecPath = path
	}

	if !filepath.IsAbs(c.TmpDir) && c.TmpDir != "" {
		path, err := filepath.Abs(c.TmpDir)
		if err != nil {
			log.Fatal().Err(err).
				Str("path", c.TmpDir).
				Msg("unable to compute absolute rooth path")
		}
		c.TmpDir = path
	}

	if ctx := m["context"]; ctx != nil {
		c.Context = makeJSONable(ctx)
	}

	return &c
}

func makeJSONable(v interface{}) interface{} {
	switch x := v.(type) {
	case map[interface{}]interface{}:
		var m = make(map[string]interface{}, len(x))
		for k, v := range x {
			m[k.(string)] = makeJSONable(v)
		}

		return m
	default:
		return x
	}
}
