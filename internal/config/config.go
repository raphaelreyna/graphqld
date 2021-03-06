package config

import (
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var Config Conf

type Conf struct {
	Hostname        string
	Addr            string
	RootDir         string
	HotReload       bool
	ResolverDir     string
	Graphiql        bool
	User            *User
	UID, GID        uint32
	MaxBodyReadSize int64

	CORS      *CORSConfig
	BasicAuth *BasicAuth
	TLS       *TLS
	Context   *Context
	Log       *Log

	Graphs []GraphConf
}

func (c Conf) readInConf() {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatal().Err(err).
				Str("configuration-file", viper.ConfigFileUsed()).
				Msg("error reading configuration file")
		}

		log.Info().Msg("no configuration file found")
	}

	Config.Hostname = viper.GetString("hostname")
	Config.Addr = viper.GetString("address")
	Config.RootDir = viper.GetString("root")
	Config.HotReload = viper.GetBool("hot")
	Config.Graphiql = viper.GetBool("graphiql")
	Config.ResolverDir = viper.GetString("resolverDir")
	Config.MaxBodyReadSize = viper.GetInt64("maxBodySize")
	Config.CORS = CORSConfigFromViper()

	if !filepath.IsAbs(Config.RootDir) {
		path, err := filepath.Abs(Config.RootDir)
		if err != nil {
			log.Fatal().Err(err).
				Str("path", Config.RootDir).
				Msg("unable to compute absolute root path")
		}
		Config.RootDir = path
	}
	if !filepath.IsAbs(Config.ResolverDir) {
		path, err := filepath.Abs(Config.ResolverDir)
		if err != nil {
			log.Fatal().Err(err).
				Str("path", Config.ResolverDir).
				Msg("unable to compute resolver dir root path")
		}
		Config.ResolverDir = path
	}

	if Config.Addr == "" {
		Config.Addr = ":" + viper.GetString("port")
	}

	Config.User = userFromName(viper.GetString("user"))

	if x, ok := viper.Get("basicAuth").(map[string]interface{}); ok {
		m := make(map[interface{}]interface{})
		for k, v := range x {
			m[k] = v
		}

		Config.BasicAuth = basicAuthFromMap(m)
	}

	if x, ok := viper.Get("tls").(map[string]interface{}); ok {
		Config.TLS = tlsFromMap(x)
	}

	if x, ok := viper.Get("context").(map[string]interface{}); ok {
		m := make(map[interface{}]interface{})
		for k, v := range x {
			m[k] = v
		}

		Config.Context = contextFromMap(m)
	}

	// Grab the contextExecPath from the environment
	{
		var (
			ctxPath = viper.GetString("contextExecPath")
			tmpDir  = viper.GetString("contextTmpDir")
		)

		if ctxPath != "" {
			if Config.Context == nil {
				Config.Context = &Context{}
			}

			Config.Context.ExecPath = ctxPath
		}

		if tmpDir != "" {
			if Config.Context == nil {
				Config.Context = &Context{}
			}

			Config.Context.TmpDir = tmpDir
		}
	}
}
