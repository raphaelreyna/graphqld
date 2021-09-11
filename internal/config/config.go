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
	ContextExecPath string
	ContextFilesDir string
	MaxBodyReadSize int64

	CORS      *CORSConfig
	BasicAuth *BasicAuth

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
	} else {
		log.Info().
			Str("file", viper.ConfigFileUsed()).
			Msg("read configuration")
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

	if x, ok := viper.Get("basicAuth").(map[string]interface{}); ok {
		m := make(map[interface{}]interface{})
		for k, v := range x {
			m[k] = v
		}

		Config.BasicAuth = basicAuthFromMap(m)
	}
}
