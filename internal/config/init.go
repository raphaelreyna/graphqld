package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("graphqld")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.config/graphqld/")
	viper.AddConfigPath("/etc/")

	viper.SetEnvKeyReplacer(strings.NewReplacer(
		"CONTEXTEXECPATH", "CTX_EXEC_PATH",
		"CONTEXTTMPDIR", "CTX_TMP_DIR",
		"RESOLVERDIR", "RESOLVER_DIR",
		"LOGJSON", "LOG_JSON",
		"LOGCOLOR", "LOG_COLOR",
		"MAXBODYSIZE", "MAX_BODY_SIZE",
	))

	viper.SetEnvPrefix("GRAPHQLD")
	viper.AutomaticEnv()

	defaults()

	Config.readInConf()

	if x, ok := viper.Get("log").(map[string]interface{}); ok {
		Config.Log = logFromMap(x)
	}

	{
		var logc = Config.Log

		if !logc.JSON {
			log.Logger = log.Output(zerolog.ConsoleWriter{
				Out:     os.Stdout,
				NoColor: !logc.Color,
			})
		}

		log.Logger = log.Logger.Level(logc.Level)
	}

	if !viper.GetBool("logJSON") {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:     os.Stdout,
			NoColor: !viper.GetBool("logColor"),
		})
	}

	log.Info().
		Str("file", viper.ConfigFileUsed()).
		Msg("read configuration")

	var (
		confGraphs = make(map[string]GraphConf)
		dirGraphs  = make([]GraphConf, 0)
	)
	{
		if iface := viper.Get("graphs"); iface != nil {
			for _, v := range iface.([]interface{}) {
				var (
					m  = v.(map[interface{}]interface{})
					gc = graphConfFromMap(m)
				)

				confGraphs[gc.ServerName] = gc
			}
		}

		dirs, err := os.ReadDir(Config.RootDir)
		if err != nil {
			log.Fatal().Err(err).
				Msg("unable to read root directory")
		}

		for _, item := range dirs {
			if !item.IsDir() {
				continue
			}

			var (
				name = item.Name()
				path = filepath.Join(Config.RootDir, name)
			)

			isOk, err := isGraphDir(path)
			if err != nil {
				log.Fatal().Err(err).
					Msg("unable to check if directory has graph")
			}

			if !isOk {
				continue
			}

			gc := GraphConf{
				HotReload:       viper.GetBool("hot"),
				Graphiql:        viper.GetBool("graphiql"),
				DocumentRoot:    path,
				ResolverDir:     viper.GetString("resolverDir"),
				MaxBodyReadSize: viper.GetInt64("maxBodySize"),
			}

			if cc := Config.CORS; cc != nil {
				gc.CORS = cc
			}

			if ba := Config.BasicAuth; ba != nil {
				gc.BasicAuth = ba
			}

			if c := Config.Context; c != nil {
				gc.Context = c
			}

			if err := checkDomain(name); err != nil {
				log.Fatal().Err(err).
					Msg("invalid domain name")
			}
			gc.ServerName = name

			dirGraphs = append(dirGraphs, gc)
		}

		if len(dirGraphs) == 0 {
			gc := GraphConf{
				HotReload:    viper.GetBool("hot"),
				Graphiql:     viper.GetBool("graphiql"),
				DocumentRoot: Config.RootDir,
				ResolverDir:  viper.GetString("resolverDir"),
			}

			if cc := Config.CORS; cc != nil {
				gc.CORS = cc
			}

			if ba := Config.BasicAuth; ba != nil {
				gc.BasicAuth = ba
			}

			if Config.Hostname != "" {
				if err := checkDomain(Config.Hostname); err != nil {
					log.Fatal().Err(err).
						Msg("invalid domain name")
				}
			}

			dirGraphs = append(dirGraphs, gc)
		}
	}

	// compare the graph configs obtained from that graphs viper vs
	// the configs generated by scanning the filesystem + default viper config
	// (individual graph config overrides)
	Config.Graphs = make([]GraphConf, 0)
	for _, graph := range dirGraphs {
		confGraph, ok := confGraphs[graph.ServerName]
		if !ok {
			Config.Graphs = append(Config.Graphs, graph)
			continue
		}

		if x := confGraph.ResolverDir; x != "" {
			graph.ResolverDir = x
		}

		if x := confGraph.HotReload; confGraph.hotReloadSet {
			graph.HotReload = x
		}

		if x := confGraph.Graphiql; confGraph.graphiqlSet {
			graph.Graphiql = x
		}

		if x := confGraph.MaxBodyReadSize; x > 0 {
			graph.MaxBodyReadSize = x
		}

		if x := confGraph.CORS; x != nil {
			graph.CORS = x
		}

		if x := confGraph.BasicAuth; x != nil {
			graph.BasicAuth = x
		}

		if x := confGraph.Context; x != nil {
			graph.Context = x
		}

		Config.Graphs = append(Config.Graphs, graph)
	}
}

func isGraphDir(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, err
	}

	for _, entry := range entries {
		var name = entry.Name()

		if name == "Query" || name == "Mutation" {
			return true, nil
		}
	}

	return false, nil
}

func defaults() {
	viper.SetDefault("hostname", "")
	viper.SetDefault("address", "")
	viper.SetDefault("root", "/var/graphqld")
	viper.SetDefault("hot", false)
	viper.SetDefault("graphiql", false)
	viper.SetDefault("contextExecPath", "")
	viper.SetDefault("contextTmpDir", "")
	viper.SetDefault("resolverDir", "/")
	viper.SetDefault("port", "80")
	viper.SetDefault("logJSON", false)
	viper.SetDefault("logColor", true)
	viper.SetDefault("resolverDir", "/")
	viper.SetDefault("maxBodySize", 1<<20) // 1 MB
}
