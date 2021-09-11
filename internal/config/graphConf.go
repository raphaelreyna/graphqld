package config

import (
	"path/filepath"

	"github.com/rs/zerolog/log"
)

type GraphConf struct {
	ServerName      string
	DocumentRoot    string
	HotReload       bool
	hotReloadSet    bool
	ResolverDir     string
	Graphiql        bool
	graphiqlSet     bool
	MaxBodyReadSize int64

	CORS      *CORSConfig
	BasicAuth *BasicAuth
	Context   *Context
}

func graphConfFromMap(m map[interface{}]interface{}) GraphConf {
	var gc GraphConf

	if x, ok := m["serverName"]; ok {
		gc.ServerName = x.(string)
	}

	if x, ok := m["hot"]; ok {
		gc.HotReload = x.(bool)
		gc.hotReloadSet = true
	}

	if x, ok := m["resolverDir"]; ok {
		gc.ResolverDir = x.(string)
	}

	if x, ok := m["graphiql"]; ok {
		gc.Graphiql = x.(bool)
		gc.graphiqlSet = true
	}

	if x, ok := m["maxBodySize"]; ok {
		gc.MaxBodyReadSize = x.(int64)
	}

	if !filepath.IsAbs(gc.ResolverDir) && gc.ResolverDir != "" {
		path, err := filepath.Abs(gc.ResolverDir)
		if err != nil {
			log.Fatal().Err(err).
				Str("path", gc.ResolverDir).
				Msg("unable to compute resolver dir root path")
		}
		gc.ResolverDir = path
	}

	if x, ok := m["cors"].(map[interface{}]interface{}); ok {
		gc.CORS = CORSConfigFromMap(x)
	}

	if x, ok := m["basicAuth"].(map[interface{}]interface{}); ok {
		gc.BasicAuth = basicAuthFromMap(x)
	}

	if x, ok := m["context"].(map[interface{}]interface{}); ok {
		gc.Context = contextFromMap(x)
	}

	return gc
}
